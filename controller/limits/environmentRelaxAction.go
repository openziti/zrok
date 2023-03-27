package limits

import (
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/zrok/controller/store"
	"github.com/sirupsen/logrus"
)

type environmentRelaxAction struct {
	str  *store.Store
	edge *rest_management_api_client.ZitiEdgeManagement
}

func newEnvironmentRelaxAction(str *store.Store, edge *rest_management_api_client.ZitiEdgeManagement) *environmentRelaxAction {
	return &environmentRelaxAction{str, edge}
}

func (a *environmentRelaxAction) HandleEnvironment(e *store.Environment, rxBytes, txBytes int64, limit *BandwidthPerPeriod) error {
	logrus.Infof("relaxing '%v'", e.ZId)
	return nil
}
