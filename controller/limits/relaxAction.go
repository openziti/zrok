package limits

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/edge-api/rest_model"
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

func relaxPublicShare(str *store.Store, ziti *automation.ZitiAutomation, shr *store.Share, trx *sqlx.Tx) error {
	env, err := str.GetEnvironment(shr.EnvironmentId, trx)
	if err != nil {
		return storeFailure(errors.Wrap(err, "error finding environment"))
	}
	if shr.FrontendSelection == nil {
		return errors.Errorf("share '%v' has no frontend selection", shr.Token)
	}

	fe, err := str.FindFrontendPubliclyNamed(*shr.FrontendSelection, trx)
	if err != nil {
		return storeFailure(errors.Wrapf(err, "error finding frontend name '%v' for '%v'", *shr.FrontendSelection, shr.Token))
	}
	policyName := env.ZId + "-" + shr.ZId + "-dial"
	policies, err := ziti.ServicePolicies.Find(&automation.FilterOptions{Filter: automation.BuildFilter("name", policyName)})
	if err != nil {
		return errors.Wrapf(err, "error finding dial service policy for '%v'", shr.Token)
	}
	if len(policies) > 0 {
		dl.Debugf("dial service policy '%v' already exists", policyName)
		return nil
	}

	opts := &automation.ServicePolicyOptions{
		BaseOptions: automation.BaseOptions{
			Name: policyName,
			Tags: automation.ZrokShareTags(shr.Token),
		},
		IdentityRoles: []string{"@" + fe.ZId},
		ServiceRoles:  []string{"@" + shr.ZId},
		PolicyType:    rest_model.DialBindDial,
		Semantic:      rest_model.SemanticAllOf,
	}

	if _, err := ziti.ServicePolicies.CreateDial(opts); err != nil {
		return errors.Wrapf(err, "error creating dial service policy for '%v'", shr.Token)
	}
	dl.Infof("added dial service policy for '%v'", shr.Token)
	return nil
}

func relaxPrivateShare(str *store.Store, ziti *automation.ZitiAutomation, shr *store.Share, trx *sqlx.Tx) error {
	fes, err := str.FindFrontendsForPrivateShare(shr.Id, trx)
	if err != nil {
		return storeFailure(errors.Wrapf(err, "error finding frontends for share '%v'", shr.Token))
	}
	for _, fe := range fes {
		if fe.EnvironmentId != nil {
			env, err := str.GetEnvironment(*fe.EnvironmentId, trx)
			if err != nil {
				return storeFailure(errors.Wrapf(err, "error getting environment for frontend '%v'", fe.Token))
			}
			policyName := fe.Token + "-" + env.ZId + "-" + shr.ZId + "-dial"
			policies, err := ziti.ServicePolicies.Find(&automation.FilterOptions{Filter: automation.BuildFilter("name", policyName)})
			if err != nil {
				return errors.Wrapf(err, "error finding dial policy for frontend '%v'", fe.Token)
			}
			if len(policies) > 0 {
				dl.Debugf("dial service policy '%v' already exists", policyName)
				continue
			}

			opts := &automation.ServicePolicyOptions{
				BaseOptions: automation.BaseOptions{
					Name: policyName,
					Tags: automation.NewTags().
						WithZrok().
						WithShareToken(shr.Token).
						WithTag("zrokEnvironmentZId", env.ZId).
						WithTag("zrokFrontendToken", fe.Token),
				},
				IdentityRoles: []string{"@" + env.ZId},
				ServiceRoles:  []string{"@" + shr.ZId},
				PolicyType:    rest_model.DialBindDial,
				Semantic:      rest_model.SemanticAllOf,
			}

			if _, err := ziti.ServicePolicies.CreateDial(opts); err != nil {
				return errors.Wrapf(err, "unable to create dial policy for frontend '%v'", fe.Token)
			}

			dl.Infof("added dial service policy for share '%v' to private frontend '%v'", shr.Token, fe.Token)
		}
	}
	return nil
}
