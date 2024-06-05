package limits

import (
	"github.com/jmoiron/sqlx"
	"github.com/openziti/zrok/controller/emailUi"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
)

type warningAction struct {
	str *store.Store
	cfg *emailUi.Config
}

func newWarningAction(cfg *emailUi.Config, str *store.Store) *warningAction {
	return &warningAction{str, cfg}
}

func (a *warningAction) HandleAccount(acct *store.Account, rxBytes, txBytes int64, limit store.BandwidthClass, _ *sqlx.Tx) error {
	logrus.Infof("warning '%v'", acct.Email)

	if a.cfg != nil {
		rxLimit := "(store.Unlimited bytes)"
		if limit.GetRxBytes() != store.Unlimited {
			rxLimit = util.BytesToSize(limit.GetRxBytes())
		}
		txLimit := "(store.Unlimited bytes)"
		if limit.GetTxBytes() != store.Unlimited {
			txLimit = util.BytesToSize(limit.GetTxBytes())
		}
		totalLimit := "(store.Unlimited bytes)"
		if limit.GetTotalBytes() != store.Unlimited {
			totalLimit = util.BytesToSize(limit.GetTotalBytes())
		}

		detail := newDetailMessage()
		detail = detail.append("Your account has received %v and sent %v (for a total of %v), which has triggered a transfer limit warning.", util.BytesToSize(rxBytes), util.BytesToSize(txBytes), util.BytesToSize(rxBytes+txBytes))
		detail = detail.append("This zrok instance only allows an account to receive %v, send %v, totalling not more than %v for each %v.", rxLimit, txLimit, totalLimit, time.Duration(limit.GetPeriodMinutes())*time.Minute)
		detail = detail.append("If you exceed the transfer limit, access to your shares will be temporarily disabled (until the last %v falls below the transfer limit)", time.Duration(limit.GetPeriodMinutes())*time.Minute)

		if err := sendLimitWarningEmail(a.cfg, acct.Email, detail); err != nil {
			return errors.Wrapf(err, "error sending limit warning email to '%v'", acct.Email)
		}
	} else {
		logrus.Warnf("skipping warning email for account limit; no email configuration specified")
	}

	return nil
}
