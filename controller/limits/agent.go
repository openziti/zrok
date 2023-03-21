package limits

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/openziti/zrok/controller/metrics"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/openziti/zrok/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
)

type Agent struct {
	cfg   *Config
	ifx   *influxReader
	zCfg  *zrokEdgeSdk.Config
	str   *store.Store
	queue chan *metrics.Usage
	close chan struct{}
	join  chan struct{}
}

func NewAgent(cfg *Config, ifxCfg *metrics.InfluxConfig, zCfg *zrokEdgeSdk.Config, str *store.Store) (*Agent, error) {
	return &Agent{
		cfg:   cfg,
		ifx:   newInfluxReader(ifxCfg),
		zCfg:  zCfg,
		str:   str,
		queue: make(chan *metrics.Usage, 1024),
		close: make(chan struct{}),
		join:  make(chan struct{}),
	}, nil
}

func (a *Agent) Start() {
	go a.run()
}

func (a *Agent) Stop() {
	close(a.close)
	<-a.join
}

func (a *Agent) CanCreateEnvironment(acctId int) (bool, error) {
	if a.cfg.Environments > Unlimited {
		trx, err := a.str.Begin()
		if err != nil {
			return false, errors.Wrap(err, "error creating transaction")
		}
		defer func() { _ = trx.Rollback() }()

		envs, err := a.str.FindEnvironmentsForAccount(acctId, trx)
		if err != nil {
			return false, err
		}
		if len(envs)+1 > a.cfg.Environments {
			return false, nil
		}
	}
	return true, nil
}

func (a *Agent) Handle(u *metrics.Usage) error {
	logrus.Debugf("handling: %v", u)
	a.queue <- u
	return nil
}

func (a *Agent) run() {
	logrus.Info("started")
	defer logrus.Info("stopped")

mainLoop:
	for {
		select {
		case usage := <-a.queue:
			a.enforce(usage)

		case <-time.After(a.cfg.Cycle):
			logrus.Info("inspection cycle")

		case <-a.close:
			close(a.join)
			break mainLoop
		}
	}
}

func (a *Agent) enforce(u *metrics.Usage) error {
	trx, err := a.str.Begin()
	if err != nil {
		return errors.Wrap(err, "error starting transaction")
	}
	defer func() { _ = trx.Rollback() }()

	if enforce, warning, err := a.checkAccountLimits(u, trx); err == nil {
		if enforce {
			logrus.Warn("enforcing account limit")
		} else if warning {
			logrus.Warn("reporting account warning")
		} else {
			if enforce, warning, err := a.checkEnvironmentLimit(u, trx); err == nil {
				if enforce {
					logrus.Warn("enforcing environment limit")
				} else if warning {
					logrus.Warn("reporting environment warning")
				} else {
					if enforce, warning, err := a.checkShareLimit(u); err == nil {
						if enforce {
							logrus.Warn("enforcing share limit")
						} else if warning {
							logrus.Warn("reporting share warning")
						}
					} else {
						logrus.Error(err)
					}
				}
			} else {
				logrus.Error(err)
			}
		}
	} else {
		logrus.Error(err)
	}

	return nil
}

func (a *Agent) checkAccountLimits(u *metrics.Usage, trx *sqlx.Tx) (enforce, warning bool, err error) {
	acct, err := a.str.GetAccount(int(u.AccountId), trx)
	if err != nil {
		return false, false, errors.Wrapf(err, "error getting account '%d'", u.AccountId)
	}

	period := 24 * time.Hour
	limit := DefaultBandwidthPerPeriod()
	if a.cfg.Bandwidth != nil && a.cfg.Bandwidth.PerAccount != nil {
		limit = a.cfg.Bandwidth.PerAccount
	}
	if limit.Period > 0 {
		period = limit.Period
	}
	rx, tx, err := a.ifx.totalRxTxForAccount(u.AccountId, period)
	if err != nil {
		logrus.Error(err)
	}

	enforce, warning = a.checkLimit(limit, rx, tx)
	if enforce || warning {
		logrus.Warnf("'%v': %v", acct.Email, a.describeLimit(limit, rx, tx))
	}

	return enforce, warning, nil
}

