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
	cfg                *Config
	ifx                *influxReader
	zCfg               *zrokEdgeSdk.Config
	str                *store.Store
	queue              chan *metrics.Usage
	acctWarningEnforce []AccountAction
	acctLimitEnforce   []AccountAction
	acctLimitRelax     []AccountAction
	envWarningEnforce  []EnvironmentAction
	envLimitEnforce    []EnvironmentAction
	envLimitRelax      []EnvironmentAction
	shrWarningEnforce  []ShareAction
	shrLimitEnforce    []ShareAction
	shrLimitRelax      []ShareAction
	close              chan struct{}
	join               chan struct{}
}

func NewAgent(cfg *Config, ifxCfg *metrics.InfluxConfig, zCfg *zrokEdgeSdk.Config, str *store.Store) (*Agent, error) {
	edge, err := zrokEdgeSdk.Client(zCfg)
	if err != nil {
		return nil, err
	}
	a := &Agent{
		cfg:               cfg,
		ifx:               newInfluxReader(ifxCfg),
		zCfg:              zCfg,
		str:               str,
		queue:             make(chan *metrics.Usage, 1024),
		shrWarningEnforce: []ShareAction{newShareWarningAction(str, edge)},
		shrLimitEnforce:   []ShareAction{newShareLimitAction(str, edge)},
		shrLimitRelax:     []ShareAction{newShareRelaxAction(str, edge)},
		close:             make(chan struct{}),
		join:              make(chan struct{}),
	}
	return a, nil
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
			if err := a.enforce(usage); err != nil {
				logrus.Errorf("error running enforcement: %v", err)
			}

		case <-time.After(a.cfg.Cycle):
			if err := a.relax(); err != nil {
				logrus.Errorf("error running relax cycle: %v", err)
			}

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

	if enforce, warning, err := a.checkAccountLimit(u.AccountId); err == nil {
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

				logrus.Warnf("enforcing account limit for '#%d'", u.AccountId)

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

				logrus.Warnf("warning account '#%d'", u.AccountId)

				if err := trx.Commit(); err != nil {
					return err
				}
			} else {
				logrus.Debugf("already warned account '#%d' at %v", u.AccountId, warnedAt)
			}

		} else {
			if enforce, warning, err := a.checkEnvironmentLimit(u.EnvironmentId); err == nil {
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

						logrus.Warnf("enforcing environment limit for environment '#%d'", u.EnvironmentId)

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

						logrus.Warnf("warning environment '#%d'", u.EnvironmentId)

						if err := trx.Commit(); err != nil {
							return err
						}
					} else {
						logrus.Debugf("already warned environment '#%d' at %v", u.EnvironmentId, warnedAt)
					}

				} else {
					if enforce, warning, err := a.checkShareLimit(u.ShareToken); err == nil {
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

								logrus.Warnf("enforcing share limit for share '%v'", shr.Token)

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

								logrus.Warnf("warning share '%v'", shr.Token)

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

func (a *Agent) relax() error {
	logrus.Info("relaxing")

	trx, err := a.str.Begin()
	if err != nil {
		return errors.Wrap(err, "error starting transaction")
	}
	defer func() { _ = trx.Rollback() }()

	commit := false

	if sljs, err := a.str.FindAllLatestShareLimitJournal(trx); err == nil {
		for _, slj := range sljs {
			if shr, err := a.str.GetShare(slj.ShareId, trx); err == nil {
				switch slj.Action {
				case store.WarningAction:
					if enforce, warning, err := a.checkShareLimit(shr.Token); err == nil {
						if !enforce && !warning {
							logrus.Infof("relaxing warning for share '%v'", shr.Token)

							if err := a.str.DeleteShareLimitJournalForShare(shr.Id, trx); err == nil {
								commit = true
							} else {
								logrus.Errorf("error deleting share_limit_journal for '%v'", shr.Token)
							}
						} else {
							logrus.Infof("share '%v' still over limit", shr.Token)
						}
					} else {
						logrus.Errorf("error checking share limit for '%v': %v", shr.Token, err)
					}

				case store.LimitAction:
					if enforce, warning, err := a.checkShareLimit(shr.Token); err == nil {
						if !enforce && !warning {
							logrus.Infof("relaxing limit for share '%v'", shr.Token)

							if err := a.str.DeleteShareLimitJournalForShare(shr.Id, trx); err == nil {
								commit = true
							} else {
								logrus.Errorf("error deleting share_limit_journal for '%v': %v", shr.Token, err)
							}
						} else {
							logrus.Infof("share '%v' still over limit", shr.Token)
						}
					} else {
						logrus.Errorf("error checking share limit for '%v': %v", shr.Token, err)
					}
				}
			} else {
				logrus.Errorf("error getting share for '#%d': %v", slj.ShareId, err)
			}
		}
	} else {
		return err
	}

	if eljs, err := a.str.FindAllLatestEnvironmentLimitJournal(trx); err == nil {
		for _, elj := range eljs {
			if env, err := a.str.GetEnvironment(elj.EnvironmentId, trx); err == nil {
				switch elj.Action {
				case store.WarningAction:
					if enforce, warning, err := a.checkEnvironmentLimit(int64(elj.EnvironmentId)); err == nil {
						if !enforce && !warning {
							logrus.Infof("relaxing warning for environment '%v'", env.ZId)

							if err := a.str.DeleteEnvironmentLimitJournalForEnvironment(env.Id, trx); err == nil {
								commit = true
							} else {
								logrus.Errorf("error deleteing environment_limit_journal for '%v': %v", env.ZId, err)
							}
						} else {
							logrus.Infof("environment '%v' still over limit", env.ZId)
						}
					} else {
						logrus.Errorf("error checking environment limit for '%v': %v", env.ZId, err)
					}

				case store.LimitAction:
					if enforce, warning, err := a.checkEnvironmentLimit(int64(elj.EnvironmentId)); err == nil {
						if !enforce && !warning {
							logrus.Infof("relaxing limit for environment '%v'", env.ZId)

							if err := a.str.DeleteEnvironmentLimitJournalForEnvironment(env.Id, trx); err == nil {
								commit = true
							} else {
								logrus.Errorf("error deleteing environment_limit_journal for '%v': %v", env.ZId, err)
							}
						} else {
							logrus.Infof("environment '%v' still over limit", env.ZId)
						}
					} else {
						logrus.Errorf("error checking environment limit for '%v': %v", env.ZId, err)
					}
				}
			} else {
				logrus.Errorf("error getting environment for '#%d': %v", elj.EnvironmentId, err)
			}
		}
	} else {
		return err
	}

	if aljs, err := a.str.FindAllLatestAccountLimitJournal(trx); err == nil {
		for _, alj := range aljs {
			if acct, err := a.str.GetAccount(alj.AccountId, trx); err == nil {
				switch alj.Action {
				case store.WarningAction:
					if enforce, warning, err := a.checkAccountLimit(int64(alj.AccountId)); err == nil {
						if !enforce && !warning {
							logrus.Infof("relaxing warning for account '%v'", acct.Email)

							if err := a.str.DeleteAccountLimitJournalForAccount(acct.Id, trx); err == nil {
								commit = true
							} else {
								logrus.Errorf("error deleting account_limit_journal for '%v': %v", acct.Email, err)
							}
						} else {
							logrus.Infof("account '%v' still over limit", acct.Email)
						}
					} else {
						logrus.Errorf("error checking account limit for '%v': %v", acct.Email, err)
					}

				case store.LimitAction:
					if enforce, warning, err := a.checkAccountLimit(int64(alj.AccountId)); err == nil {
						if !enforce && !warning {
							logrus.Infof("relaxing limit for account '%v'", acct.Email)

							if err := a.str.DeleteAccountLimitJournalForAccount(acct.Id, trx); err == nil {
								commit = true
							} else {
								logrus.Errorf("error deleting account_limit_journal for '%v': %v", acct.Email, err)
							}
						} else {
							logrus.Infof("account '%v' still over limit", acct.Email)
						}
					} else {
						logrus.Errorf("error checking account limit for '%v': %v", acct.Email, err)
					}
				}
			} else {
				logrus.Errorf("error getting account for '#%d': %v", alj.AccountId, err)
			}
		}
	} else {
		return err
	}

	if commit {
		if err := trx.Commit(); err != nil {
			return err
		}
	}

	return nil
}

func (a *Agent) checkAccountLimit(acctId int64) (enforce, warning bool, err error) {
	period := 24 * time.Hour
	limit := DefaultBandwidthPerPeriod()
	if a.cfg.Bandwidth != nil && a.cfg.Bandwidth.PerAccount != nil {
		limit = a.cfg.Bandwidth.PerAccount
	}
	if limit.Period > 0 {
		period = limit.Period
	}
	rx, tx, err := a.ifx.totalRxTxForAccount(acctId, period)
	if err != nil {
		logrus.Error(err)
	}

	enforce, warning = a.checkLimit(limit, rx, tx)
	return enforce, warning, nil
}

func (a *Agent) checkEnvironmentLimit(envId int64) (enforce, warning bool, err error) {
	period := 24 * time.Hour
	limit := DefaultBandwidthPerPeriod()
	if a.cfg.Bandwidth != nil && a.cfg.Bandwidth.PerEnvironment != nil {
		limit = a.cfg.Bandwidth.PerEnvironment
	}
	if limit.Period > 0 {
		period = limit.Period
	}
	rx, tx, err := a.ifx.totalRxTxForEnvironment(envId, period)
	if err != nil {
		logrus.Error(err)
	}

	enforce, warning = a.checkLimit(limit, rx, tx)
	return enforce, warning, nil
}

func (a *Agent) checkShareLimit(shrToken string) (enforce, warning bool, err error) {
	period := 24 * time.Hour
	limit := DefaultBandwidthPerPeriod()
	if a.cfg.Bandwidth != nil && a.cfg.Bandwidth.PerShare != nil {
		limit = a.cfg.Bandwidth.PerShare
	}
	if limit.Period > 0 {
		period = limit.Period
	}
	rx, tx, err := a.ifx.totalRxTxForShare(shrToken, period)
	if err != nil {
		logrus.Error(err)
	}

	enforce, warning = a.checkLimit(limit, rx, tx)
	if enforce || warning {
		logrus.Debugf("'%v': %v", shrToken, a.describeLimit(limit, rx, tx))
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
