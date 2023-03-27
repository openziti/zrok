package limits

import (
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/zrok/controller/store"
	"github.com/sirupsen/logrus"
)

type shareLimitAction struct {
	str  *store.Store
	edge *rest_management_api_client.ZitiEdgeManagement
}

func newShareLimitAction(str *store.Store, edge *rest_management_api_client.ZitiEdgeManagement) *shareLimitAction {
	return &shareLimitAction{str, edge}
}

func (a *shareLimitAction) HandleShare(s *store.Share, rxBytes, txBytes int64, limit *BandwidthPerPeriod) error {
	logrus.Infof("limiting '%v'", s.Token)
	return nil
}
