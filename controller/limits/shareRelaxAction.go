package limits

import (
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/zrok/controller/store"
	"github.com/sirupsen/logrus"
)

type shareRelaxAction struct {
	str  *store.Store
	edge *rest_management_api_client.ZitiEdgeManagement
}

func newShareRelaxAction(str *store.Store, edge *rest_management_api_client.ZitiEdgeManagement) *shareRelaxAction {
	return &shareRelaxAction{str, edge}
}

func (a *shareRelaxAction) HandleShare(s *store.Share, rxBytes, txBytes int64, limit *BandwidthPerPeriod) error {
	logrus.Infof("relaxing '%v'", s.Token)
	return nil
}
