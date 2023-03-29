package limits

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/openziti/edge/rest_management_api_client"
	"github.com/openziti/zrok/controller/emailUi"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/util"
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

	rxLimit := "(unlimited bytes)"
	if limit.Limit.Rx != Unlimited {
		rxLimit = util.BytesToSize(limit.Limit.Rx)
	}
	txLimit := "(unlimited bytes)"
	if limit.Limit.Tx != Unlimited {
		txLimit = util.BytesToSize(limit.Limit.Tx)
	}
	totalLimit := "(unlimited bytes)"
	if limit.Limit.Total != Unlimited {
		totalLimit = util.BytesToSize(limit.Limit.Total)
	}
	detail := fmt.Sprintf("Your account has received %v and sent %v (for a total of %v), which has triggered a transfer limit warning.", util.BytesToSize(rxBytes), util.BytesToSize(txBytes), util.BytesToSize(rxBytes+txBytes)) +
		fmt.Sprintf(" This zrok instance only allows an account to receive %v, send %v, totalling not more than %v for each %v.", rxLimit, txLimit, totalLimit, limit.Period) +
		fmt.Sprintf(" If you exceed the transfer limit, access to your shares will be temporarily disabled (until the last %v falls below the transfer limit)", limit.Period)

	if err := sendLimitWarningEmail(a.cfg, acct.Email, detail); err != nil {
		return errors.Wrapf(err, "error sending limit warning email to '%v'", acct.Email)
	}

	return nil
}