func (a *Agent) checkEnvironmentLimit(u *metrics.Usage, trx *sqlx.Tx) (enforce, warning bool, err error) {
	env, err := a.str.GetEnvironment(int(u.EnvironmentId), trx)
	if err != nil {
		return false, false, errors.Wrapf(err, "error getting account '%d'", u.EnvironmentId)
	}

	period := 24 * time.Hour
	limit := DefaultBandwidthPerPeriod()
	if a.cfg.Bandwidth != nil && a.cfg.Bandwidth.PerEnvironment != nil {
		limit = a.cfg.Bandwidth.PerEnvironment
	}
	if limit.Period > 0 {
		period = limit.Period
	}
	rx, tx, err := a.ifx.totalRxTxForEnvironment(u.EnvironmentId, period)
	if err != nil {
		logrus.Error(err)
	}

	enforce, warning = a.checkLimit(limit, rx, tx)
	if enforce || warning {
		logrus.Warnf("'%v': %v", env.ZId, a.describeLimit(limit, rx, tx))
	}

	return enforce, warning, nil
}

func (a *Agent) checkShareLimit(u *metrics.Usage) (enforce, warning bool, err error) {
	period := 24 * time.Hour
	limit := DefaultBandwidthPerPeriod()
	if a.cfg.Bandwidth != nil && a.cfg.Bandwidth.PerShare != nil {
		limit = a.cfg.Bandwidth.PerShare
	}
	if limit.Period > 0 {
		period = limit.Period
	}
	rx, tx, err := a.ifx.totalRxTxForShare(u.ShareToken, period)
	if err != nil {
		logrus.Error(err)
	}

	enforce, warning = a.checkLimit(limit, rx, tx)
	if enforce || warning {
		logrus.Warnf("'%v': %v", u.ShareToken, a.describeLimit(limit, rx, tx))
	}

	return enforce, warning, nil
}

func (a *Agent) checkLimit(cfg *BandwidthPerPeriod, rx, tx int64) (enforce, warning bool) {
	if cfg.Limit.Rx != Unlimited && rx > cfg.Limit.Rx {
		return true, false
	}
	if cfg.Limit.Tx != Unlimited && tx > cfg.Limit.Tx {
		return true, false
	}
	if cfg.Limit.Total != Unlimited && rx+tx > cfg.Limit.Total {
		return true, false
	}

	if cfg.Warning.Rx != Unlimited && rx > cfg.Warning.Rx {
		return false, true
	}
	if cfg.Warning.Tx != Unlimited && tx > cfg.Warning.Tx {
		return false, true
	}
	if cfg.Warning.Total != Unlimited && rx+tx > cfg.Warning.Total {
		return false, true
	}

	return false, false
}

func (a *Agent) describeLimit(cfg *BandwidthPerPeriod, rx, tx int64) string {
	out := ""

	if cfg.Limit.Rx != Unlimited && rx > cfg.Limit.Rx {
		out += fmt.Sprintf("['%v' over rx limit '%v']", util.BytesToSize(rx), util.BytesToSize(cfg.Limit.Rx))
	}
	if cfg.Limit.Tx != Unlimited && tx > cfg.Limit.Tx {
		out += fmt.Sprintf("['%v' over tx limit '%v']", util.BytesToSize(tx), util.BytesToSize(cfg.Limit.Tx))
	}
	if cfg.Limit.Total != Unlimited && rx+tx > cfg.Limit.Total {
		out += fmt.Sprintf("['%v' over total limit '%v']", util.BytesToSize(rx+tx), util.BytesToSize(cfg.Limit.Total))
	}

	if cfg.Warning.Rx != Unlimited && rx > cfg.Warning.Rx {
		out += fmt.Sprintf("['%v' over rx warning '%v']", util.BytesToSize(rx), util.BytesToSize(cfg.Warning.Rx))
	}
	if cfg.Warning.Tx != Unlimited && tx > cfg.Warning.Tx {
		out += fmt.Sprintf("['%v' over tx warning '%v']", util.BytesToSize(tx), util.BytesToSize(cfg.Warning.Tx))
	}
	if cfg.Warning.Total != Unlimited && rx+tx > cfg.Warning.Total {
		out += fmt.Sprintf("['%v' over total warning '%v']", util.BytesToSize(rx+tx), util.BytesToSize(cfg.Warning.Total))
	}

	return out
}
