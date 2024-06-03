package limits

import (
	"github.com/jmoiron/sqlx"
	"github.com/openziti/zrok/controller/emailUi"
	"github.com/openziti/zrok/controller/metrics"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"reflect"
	"time"
)

type Agent struct {
	cfg            *Config
	ifx            *influxReader
	zCfg           *zrokEdgeSdk.Config
	str            *store.Store
	queue          chan *metrics.Usage
	warningActions []AccountAction
	limitActions   []AccountAction
	relaxActions   []AccountAction
	close          chan struct{}
	join           chan struct{}
}

func NewAgent(cfg *Config, ifxCfg *metrics.InfluxConfig, zCfg *zrokEdgeSdk.Config, emailCfg *emailUi.Config, str *store.Store) (*Agent, error) {
	a := &Agent{
		cfg:            cfg,
		ifx:            newInfluxReader(ifxCfg),
		zCfg:           zCfg,
		str:            str,
		queue:          make(chan *metrics.Usage, 1024),
		warningActions: []AccountAction{newAccountWarningAction(emailCfg, str)},
		limitActions:   []AccountAction{newAccountLimitAction(str, zCfg)},
		relaxActions:   []AccountAction{newAccountRelaxAction(str, zCfg)},
		close:          make(chan struct{}),
		join:           make(chan struct{}),
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
		if err := a.str.LimitCheckLock(acctId, trx); err != nil {
			return false, err
		}
		if empty, err := a.str.IsBandwidthLimitJournalEmpty(acctId, trx); err == nil && !empty {
			alj, err := a.str.FindLatestBandwidthLimitJournal(acctId, trx)
			if err != nil {
				return false, err
			}
			if alj.Action == store.LimitLimitAction {
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

func (a *Agent) CanCreateShare(acctId, envId int, reserved, uniqueName bool, _ sdk.ShareMode, _ sdk.BackendMode, trx *sqlx.Tx) (bool, error) {
	if a.cfg.Enforcing {
		if err := a.str.LimitCheckLock(acctId, trx); err != nil {
			return false, err
		}
		if empty, err := a.str.IsBandwidthLimitJournalEmpty(acctId, trx); err == nil && !empty {
			alj, err := a.str.FindLatestBandwidthLimitJournal(acctId, trx)
			if err != nil {
				return false, err
			}
			if alj.Action == store.LimitLimitAction {
				return false, nil
			}
		} else if err != nil {
			return false, err
		}

		alc, err := a.str.FindLimitClassesForAccount(acctId, trx)
		if err != nil {
			logrus.Errorf("error finding limit classes for account with id '%d': %v", acctId, err)
			return false, err
		}
		sortLimitClasses(alc)
		if len(alc) > 0 {
			logrus.Infof("selected limit class: %v", alc[0])
		}

		if a.cfg.Shares > Unlimited || (reserved && a.cfg.ReservedShares > Unlimited) || (reserved && uniqueName && a.cfg.UniqueNames > Unlimited) {
			envs, err := a.str.FindEnvironmentsForAccount(acctId, trx)
			if err != nil {
				return false, err
			}
			total := 0
			reserveds := 0
			uniqueNames := 0
			for i := range envs {
				shrs, err := a.str.FindSharesForEnvironment(envs[i].Id, trx)
				if err != nil {
					return false, errors.Wrapf(err, "unable to find shares for environment '%v'", envs[i].ZId)
				}
				total += len(shrs)
				for _, shr := range shrs {
					if shr.Reserved {
						reserveds++
					}
					if shr.UniqueName {
						uniqueNames++
					}
				}
				if total+1 > a.cfg.Shares {
					logrus.Debugf("account '%d', environment '%d' over shares limit '%d'", acctId, envId, a.cfg.Shares)
					return false, nil
				}
				if reserved && reserveds+1 > a.cfg.ReservedShares {
					logrus.Debugf("account '%v', environment '%d' over reserved shares limit '%d'", acctId, envId, a.cfg.ReservedShares)
					return false, nil
				}
				if reserved && uniqueName && uniqueNames+1 > a.cfg.UniqueNames {
					logrus.Debugf("account '%v', environment '%d' over unique names limit '%d'", acctId, envId, a.cfg.UniqueNames)
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
		if empty, err := a.str.IsBandwidthLimitJournalEmpty(shr.Id, trx); err == nil && !empty {
			slj, err := a.str.FindLatestBandwidthLimitJournal(shr.Id, trx)
			if err != nil {
				return false, err
			}
			if slj.Action == store.LimitLimitAction {
				return false, nil
			}
		} else if err != nil {
			return false, err
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

	shr, err := a.str.FindShareWithToken(u.ShareToken, trx)
	if err != nil {
		return err
	}
	logrus.Infof("share: '%v', shareMode: '%v', backendMode: '%v'", shr.Token, shr.ShareMode, shr.BackendMode)

	if enforce, warning, rxBytes, txBytes, err := a.checkBandwidthLimit(u.AccountId); err == nil {
		if enforce {
			enforced := false
			var enforcedAt time.Time
			if empty, err := a.str.IsBandwidthLimitJournalEmpty(int(u.AccountId), trx); err == nil && !empty {
				if latest, err := a.str.FindLatestBandwidthLimitJournal(int(u.AccountId), trx); err == nil {
					enforced = latest.Action == store.LimitLimitAction
					enforcedAt = latest.UpdatedAt
				}
			}

			if !enforced {
				_, err := a.str.CreateBandwidthLimitJournalEntry(&store.BandwidthLimitJournalEntry{
					AccountId: int(u.AccountId),
					RxBytes:   rxBytes,
					TxBytes:   txBytes,
					Action:    store.LimitLimitAction,
				}, trx)
				if err != nil {
					return err
				}
				acct, err := a.str.GetAccount(int(u.AccountId), trx)
				if err != nil {
					return err
				}
				for _, action := range a.limitActions {
					if err := action.HandleAccount(acct, rxBytes, txBytes, a.cfg.Bandwidth, trx); err != nil {
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
			if empty, err := a.str.IsBandwidthLimitJournalEmpty(int(u.AccountId), trx); err == nil && !empty {
				if latest, err := a.str.FindLatestBandwidthLimitJournal(int(u.AccountId), trx); err == nil {
					warned = latest.Action == store.WarningLimitAction || latest.Action == store.LimitLimitAction
					warnedAt = latest.UpdatedAt
				}
			}

			if !warned {
				_, err := a.str.CreateBandwidthLimitJournalEntry(&store.BandwidthLimitJournalEntry{
					AccountId: int(u.AccountId),
					RxBytes:   rxBytes,
					TxBytes:   txBytes,
					Action:    store.WarningLimitAction,
				}, trx)
				if err != nil {
					return err
				}
				acct, err := a.str.GetAccount(int(u.AccountId), trx)
				if err != nil {
					return err
				}
				for _, action := range a.warningActions {
					if err := action.HandleAccount(acct, rxBytes, txBytes, a.cfg.Bandwidth, trx); err != nil {
						return errors.Wrapf(err, "%v", reflect.TypeOf(action).String())
					}
				}
				if err := trx.Commit(); err != nil {
					return err
				}
			} else {
				logrus.Debugf("already warned account '#%d' at %v", u.AccountId, warnedAt)
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

	if aljs, err := a.str.FindAllLatestBandwidthLimitJournal(trx); err == nil {
		for _, alj := range aljs {
			if acct, err := a.str.GetAccount(alj.AccountId, trx); err == nil {
				if alj.Action == store.WarningLimitAction || alj.Action == store.LimitLimitAction {
					if enforce, warning, rxBytes, txBytes, err := a.checkBandwidthLimit(int64(alj.AccountId)); err == nil {
						if !enforce && !warning {
							if alj.Action == store.LimitLimitAction {
								// run relax actions for account
								for _, action := range a.relaxActions {
									if err := action.HandleAccount(acct, rxBytes, txBytes, a.cfg.Bandwidth, trx); err != nil {
										return errors.Wrapf(err, "%v", reflect.TypeOf(action).String())
									}
								}
							} else {
								logrus.Infof("relaxing warning for '%v'", acct.Email)
							}
							if err := a.str.DeleteBandwidthLimitJournal(acct.Id, trx); err == nil {
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

func (a *Agent) checkBandwidthLimit(acctId int64) (enforce, warning bool, rxBytes, txBytes int64, err error) {
	period := 24 * time.Hour
	limit := DefaultBandwidthPerPeriod()
	if a.cfg.Bandwidth != nil {
		limit = a.cfg.Bandwidth
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
