package limits

import (
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/zrok/controller/store"
	"github.com/sirupsen/logrus"
)

type shareWarningAction struct {
	str  *store.Store
	edge *rest_management_api_client.ZitiEdgeManagement
}

func newShareWarningAction(str *store.Store, edge *rest_management_api_client.ZitiEdgeManagement) *shareWarningAction {
	return &shareWarningAction{str, edge}
}

func (a *shareWarningAction) HandleShare(s *store.Share, rxBytes, txBytes int64, limit *BandwidthPerPeriod) error {
	logrus.Infof("warning '%v'", s.Token)
	return nil
}
