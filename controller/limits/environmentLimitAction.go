package limits

import (
	"github.com/jmoiron/sqlx"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type environmentLimitAction struct {
	str  *store.Store
	zCfg *zrokEdgeSdk.Config
}

func newEnvironmentLimitAction(str *store.Store, zCfg *zrokEdgeSdk.Config) *environmentLimitAction {
	return &environmentLimitAction{str, zCfg}
}

func (a *environmentLimitAction) HandleEnvironment(env *store.Environment, _, _ int64, _ *BandwidthPerPeriod, trx *sqlx.Tx) error {
	logrus.Infof("limiting '%v'", env.ZId)

	shrs, err := a.str.FindSharesForEnvironment(env.Id, trx)
	if err != nil {
		return errors.Wrapf(err, "error finding shares for environment '%v'", env.ZId)
	}

	edge, err := zrokEdgeSdk.Client(a.zCfg)
	if err != nil {
		return err
	}

	for _, shr := range shrs {
		if err := zrokEdgeSdk.DeleteServicePoliciesDial(env.ZId, shr.Token, edge); err != nil {
			return errors.Wrapf(err, "error deleting dial service policy for '%v'", shr.Token)
		}
		logrus.Infof("removed dial service policy for share '%v' of environment '%v'", shr.Token, env.ZId)
	}

	return nil
}
