package limits

import (
	"fmt"
	"io"

	"github.com/jmoiron/sqlx"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/automation"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
	"github.com/pkg/errors"
)

// RepairReport is the outcome of a RepairDialPolicies run. Checked and Missing count dial policies; Failed
// counts failed policies and the FailedShares whose desired policies could not be derived.
type RepairReport struct {
	Apply           bool
	Checked         int
	Missing         []*MissingDialPolicy
	Created         int
	Failed          int
	SkippedAccounts []string
	SkippedShares   []string
	FailedShares    []*FailedShare
}

// FailedShare is a v1-created share whose frontend selection names no live frontend.
type FailedShare struct {
	Account          string
	ShareToken       string
	SelectedFrontend string
}

// MissingDialPolicy is a desired dial policy that was not found by name.
type MissingDialPolicy struct {
	Account    string
	ShareToken string
	Name       string
	opts       *automation.ServicePolicyOptions
}

type repairShare struct {
	account string
	shr     *store.Share
	desired []*automation.ServicePolicyOptions
}

// RepairDialPolicies finds the live shares whose dial policies are missing although their account is not
// limited, and with apply creates them. it honours the bandwidth limit journal: an account limited as a
// whole is skipped, and so are the shares of a backend mode limited by a scoped limit class; restoring
// those would lift a limit the journal still records. it never deletes anything.
func RepairDialPolicies(str *store.Store, ziti *automation.ZitiAutomation, apply bool, out io.Writer) (*RepairReport, error) {
	rpt := &RepairReport{Apply: apply}
	plan, err := repairPlan(str, rpt)
	if err != nil {
		return nil, err
	}

	for _, rs := range plan {
		for _, opts := range rs.desired {
			rpt.Checked++
			exists, err := dialPolicyExists(ziti, opts.Name)
			if err != nil {
				rpt.Failed++
				dl.Errorf("error checking dial policy '%v' for share '%v' of '%v': %v", opts.Name, rs.shr.Token, rs.account, err)
				continue
			}
			if exists {
				continue
			}
			missing := &MissingDialPolicy{Account: rs.account, ShareToken: rs.shr.Token, Name: opts.Name, opts: opts}
			rpt.Missing = append(rpt.Missing, missing)
			dl.Infof("missing dial policy '%v' for share '%v' of '%v'", opts.Name, rs.shr.Token, rs.account)
		}
	}

	if apply {
		for _, missing := range rpt.Missing {
			created, err := ensureDialPolicy(ziti, missing.opts)
			if err != nil {
				rpt.Failed++
				dl.Errorf("error creating dial policy '%v' for share '%v' of '%v': %v", missing.Name, missing.ShareToken, missing.Account, err)
				continue
			}
			if created {
				rpt.Created++
				dl.Infof("created dial policy '%v' for share '%v' of '%v'", missing.Name, missing.ShareToken, missing.Account)
			}
		}
	}

	rpt.print(out)
	if rpt.Failed > 0 {
		return rpt, errors.Errorf("'%d' failed; see the report and the log", rpt.Failed)
	}
	return rpt, nil
}

