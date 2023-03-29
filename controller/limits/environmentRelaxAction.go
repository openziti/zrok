package limits

import (
	"github.com/jmoiron/sqlx"
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/zrok/controller/store"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type environmentRelaxAction struct {
	str  *store.Store
	edge *rest_management_api_client.ZitiEdgeManagement
}

func newEnvironmentRelaxAction(str *store.Store, edge *rest_management_api_client.ZitiEdgeManagement) *environmentRelaxAction {
	return &environmentRelaxAction{str, edge}
}

func (a *environmentRelaxAction) HandleEnvironment(env *store.Environment, rxBytes, txBytes int64, limit *BandwidthPerPeriod, trx *sqlx.Tx) error {
	logrus.Infof("relaxing '%v'", env.ZId)

	shrs, err := a.str.FindSharesForEnvironment(env.Id, trx)
	if err != nil {
		return errors.Wrapf(err, "error finding shares for environment '%v'", env.ZId)
	}

	for _, shr := range shrs {
		if !shr.Deleted {
			switch shr.ShareMode {
			case "public":
				if err := relaxPublicShare(a.str, a.edge, shr, trx); err != nil {
					return err
				}
			case "private":
				if err := relaxPrivateShare(a.str, a.edge, shr, trx); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
