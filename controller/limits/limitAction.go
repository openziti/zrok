package limits

import (
	"github.com/jmoiron/sqlx"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type accountLimitAction struct {
	str  *store.Store
	zCfg *zrokEdgeSdk.Config
}

func newAccountLimitAction(str *store.Store, zCfg *zrokEdgeSdk.Config) *accountLimitAction {
	return &accountLimitAction{str, zCfg}
}

func (a *accountLimitAction) HandleAccount(acct *store.Account, _, _ int64, _ *BandwidthPerPeriod, trx *sqlx.Tx) error {
	logrus.Infof("limiting '%v'", acct.Email)

	envs, err := a.str.FindEnvironmentsForAccount(acct.Id, trx)
	if err != nil {
		return errors.Wrapf(err, "error finding environments for account '%v'", acct.Email)
	}

	edge, err := zrokEdgeSdk.Client(a.zCfg)
	if err != nil {
		return err
	}

	for _, env := range envs {
		shrs, err := a.str.FindSharesForEnvironment(env.Id, trx)
		if err != nil {
			return errors.Wrapf(err, "error finding shares for environment '%v'", env.ZId)
		}

		for _, shr := range shrs {
			if err := zrokEdgeSdk.DeleteServicePoliciesDial(env.ZId, shr.Token, edge); err != nil {
				return errors.Wrapf(err, "error deleting dial service policy for '%v'", shr.Token)
			}
			logrus.Infof("removed dial service policy for share '%v' of environment '%v'", shr.Token, env.ZId)
		}
	}

	return nil
}
