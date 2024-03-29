package limits

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/openziti/zrok/controller/emailUi"
	"github.com/openziti/zrok/controller/metrics"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/openziti/zrok/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"reflect"
	"time"
)

type Agent struct {
	cfg                *Config
	ifx                *influxReader
	zCfg               *zrokEdgeSdk.Config
	str                *store.Store
	queue              chan *metrics.Usage
	acctWarningActions []AccountAction
	acctLimitActions   []AccountAction
	acctRelaxActions   []AccountAction
	envWarningActions  []EnvironmentAction
	envLimitActions    []EnvironmentAction
	envRelaxActions    []EnvironmentAction
	shrWarningActions  []ShareAction
	shrLimitActions    []ShareAction
	shrRelaxActions    []ShareAction
	close              chan struct{}
	join               chan struct{}
}

func NewAgent(cfg *Config, ifxCfg *metrics.InfluxConfig, zCfg *zrokEdgeSdk.Config, emailCfg *emailUi.Config, str *store.Store) (*Agent, error) {
	a := &Agent{
		cfg:                cfg,
		ifx:                newInfluxReader(ifxCfg),
		zCfg:               zCfg,
		str:                str,
		queue:              make(chan *metrics.Usage, 1024),
		acctWarningActions: []AccountAction{newAccountWarningAction(emailCfg, str)},
		acctLimitActions:   []AccountAction{newAccountLimitAction(str, zCfg)},
		acctRelaxActions:   []AccountAction{newAccountRelaxAction(str, zCfg)},
		envWarningActions:  []EnvironmentAction{newEnvironmentWarningAction(emailCfg, str)},
		envLimitActions:    []EnvironmentAction{newEnvironmentLimitAction(str, zCfg)},
		envRelaxActions:    []EnvironmentAction{newEnvironmentRelaxAction(str, zCfg)},
		shrWarningActions:  []ShareAction{newShareWarningAction(emailCfg, str)},
		shrLimitActions:    []ShareAction{newShareLimitAction(str, zCfg)},
		shrRelaxActions:    []ShareAction{newShareRelaxAction(str, zCfg)},
		close:              make(chan struct{}),
		join:               make(chan struct{}),
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
	if a.cfg.Enforcing {
		if empty, err := a.str.IsAccountLimitJournalEmpty(acctId, trx); err == nil && !empty {
			alj, err := a.str.FindLatestAccountLimitJournal(acctId, trx)
			if err != nil {
				return false, err
			}
			if alj.Action == store.LimitAction {
				return false, nil
			}
		} else if err != nil {
			return false, err
		}

		if a.cfg.Environments > Unlimited {
			envs, err := a.str.FindEnvironmentsForAccount(acctId, trx)
			if err != nil {
				return false, err
			}
			if len(envs)+1 > a.cfg.Environments {
				return false, nil
			}
		}
	}
	return true, nil
}

func (a *Agent) CanCreateShare(acctId, envId int, trx *sqlx.Tx) (bool, error) {
	if a.cfg.Enforcing {
		if empty, err := a.str.IsAccountLimitJournalEmpty(acctId, trx); err == nil && !empty {
			alj, err := a.str.FindLatestAccountLimitJournal(acctId, trx)
			if err != nil {
				return false, err
			}
			if alj.Action == store.LimitAction {
				return false, nil
			}
		} else if err != nil {
			return false, err
		}

		if empty, err := a.str.IsEnvironmentLimitJournalEmpty(envId, trx); err == nil && !empty {
			elj, err := a.str.FindLatestEnvironmentLimitJournal(envId, trx)
			if err != nil {
				return false, err
			}
			if elj.Action == store.LimitAction {
				return false, nil
			}
		} else if err != nil {
			return false, err
		}

		if a.cfg.Shares > Unlimited {
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
	}
	return true, nil
}

func (a *Agent) CanAccessShare(shrId int, trx *sqlx.Tx) (bool, error) {
	if a.cfg.Enforcing {
		shr, err := a.str.GetShare(shrId, trx)
		if err != nil {
			return false, err
		}
		if empty, err := a.str.IsShareLimitJournalEmpty(shr.Id, trx); err == nil && !empty {
			slj, err := a.str.FindLatestShareLimitJournal(shr.Id, trx)
			if err != nil {
				return false, err
			}
			if slj.Action == store.LimitAction {
				return false, nil
			}
		} else if err != nil {
			return false, err
		}

		env, err := a.str.GetEnvironment(shr.EnvironmentId, trx)
		if err != nil {
			return false, err
		}
		if empty, err := a.str.IsEnvironmentLimitJournalEmpty(env.Id, trx); err == nil && !empty {
			elj, err := a.str.FindLatestEnvironmentLimitJournal(env.Id, trx)
			if err != nil {
				return false, err
			}
			if elj.Action == store.LimitAction {
				return false, nil
			}
		} else if err != nil {
			return false, err
		}

		if env.AccountId != nil {
			acct, err := a.str.GetAccount(*env.AccountId, trx)
			if err != nil {
				return false, err
			}
			if empty, err := a.str.IsAccountLimitJournalEmpty(acct.Id, trx); err == nil && !empty {
				alj, err := a.str.FindLatestAccountLimitJournal(acct.Id, trx)
				if err != nil {
					return false, err
				}
				if alj.Action == store.LimitAction {
					return false, nil
				}
			} else if err != nil {
				return false, err
			}
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

	lastCycle := time.Now()
mainLoop:
	for {
		select {
		case usage := <-a.queue:
			if usage.ShareToken != "" {
				if err := a.enforce(usage); err != nil {
					logrus.Errorf("error running enforcement: %v", err)
				}
				if time.Since(lastCycle) > a.cfg.Cycle {
					if err := a.relax(); err != nil {
						logrus.Errorf("error running relax cycle: %v", err)
					}
					lastCycle = time.Now()
				}
			} else {
				logrus.Warnf("not enforcing for usage with no share token: %v", usage.String())
			}

		case <-time.After(a.cfg.Cycle):
			if err := a.relax(); err != nil {
				logrus.Errorf("error running relax cycle: %v", err)
			}
			lastCycle = time.Now()

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

	acct, err := a.str.GetAccount(int(u.AccountId), trx)
	if err != nil {
		return err
	}
	if acct.Limitless {
		return nil
	}

	if enforce, warning, rxBytes, txBytes, err := a.checkAccountLimit(u.AccountId); err == nil {
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
					RxBytes:   rxBytes,
					TxBytes:   txBytes,
					Action:    store.LimitAction,
				}, trx)
				if err != nil {
					return err
				}
				acct, err := a.str.GetAccount(int(u.AccountId), trx)
				if err != nil {
					return err
				}
				// run account limit actions
				for _, action := range a.acctLimitActions {
					if err := action.HandleAccount(acct, rxBytes, txBytes, a.cfg.Bandwidth.PerAccount, trx); err != nil {
						return errors.Wrapf(err, "%v", reflect.TypeOf(action).String())
					}
				}
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
					RxBytes:   rxBytes,
					TxBytes:   txBytes,
					Action:    store.WarningAction,
				}, trx)
				if err != nil {
					return err
				}
				acct, err := a.str.GetAccount(int(u.AccountId), trx)
				if err != nil {
					return err
				}
				// run account warning actions
				for _, action := range a.acctWarningActions {
					if err := action.HandleAccount(acct, rxBytes, txBytes, a.cfg.Bandwidth.PerAccount, trx); err != nil {
						return errors.Wrapf(err, "%v", reflect.TypeOf(action).String())
					}
				}
				if err := trx.Commit(); err != nil {
					return err
				}
			} else {
				logrus.Debugf("already warned account '#%d' at %v", u.AccountId, warnedAt)
			}

		} else {
			if enforce, warning, rxBytes, txBytes, err := a.checkEnvironmentLimit(u.EnvironmentId); err == nil {
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
							RxBytes:       rxBytes,
							TxBytes:       txBytes,
							Action:        store.LimitAction,
						}, trx)
						if err != nil {
							return err
						}
						env, err := a.str.GetEnvironment(int(u.EnvironmentId), trx)
						if err != nil {
							return err
						}
						// run environment limit actions
						for _, action := range a.envLimitActions {
							if err := action.HandleEnvironment(env, rxBytes, txBytes, a.cfg.Bandwidth.PerEnvironment, trx); err != nil {
								return errors.Wrapf(err, "%v", reflect.TypeOf(action).String())
							}
						}
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
							RxBytes:       rxBytes,
							TxBytes:       txBytes,
							Action:        store.WarningAction,
						}, trx)
						if err != nil {
							return err
						}
						env, err := a.str.GetEnvironment(int(u.EnvironmentId), trx)
						if err != nil {
							return err
						}
						// run environment warning actions
						for _, action := range a.envWarningActions {
							if err := action.HandleEnvironment(env, rxBytes, txBytes, a.cfg.Bandwidth.PerEnvironment, trx); err != nil {
								return errors.Wrapf(err, "%v", reflect.TypeOf(action).String())
							}
						}
						if err := trx.Commit(); err != nil {
							return err
						}
					} else {
						logrus.Debugf("already warned environment '#%d' at %v", u.EnvironmentId, warnedAt)
					}

				} else {
					if enforce, warning, rxBytes, txBytes, err := a.checkShareLimit(u.ShareToken); err == nil {
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
									RxBytes: rxBytes,
									TxBytes: txBytes,
									Action:  store.LimitAction,
								}, trx)
								if err != nil {
									return err
								}
								// run share limit actions
								for _, action := range a.shrLimitActions {
									if err := action.HandleShare(shr, rxBytes, txBytes, a.cfg.Bandwidth.PerShare, trx); err != nil {
										return errors.Wrapf(err, "%v", reflect.TypeOf(action).String())
									}
								}
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
									RxBytes: rxBytes,
									TxBytes: txBytes,
									Action:  store.WarningAction,
								}, trx)
								if err != nil {
									return err
								}
								// run share warning actions
								for _, action := range a.shrWarningActions {
									if err := action.HandleShare(shr, rxBytes, txBytes, a.cfg.Bandwidth.PerShare, trx); err != nil {
										return errors.Wrapf(err, "%v", reflect.TypeOf(action).String())
									}
								}
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
	logrus.Debug("relaxing")

	trx, err := a.str.Begin()
	if err != nil {
		return errors.Wrap(err, "error starting transaction")
	}
	defer func() { _ = trx.Rollback() }()

	commit := false

	if sljs, err := a.str.FindAllLatestShareLimitJournal(trx); err == nil {
		for _, slj := range sljs {
			if shr, err := a.str.GetShare(slj.ShareId, trx); err == nil {
				if slj.Action == store.WarningAction || slj.Action == store.LimitAction {
					if enforce, warning, rxBytes, txBytes, err := a.checkShareLimit(shr.Token); err == nil {
						if !enforce && !warning {
							if slj.Action == store.LimitAction {
								// run relax actions for share
								for _, action := range a.shrRelaxActions {
									if err := action.HandleShare(shr, rxBytes, txBytes, a.cfg.Bandwidth.PerShare, trx); err != nil {
										return errors.Wrapf(err, "%v", reflect.TypeOf(action).String())
									}
								}
							} else {
								logrus.Infof("relaxing warning for '%v'", shr.Token)
							}
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
				if elj.Action == store.WarningAction || elj.Action == store.LimitAction {
					if enforce, warning, rxBytes, txBytes, err := a.checkEnvironmentLimit(int64(elj.EnvironmentId)); err == nil {
						if !enforce && !warning {
							if elj.Action == store.LimitAction {
								// run relax actions for environment
								for _, action := range a.envRelaxActions {
									if err := action.HandleEnvironment(env, rxBytes, txBytes, a.cfg.Bandwidth.PerEnvironment, trx); err != nil {
										return errors.Wrapf(err, "%v", reflect.TypeOf(action).String())
									}
								}
							} else {
								logrus.Infof("relaxing warning for '%v'", env.ZId)
							}
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
				if alj.Action == store.WarningAction || alj.Action == store.LimitAction {
					if enforce, warning, rxBytes, txBytes, err := a.checkAccountLimit(int64(alj.AccountId)); err == nil {
						if !enforce && !warning {
							if alj.Action == store.LimitAction {
								// run relax actions for account
								for _, action := range a.acctRelaxActions {
									if err := action.HandleAccount(acct, rxBytes, txBytes, a.cfg.Bandwidth.PerAccount, trx); err != nil {
										return errors.Wrapf(err, "%v", reflect.TypeOf(action).String())
									}
								}
							} else {
								logrus.Infof("relaxing warning for '%v'", acct.Email)
							}
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

func (a *Agent) checkAccountLimit(acctId int64) (enforce, warning bool, rxBytes, txBytes int64, err error) {
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
	return enforce, warning, rx, tx, nil
}

func (a *Agent) checkEnvironmentLimit(envId int64) (enforce, warning bool, rxBytes, txBytes int64, err error) {
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
	return enforce, warning, rx, tx, nil
}

func (a *Agent) checkShareLimit(shrToken string) (enforce, warning bool, rxBytes, txBytes int64, err error) {
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
		logrus.Debugf("'%v': %v", shrToken, describeLimit(limit, rx, tx))
	}

	return enforce, warning, rx, tx, nil
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

func describeLimit(cfg *BandwidthPerPeriod, rx, tx int64) string {
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
