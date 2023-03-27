package limits

import (
	"github.com/jmoiron/sqlx"
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/zrok/controller/store"
	"github.com/sirupsen/logrus"
)

type accountRelaxAction struct {
	str  *store.Store
	edge *rest_management_api_client.ZitiEdgeManagement
}

func newAccountRelaxAction(str *store.Store, edge *rest_management_api_client.ZitiEdgeManagement) *accountRelaxAction {
	return &accountRelaxAction{str, edge}
}

func (a *accountRelaxAction) HandleAccount(acct *store.Account, rxBytes, txBytes int64, limit *BandwidthPerPeriod, trx *sqlx.Tx) error {
	logrus.Infof("relaxing '%v'", acct.Email)
	return nil
}
