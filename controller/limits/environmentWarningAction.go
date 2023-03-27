package limits

import (
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/zrok/controller/store"
	"github.com/sirupsen/logrus"
)

type environmentWarningAction struct {
	str  *store.Store
	edge *rest_management_api_client.ZitiEdgeManagement
}

func newEnvironmentWarningAction(str *store.Store, edge *rest_management_api_client.ZitiEdgeManagement) *environmentWarningAction {
	return &environmentWarningAction{str, edge}
}

func (a *environmentWarningAction) HandleEnvironment(e *store.Environment, rxBytes, txBytes int64, limit *BandwidthPerPeriod) error {
	logrus.Infof("warning '%v'", e.ZId)
	return nil
}
