package limits

import (
	"github.com/jmoiron/sqlx"
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/zrok/controller/store"
	"github.com/sirupsen/logrus"
)

type accountWarningAction struct {
	str  *store.Store
	edge *rest_management_api_client.ZitiEdgeManagement
}

func newAccountWarningAction(str *store.Store, edge *rest_management_api_client.ZitiEdgeManagement) *accountWarningAction {
	return &accountWarningAction{str, edge}
}

func (a *accountWarningAction) HandleAccount(acct *store.Account, rxBytes, txBytes int64, limit *BandwidthPerPeriod, trx *sqlx.Tx) error {
	logrus.Infof("warning '%v'", acct.Email)
	return nil
}
