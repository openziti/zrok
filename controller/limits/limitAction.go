package limits

import (
	"github.com/jmoiron/sqlx"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type limitAction struct {
	str  *store.Store
	zCfg *zrokEdgeSdk.Config
}

func newLimitAction(str *store.Store, zCfg *zrokEdgeSdk.Config) *limitAction {
	return &limitAction{str, zCfg}
}

func (a *limitAction) HandleAccount(acct *store.Account, _, _ int64, bwc store.BandwidthClass, ul *userLimits, trx *sqlx.Tx) error {
	logrus.Infof("limiting '%v'", acct.Email)

	envs, err := a.str.FindEnvironmentsForAccount(acct.Id, trx)
	if err != nil {
		return errors.Wrapf(err, "error finding environments for account '%v'", acct.Email)
	}

	edge, err := zrokEdgeSdk.Client(a.zCfg)
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
				if err := zrokEdgeSdk.DeleteServicePoliciesDial(env.ZId, shr.Token, edge); err != nil {
					return errors.Wrapf(err, "error deleting dial service policy for '%v'", shr.Token)
				}
				logrus.Infof("removed dial service policy for share '%v' of environment '%v'", shr.Token, env.ZId)
			} else {
				logrus.Debugf("ignoring share '%v' for '%v' with backend mode '%v'", shr.Token, acct.Email, shr.BackendMode)
			}
		}
	}

	return nil
}
