package limits

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/emailUi"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/openziti/zrok/v2/util"
	"github.com/pkg/errors"
)

type warningAction struct {
	str *store.Store
	cfg *emailUi.Config
}

func newWarningAction(cfg *emailUi.Config, str *store.Store) *warningAction {
	return &warningAction{str, cfg}
}

func (a *warningAction) HandleAccount(acct *store.Account, rxBytes, txBytes int64, bwc store.BandwidthClass, _ *userLimits, _ *sqlx.Tx) error {
	dl.Infof("warning '%v'", acct.Email)

	if a.cfg != nil {
		rxLimit := "(store.Unlimited bytes)"
		if bwc.GetRxBytes() != store.Unlimited {
			rxLimit = util.BytesToSize(bwc.GetRxBytes())
		}
		txLimit := "(store.Unlimited bytes)"
		if bwc.GetTxBytes() != store.Unlimited {
			txLimit = util.BytesToSize(bwc.GetTxBytes())
		}
		totalLimit := "(store.Unlimited bytes)"
		if bwc.GetTotalBytes() != store.Unlimited {
			totalLimit = util.BytesToSize(bwc.GetTotalBytes())
		}

		detail := newDetailMessage()
		detail = detail.append("Your account has received %v and sent %v (for a total of %v), which has triggered a transfer limit warning.", util.BytesToSize(rxBytes), util.BytesToSize(txBytes), util.BytesToSize(rxBytes+txBytes))
		detail = detail.append("This zrok instance only allows an account to receive %v, send %v, totalling not more than %v for each %v.", rxLimit, txLimit, totalLimit, time.Duration(bwc.GetPeriodMinutes())*time.Minute)
		detail = detail.append("If you exceed the transfer limit, access to your shares will be temporarily disabled (until the last %v falls below the transfer limit)", time.Duration(bwc.GetPeriodMinutes())*time.Minute)

		if err := sendLimitWarningEmail(a.cfg, acct.Email, detail); err != nil {
			return errors.Wrapf(err, "error sending limit warning email to '%v'", acct.Email)
		}
	} else {
		dl.Warnf("skipping warning email for account limit; no email configuration specified")
	}

	return nil
}