// repairPlan reads, in one transaction, every live share not held back by the journal and its desired
// dial policies. the transaction is closed before ziti is touched.
func repairPlan(str *store.Store, rpt *RepairReport) ([]*repairShare, error) {
	trx, err := str.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "error starting transaction")
	}
	defer func() { _ = trx.Rollback() }()

	shrs, err := str.FindAllShares(trx)
	if err != nil {
		return nil, err
	}

	type accountState struct {
		label           string
		limited         bool
		limitedBackends map[sdk.BackendMode]bool
	}
	envs := make(map[int]*store.Environment)
	accounts := make(map[int]*accountState)
	var plan []*repairShare
	for _, shr := range shrs {
		env, found := envs[shr.EnvironmentId]
		if !found {
			env, err = str.GetEnvironment(shr.EnvironmentId, trx)
			if err != nil {
				return nil, errors.Wrapf(err, "error finding environment for share '%v'", shr.Token)
			}
			envs[shr.EnvironmentId] = env
		}

		label := fmt.Sprintf("ephemeral environment '%v'", env.ZId)
		if env.AccountId != nil {
			as, found := accounts[*env.AccountId]
			if !found {
				acct, err := str.GetAccount(*env.AccountId, trx)
				if err != nil {
					return nil, errors.Wrapf(err, "error finding account for environment '%v'", env.ZId)
				}
				limited, limitedBackends, err := journalLimits(str, acct.Id, trx)
				if err != nil {
					return nil, errors.Wrapf(err, "error reading bandwidth limit journal for '%v'", acct.Email)
				}
				as = &accountState{label: acct.Email, limited: limited, limitedBackends: limitedBackends}
				accounts[*env.AccountId] = as
				if limited {
					rpt.SkippedAccounts = append(rpt.SkippedAccounts, acct.Email)
					dl.Infof("skipping limited account '%v'", acct.Email)
				}
			}
			if as.limited {
				continue
			}
			if as.limitedBackends[sdk.BackendMode(shr.BackendMode)] {
				rpt.SkippedShares = append(rpt.SkippedShares, shr.Token)
				dl.Infof("skipping share '%v' of '%v'; backend mode '%v' is limited", shr.Token, as.label, shr.BackendMode)
				continue
			}
			label = as.label
		}

		desired, err := desiredDialPolicies(str, shr, trx)
		var missingFe missingPublicFrontendError
		if errors.As(err, &missingFe) {
			rpt.Failed++
			rpt.FailedShares = append(rpt.FailedShares, &FailedShare{Account: label, ShareToken: shr.Token, SelectedFrontend: missingFe.publicName})
			dl.Errorf("share '%v' of '%v' selects frontend '%v', which does not exist", shr.Token, label, missingFe.publicName)
			continue
		}
		if err != nil {
			return nil, err
		}
		plan = append(plan, &repairShare{account: label, shr: shr, desired: desired})
	}
	return plan, nil
}

// journalLimits reads which of an account's shares the journal holds limited, the way relaxAction does: a
// limit entry with no limit class limits the whole account, and a limit class with a backend mode limits
// that mode. a limit class with no backend mode limits the whole account too.
func journalLimits(str *store.Store, acctId int, trx *sqlx.Tx) (bool, map[sdk.BackendMode]bool, error) {
	jes, err := str.FindAllLatestBandwidthLimitJournalForAccount(acctId, trx)
	if err != nil {
		return false, nil, err
	}
	limited := false
	limitedBackends := make(map[sdk.BackendMode]bool)
	for _, je := range jes {
		if je.LimitClassId == nil {
			if je.Action == store.LimitLimitAction {
				limited = true
			}
			continue
		}
		lc, err := str.GetLimitClass(*je.LimitClassId, trx)
		if err != nil {
			return false, nil, err
		}
		if lc.LimitAction != store.LimitLimitAction {
			continue
		}
		if lc.BackendMode == nil {
			limited = true
		} else {
			limitedBackends[*lc.BackendMode] = true
		}
	}
	return limited, limitedBackends, nil
}

func (rpt *RepairReport) print(out io.Writer) {
	if rpt.Apply {
		_, _ = fmt.Fprintf(out, "dial policy repair (creating missing policies)\n")
	} else {
		_, _ = fmt.Fprintf(out, "dial policy repair dry run (nothing created; run with --apply to create missing policies)\n")
	}
	for _, account := range rpt.SkippedAccounts {
		_, _ = fmt.Fprintf(out, "skipped limited account '%s'\n", account)
	}
	for _, token := range rpt.SkippedShares {
		_, _ = fmt.Fprintf(out, "skipped share '%s' of a limited backend mode\n", token)
	}
	if len(rpt.Missing) == 0 {
		_, _ = fmt.Fprintf(out, "no missing dial policies\n")
	} else {
		_, _ = fmt.Fprintf(out, "missing dial policies:\n")
		for _, missing := range rpt.Missing {
			_, _ = fmt.Fprintf(out, "  account '%s', share '%s', policy '%s'\n", missing.Account, missing.ShareToken, missing.Name)
		}
	}
	if len(rpt.FailedShares) > 0 {
		_, _ = fmt.Fprintf(out, "failed shares:\n")
		for _, failed := range rpt.FailedShares {
			_, _ = fmt.Fprintf(out, "  account '%s', share '%s', selects missing frontend '%s'\n", failed.Account, failed.ShareToken, failed.SelectedFrontend)
		}
	}
	_, _ = fmt.Fprintf(out, "checked: %d, missing: %d, created: %d, failed: %d\n", rpt.Checked, len(rpt.Missing), rpt.Created, rpt.Failed)
}
