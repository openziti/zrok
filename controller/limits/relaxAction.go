package limits

import (
	"github.com/jmoiron/sqlx"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/edge-api/rest_model"
	"github.com/openziti/zrok/v2/controller/automation"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/openziti/zrok/v2/sdk/golang/sdk"
	"github.com/pkg/errors"
)

type relaxAction struct {
	str  *store.Store
	zCfg *automation.Config
}

func newRelaxAction(str *store.Store, zCfg *automation.Config) *relaxAction {
	return &relaxAction{str, zCfg}
}

func (a *relaxAction) HandleAccount(acct *store.Account, _, _ int64, bwc store.BandwidthClass, _ *userLimits, trx *sqlx.Tx) error {
	dl.Debugf("relaxing '%v'", acct.Email)

	envs, err := a.str.FindEnvironmentsForAccount(acct.Id, trx)
	if err != nil {
		return errors.Wrapf(err, "error finding environments for account '%v'", acct.Email)
	}

	jes, err := a.str.FindAllLatestBandwidthLimitJournalForAccount(acct.Id, trx)
	if err != nil {
		return errors.Wrapf(err, "error finding latest bandwidth limit journal entries for account '%v'", acct.Email)
	}
	limitedBackends := make(map[sdk.BackendMode]bool)
	for _, je := range jes {
		if je.LimitClassId != nil {
			lc, err := a.str.GetLimitClass(*je.LimitClassId, trx)
			if err != nil {
				return err
			}
			if lc.BackendMode != nil && lc.LimitAction == store.LimitLimitAction {
				limitedBackends[*lc.BackendMode] = true
			}
		}
	}

	ziti, err := automation.NewZitiAutomation(a.zCfg)
	if err != nil {
		return err
	}

	for _, env := range envs {
		shrs, err := a.str.FindSharesForEnvironment(env.Id, trx)
		if err != nil {
			return errors.Wrapf(err, "error finding shares for environment '%v'", env.ZId)
		}

		for _, shr := range shrs {
			_, stayLimited := limitedBackends[sdk.BackendMode(shr.BackendMode)]
			if (!bwc.IsScoped() && !stayLimited) || bwc.GetBackendMode() == sdk.BackendMode(shr.BackendMode) {
				switch shr.ShareMode {
				case string(sdk.PublicShareMode):
					if err := relaxPublicShare(a.str, ziti, shr, trx); err != nil {
						dl.Errorf("error relaxing public share '%v' for account '%v' (ignoring): %v", shr.Token, acct.Email, err)
					}
				case string(sdk.PrivateShareMode):
					if err := relaxPrivateShare(a.str, ziti, shr, trx); err != nil {
						dl.Errorf("error relaxing private share '%v' for account '%v' (ignoring): %v", shr.Token, acct.Email, err)
					}
				}
			}
		}
	}

	return nil
}

func relaxPublicShare(str *store.Store, ziti *automation.ZitiAutomation, shr *store.Share, trx *sqlx.Tx) error {
	env, err := str.GetEnvironment(shr.EnvironmentId, trx)
	if err != nil {
		return errors.Wrap(err, "error finding environment")
	}

	fe, err := str.FindFrontendPubliclyNamed(*shr.FrontendSelection, trx)
	if err != nil {
		return errors.Wrapf(err, "error finding frontend name '%v' for '%v'", *shr.FrontendSelection, shr.Token)
	}

	opts := &automation.ServicePolicyOptions{
		BaseOptions: automation.BaseOptions{
			Name: env.ZId + "-" + shr.ZId + "-dial",
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
		return errors.Wrapf(err, "error finding frontends for share '%v'", shr.Token)
	}
	for _, fe := range fes {
		if fe.EnvironmentId != nil {
			env, err := str.GetEnvironment(*fe.EnvironmentId, trx)
			if err != nil {
				return errors.Wrapf(err, "error getting environment for frontend '%v'", fe.Token)
			}

			opts := &automation.ServicePolicyOptions{
				BaseOptions: automation.BaseOptions{
					Name: fe.Token + "-" + env.ZId + "-" + shr.ZId + "-dial",
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
