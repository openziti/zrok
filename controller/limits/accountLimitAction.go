package limits

import (
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/zrok/controller/store"
	"github.com/sirupsen/logrus"
)

type accountLimitAction struct {
	str  *store.Store
	edge *rest_management_api_client.ZitiEdgeManagement
}

func newAccountLimitAction(str *store.Store, edge *rest_management_api_client.ZitiEdgeManagement) *accountLimitAction {
	return &accountLimitAction{str, edge}
}

func (a *accountLimitAction) HandleAccount(acct *store.Account, rxBytes, txBytes int64, limit *BandwidthPerPeriod) error {
	logrus.Infof("limiting '%v'", acct.Email)
	return nil
}
