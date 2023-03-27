package limits

import (
	"github.com/jmoiron/sqlx"
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/zrok/controller/store"
	"github.com/sirupsen/logrus"
)

type environmentLimitAction struct {
	str  *store.Store
	edge *rest_management_api_client.ZitiEdgeManagement
}

func newEnvironmentLimitAction(str *store.Store, edge *rest_management_api_client.ZitiEdgeManagement) *environmentLimitAction {
	return &environmentLimitAction{str, edge}
}

func (a *environmentLimitAction) HandleEnvironment(e *store.Environment, rxBytes, txBytes int64, limit *BandwidthPerPeriod, trx *sqlx.Tx) error {
	logrus.Infof("limiting '%v'", e.ZId)
	return nil
}
