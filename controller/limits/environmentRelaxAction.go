package limits

import (
	"github.com/jmoiron/sqlx"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type environmentRelaxAction struct {
	str  *store.Store
	zCfg *zrokEdgeSdk.Config
}

func newEnvironmentRelaxAction(str *store.Store, zCfg *zrokEdgeSdk.Config) *environmentRelaxAction {
	return &environmentRelaxAction{str, zCfg}
}

func (a *environmentRelaxAction) HandleEnvironment(env *store.Environment, rxBytes, txBytes int64, limit *BandwidthPerPeriod, trx *sqlx.Tx) error {
	logrus.Infof("relaxing '%v'", env.ZId)

	shrs, err := a.str.FindSharesForEnvironment(env.Id, trx)
	if err != nil {
		return errors.Wrapf(err, "error finding shares for environment '%v'", env.ZId)
	}

	edge, err := zrokEdgeSdk.Client(a.zCfg)
	if err != nil {
		return err
	}

	for _, shr := range shrs {
		if !shr.Deleted {
			switch shr.ShareMode {
			case string(sdk.PublicShareMode):
				if err := relaxPublicShare(a.str, edge, shr, trx); err != nil {
					return err
				}
			case string(sdk.PrivateShareMode):
				if err := relaxPrivateShare(a.str, edge, shr, trx); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
