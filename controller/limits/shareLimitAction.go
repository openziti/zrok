package limits

import (
	"github.com/jmoiron/sqlx"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/sirupsen/logrus"
)

type shareLimitAction struct {
	str  *store.Store
	zCfg *zrokEdgeSdk.Config
}

func newShareLimitAction(str *store.Store, zCfg *zrokEdgeSdk.Config) *shareLimitAction {
	return &shareLimitAction{str, zCfg}
}

func (a *shareLimitAction) HandleShare(shr *store.Share, _, _ int64, _ *BandwidthPerPeriod, trx *sqlx.Tx) error {
	logrus.Infof("limiting '%v'", shr.Token)

	env, err := a.str.GetEnvironment(shr.EnvironmentId, trx)
	if err != nil {
		return err
	}

	edge, err := zrokEdgeSdk.Client(a.zCfg)
	if err != nil {
		return err
	}

	if err := zrokEdgeSdk.DeleteServicePoliciesDial(env.ZId, shr.Token, edge); err != nil {
		return err
	}
	logrus.Infof("removed dial service policy for '%v'", shr.Token)

	return nil
}
