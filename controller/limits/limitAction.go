package limits

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/controller/automation"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/pkg/errors"
)

type limitAction struct {
	str  *store.Store
	zCfg *automation.Config
}

func newLimitAction(str *store.Store, zCfg *automation.Config) *limitAction {
	return &limitAction{str, zCfg}
}

func (a *limitAction) HandleAccount(acct *store.Account, _, _ int64, bwc store.BandwidthClass, ul *userLimits, trx *sqlx.Tx) error {
	envs, err := a.str.FindEnvironmentsForAccount(acct.Id, trx)
	if err != nil {
		return errors.Wrapf(err, "error finding environments for account '%v'", acct.Email)
	}

	ziti, err := automation.NewZitiAutomation(a.zCfg)
	if err != nil {
		return err
	}

	ignoreBackends := ul.ignoreBackends(bwc)
	for _, env := range envs {
		shrs, err := a.str.FindSharesForEnvironment(env.Id, trx)
		if err != nil {
			return errors.Wrapf(err, "error finding shares for environment '%v'", env.ZId)
		}

		for _, shr := range shrs {
			if _, ignore := ignoreBackends[sdk.BackendMode(shr.BackendMode)]; !ignore {
				// delete dial polcies for share
				filter := fmt.Sprintf("tags.zrokShareToken=\"%v\" and type=1", shr.Token)
				if err := ziti.ServicePolicies.DeleteWithFilter(filter); err != nil {
					return errors.Wrapf(err, "error deleting dial service policy for '%v'", shr.Token)
				}
				dl.Infof("removed dial service policy for share '%v' of environment '%v'", shr.Token, env.ZId)
			} else {
				dl.Debugf("ignoring share '%v' for '%v' with backend mode '%v'", shr.Token, acct.Email, shr.BackendMode)
			}
		}
	}

	return nil
}
