package limits

import (
	"github.com/jmoiron/sqlx"
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/zrok/controller/emailUi"
	"github.com/openziti/zrok/controller/store"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type accountWarningAction struct {
	str  *store.Store
	edge *rest_management_api_client.ZitiEdgeManagement
	cfg  *emailUi.Config
}

func newAccountWarningAction(cfg *emailUi.Config, str *store.Store, edge *rest_management_api_client.ZitiEdgeManagement) *accountWarningAction {
	return &accountWarningAction{str, edge, cfg}
}

func (a *accountWarningAction) HandleAccount(acct *store.Account, rxBytes, txBytes int64, limit *BandwidthPerPeriod, trx *sqlx.Tx) error {
	logrus.Infof("warning '%v'", acct.Email)

	if err := sendLimitWarningEmail(a.cfg, acct.Email, limit, rxBytes, txBytes); err != nil {
		return errors.Wrapf(err, "error sending limit warning email to '%v'", acct.Email)
	}

	return nil
}
