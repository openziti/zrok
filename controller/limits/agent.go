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

func (a *Agent) CanCreateEnvironment(acctId int, trx *sqlx.Tx) (bool, error) {
	if a.cfg.Enforcing && a.cfg.Environments > Unlimited {
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

func (a *Agent) CanCreateShare(acctId int, trx *sqlx.Tx) (bool, error) {
	if a.cfg.Enforcing && a.cfg.Shares > Unlimited {
		envs, err := a.str.FindEnvironmentsForAccount(acctId, trx)
		if err != nil {
			return false, err
		}
		total := 0
		for i := range envs {
			shrs, err := a.str.FindSharesForEnvironment(envs[i].Id, trx)
			if err != nil {
				return false, errors.Wrapf(err, "unable to find shares for environment '%v'", envs[i].ZId)
			}
			total += len(shrs)
			if total+1 > a.cfg.Shares {
				return false, nil
			}
			logrus.Infof("total = %d", total)
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
			enforced := false
			var enforcedAt time.Time
			if empty, err := a.str.IsAccountLimitJournalEmpty(int(u.AccountId), trx); err == nil && !empty {
				if latest, err := a.str.FindLatestAccountLimitJournal(int(u.AccountId), trx); err == nil {
					enforced = latest.Action == store.LimitAction
					enforcedAt = latest.UpdatedAt
				}
			}

			if !enforced {
				_, err := a.str.CreateAccountLimitJournal(&store.AccountLimitJournal{
					AccountId: int(u.AccountId),
					RxBytes:   u.BackendRx,
					TxBytes:   u.BackendTx,
					Action:    store.LimitAction,
				}, trx)
				if err != nil {
					return err
				}

				logrus.Warnf("enforcing account limit for '#%d': %v", u.AccountId, a.describeLimit(a.cfg.Bandwidth.PerAccount, u.BackendRx, u.BackendTx))

				if err := trx.Commit(); err != nil {
					return err
				}
			} else {
				logrus.Debugf("already enforced limit for account '#%d' at %v", u.AccountId, enforcedAt)
			}

		} else if warning {
			warned := false
			var warnedAt time.Time
			if empty, err := a.str.IsAccountLimitJournalEmpty(int(u.AccountId), trx); err == nil && !empty {
				if latest, err := a.str.FindLatestAccountLimitJournal(int(u.AccountId), trx); err == nil {
					warned = latest.Action == store.WarningAction || latest.Action == store.LimitAction
					warnedAt = latest.UpdatedAt
				}
			}

			if !warned {
				_, err := a.str.CreateAccountLimitJournal(&store.AccountLimitJournal{
					AccountId: int(u.AccountId),
					RxBytes:   u.BackendRx,
					TxBytes:   u.BackendTx,
					Action:    store.WarningAction,
				}, trx)
				if err != nil {
					return err
				}

				logrus.Warnf("warning account '#%d': %v", u.AccountId, a.describeLimit(a.cfg.Bandwidth.PerAccount, u.BackendRx, u.BackendTx))

				if err := trx.Commit(); err != nil {
					return err
				}
			} else {
				logrus.Debugf("already warned account '#%d' at %v", u.AccountId, warnedAt)
			}

		} else {
			if enforce, warning, err := a.checkEnvironmentLimit(u, trx); err == nil {
				if enforce {
					enforced := false
					var enforcedAt time.Time
					if empty, err := a.str.IsEnvironmentLimitJournalEmpty(int(u.EnvironmentId), trx); err == nil && !empty {
						if latest, err := a.str.FindLatestEnvironmentLimitJournal(int(u.EnvironmentId), trx); err == nil {
							enforced = latest.Action == store.LimitAction
							enforcedAt = latest.UpdatedAt
						}
					}

					if !enforced {
						_, err := a.str.CreateEnvironmentLimitJournal(&store.EnvironmentLimitJournal{
							EnvironmentId: int(u.EnvironmentId),
							RxBytes:       u.BackendRx,
							TxBytes:       u.BackendTx,
							Action:        store.LimitAction,
						}, trx)
						if err != nil {
							return err
						}

						logrus.Warnf("enforcing environment limit for environment '#%d': %v", u.EnvironmentId, a.describeLimit(a.cfg.Bandwidth.PerEnvironment, u.BackendRx, u.BackendTx))

						if err := trx.Commit(); err != nil {
							return err
						}
					} else {
						logrus.Debugf("already enforced limit for environment '#%d' at %v", u.EnvironmentId, enforcedAt)
					}

				} else if warning {
					warned := false
					var warnedAt time.Time
					if empty, err := a.str.IsEnvironmentLimitJournalEmpty(int(u.EnvironmentId), trx); err == nil && !empty {
						if latest, err := a.str.FindLatestEnvironmentLimitJournal(int(u.EnvironmentId), trx); err == nil {
							warned = latest.Action == store.WarningAction || latest.Action == store.LimitAction
							warnedAt = latest.UpdatedAt
						}
					}

					if !warned {
						_, err := a.str.CreateEnvironmentLimitJournal(&store.EnvironmentLimitJournal{
							EnvironmentId: int(u.EnvironmentId),
							RxBytes:       u.BackendRx,
							TxBytes:       u.BackendTx,
							Action:        store.WarningAction,
						}, trx)
						if err != nil {
							return err
						}

						logrus.Warnf("warning environment '#%d': %v", u.EnvironmentId, a.describeLimit(a.cfg.Bandwidth.PerEnvironment, u.BackendRx, u.BackendTx))

						if err := trx.Commit(); err != nil {
							return err
						}
					} else {
						logrus.Debugf("already warned environment '#%d' at %v", u.EnvironmentId, warnedAt)
					}

				} else {
					if enforce, warning, err := a.checkShareLimit(u); err == nil {
						if enforce {
							shr, err := a.str.FindShareWithToken(u.ShareToken, trx)
							if err != nil {
								return err
							}

							enforced := false
							var enforcedAt time.Time
							if empty, err := a.str.IsShareLimitJournalEmpty(shr.Id, trx); err == nil && !empty {
								if latest, err := a.str.FindLatestShareLimitJournal(shr.Id, trx); err == nil {
									enforced = latest.Action == store.LimitAction
									enforcedAt = latest.UpdatedAt
								}
							}

							if !enforced {
								_, err := a.str.CreateShareLimitJournal(&store.ShareLimitJournal{
									ShareId: shr.Id,
									RxBytes: u.BackendRx,
									TxBytes: u.BackendTx,
									Action:  store.LimitAction,
								}, trx)
								if err != nil {
									return err
								}

								logrus.Warnf("enforcing share limit for share '%v': %v", shr.Token, a.describeLimit(a.cfg.Bandwidth.PerShare, u.BackendRx, u.BackendTx))

								if err := trx.Commit(); err != nil {
									return err
								}
							} else {
								logrus.Debugf("already enforced limit for share '%v' at %v", shr.Token, enforcedAt)
							}

						} else if warning {
							shr, err := a.str.FindShareWithToken(u.ShareToken, trx)
							if err != nil {
								return err
							}

							warned := false
							var warnedAt time.Time
							if empty, err := a.str.IsShareLimitJournalEmpty(shr.Id, trx); err == nil && !empty {
								if latest, err := a.str.FindLatestShareLimitJournal(shr.Id, trx); err == nil {
									warned = latest.Action == store.WarningAction || latest.Action == store.LimitAction
									warnedAt = latest.UpdatedAt
								}
							}

							if !warned {
								_, err := a.str.CreateShareLimitJournal(&store.ShareLimitJournal{
									ShareId: shr.Id,
									RxBytes: u.BackendRx,
									TxBytes: u.BackendTx,
									Action:  store.WarningAction,
								}, trx)
								if err != nil {
									return err
								}

								logrus.Warnf("warning share '%v': %v", shr.Token, a.describeLimit(a.cfg.Bandwidth.PerShare, u.BackendRx, u.BackendTx))

								if err := trx.Commit(); err != nil {
									return err
								}
							} else {
								logrus.Debugf("already warned share '%v' at %v", shr.Token, warnedAt)
							}
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
