package limits

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/automation"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
	"github.com/pkg/errors"
	"strings"
)

type relaxAction struct {
	str     *store.Store
	newZiti func() (*automation.ZitiAutomation, error)
}

func newRelaxAction(str *store.Store, newZiti func() (*automation.ZitiAutomation, error)) *relaxAction {
	return &relaxAction{str, newZiti}
}

// storeRelaxError distinguishes a failed SQL operation from a retryable Ziti failure.
type storeRelaxError struct{ error }

func storeFailure(err error) error { return storeRelaxError{err} }

func (a *relaxAction) HandleAccount(acct *store.Account, _, _ int64, bwc store.BandwidthClass, _ *userLimits, trx *sqlx.Tx) error {
	dl.Debugf("relaxing '%v'", acct.Email)

	envs, err := a.str.FindEnvironmentsForAccount(acct.Id, trx)
	if err != nil {
		return storeFailure(errors.Wrapf(err, "error finding environments for account '%v'", acct.Email))
	}

	jes, err := a.str.FindAllLatestBandwidthLimitJournalForAccount(acct.Id, trx)
	if err != nil {
		return storeFailure(errors.Wrapf(err, "error finding latest bandwidth limit journal entries for account '%v'", acct.Email))
	}
	limitedBackends := make(map[sdk.BackendMode]bool)
	for _, je := range jes {
		if je.LimitClassId != nil {
			lc, err := a.str.GetLimitClass(*je.LimitClassId, trx)
			if err != nil {
				return storeFailure(err)
			}
			if lc.BackendMode != nil && lc.LimitAction == store.LimitLimitAction {
				limitedBackends[*lc.BackendMode] = true
			}
		}
	}

	ziti, err := a.newZiti()
	if err != nil {
		return err
	}

	var failures []string
	for _, env := range envs {
		shrs, err := a.str.FindSharesForEnvironment(env.Id, trx)
		if err != nil {
			return storeFailure(errors.Wrapf(err, "error finding shares for environment '%v'", env.ZId))
		}

		for _, shr := range shrs {
			_, stayLimited := limitedBackends[sdk.BackendMode(shr.BackendMode)]
			if (!bwc.IsScoped() && !stayLimited) || bwc.GetBackendMode() == sdk.BackendMode(shr.BackendMode) {
				switch shr.ShareMode {
				case string(sdk.PublicShareMode):
					if err := relaxPublicShare(a.str, ziti, shr, trx); err != nil {
						var storeErr storeRelaxError
						if errors.As(err, &storeErr) {
							return err
						}
						failures = append(failures, fmt.Sprintf("share '%v': %v", shr.Token, err))
					}
				case string(sdk.PrivateShareMode):
					if err := relaxPrivateShare(a.str, ziti, shr, trx); err != nil {
						var storeErr storeRelaxError
						if errors.As(err, &storeErr) {
							return err
						}
						failures = append(failures, fmt.Sprintf("share '%v': %v", shr.Token, err))
					}
				}
			}
		}
	}

	if len(failures) > 0 {
		return errors.New(strings.Join(failures, "; "))
	}
	return nil
}

// relaxPublicShare restores the frontend dial policy of a public share and the policies of its private
// accesses.
func relaxPublicShare(str *store.Store, ziti *automation.ZitiAutomation, shr *store.Share, trx *sqlx.Tx) error {
	desired, err := desiredPublicDialPolicies(str, shr, trx)
	if err != nil {
		var missing missingPublicFrontendError
		if errors.As(err, &missing) {
			return err
		}
		return storeFailure(err)
	}
	if err := ensureDialPolicies(ziti, shr, desired); err != nil {
		return err
	}
	return relaxAccessFrontends(str, ziti, shr, trx)
}

func relaxPrivateShare(str *store.Store, ziti *automation.ZitiAutomation, shr *store.Share, trx *sqlx.Tx) error {
	return relaxAccessFrontends(str, ziti, shr, trx)
}

// relaxAccessFrontends restores the dial policy of each private access to the share, whatever its mode.
func relaxAccessFrontends(str *store.Store, ziti *automation.ZitiAutomation, shr *store.Share, trx *sqlx.Tx) error {
	desired, err := desiredAccessDialPolicies(str, shr, trx)
	if err != nil {
		return storeFailure(err)
	}
	return ensureDialPolicies(ziti, shr, desired)
}

func ensureDialPolicies(ziti *automation.ZitiAutomation, shr *store.Share, desired []*automation.ServicePolicyOptions) error {
	for _, opts := range desired {
		created, err := ensureDialPolicy(ziti, opts)
		if err != nil {
			return err
		}
		if created {
			dl.Infof("added dial service policy '%v' for share '%v'", opts.Name, shr.Token)
		}
	}
	return nil
}
