package limits

import (
	"github.com/jmoiron/sqlx"
	"github.com/openziti/zrok/controller/emailUi"
	"github.com/openziti/zrok/controller/metrics"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/openziti/zrok/util"
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
		warningActions: []AccountAction{newWarningAction(emailCfg, str)},
		limitActions:   []AccountAction{newLimitAction(str, zCfg)},
		relaxActions:   []AccountAction{newRelaxAction(str, zCfg)},
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

		ul, err := a.getUserLimits(acctId, trx)
		if err != nil {
			return false, err
		}

		if ul.resource.GetEnvironments() > store.Unlimited {
			envs, err := a.str.FindEnvironmentsForAccount(acctId, trx)
			if err != nil {
				return false, err
			}
			if len(envs)+1 > ul.resource.GetEnvironments() {
				return false, nil
			}
		}
	}

	return true, nil
}

func (a *Agent) CanCreateShare(acctId, envId int, reserved, uniqueName bool, _ sdk.ShareMode, backendMode sdk.BackendMode, trx *sqlx.Tx) (bool, error) {
	if a.cfg.Enforcing {
		if err := a.str.LimitCheckLock(acctId, trx); err != nil {
			return false, err
		}

		ul, err := a.getUserLimits(acctId, trx)
		if err != nil {
			return false, err
		}

		if scopedBwc, found := ul.scopes[backendMode]; found {
			latestScopedJe, err := a.isBandwidthClassLimitedForAccount(acctId, scopedBwc, trx)
			if err != nil {
				return false, err
			}
			if latestScopedJe != nil {
				return false, nil
			}
		} else {
			for _, bwc := range ul.bandwidth {
				latestJe, err := a.isBandwidthClassLimitedForAccount(acctId, bwc, trx)
				if err != nil {
					return false, err
				}
				if latestJe != nil {
					return false, nil
				}
			}
		}

		rc := ul.resource
		if scopeRc, found := ul.scopes[backendMode]; found {
			rc = scopeRc
		}
		if rc.GetShares() > store.Unlimited || (reserved && rc.GetReservedShares() > store.Unlimited) || (reserved && uniqueName && rc.GetUniqueNames() > store.Unlimited) {
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
				if rc.GetShares() > store.Unlimited && total+1 > rc.GetShares() {
					logrus.Debugf("account '#%d', environment '%d' over shares limit '%d'", acctId, envId, a.cfg.Shares)
					return false, nil
				}
				if reserved && rc.GetReservedShares() > store.Unlimited && reserveds+1 > rc.GetReservedShares() {
					logrus.Debugf("account '#%d', environment '%d' over reserved shares limit '%d'", acctId, envId, a.cfg.ReservedShares)
					return false, nil
				}
				if reserved && uniqueName && rc.GetUniqueNames() > store.Unlimited && uniqueNames+1 > rc.GetUniqueNames() {
					logrus.Debugf("account '#%d', environment '%d' over unique names limit '%d'", acctId, envId, a.cfg.UniqueNames)
					return false, nil
				}
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
		env, err := a.str.GetEnvironment(shr.EnvironmentId, trx)
		if err != nil {
			return false, err
		}
		if env.AccountId != nil {
			if err := a.str.LimitCheckLock(*env.AccountId, trx); err != nil {
				return false, err
			}

			ul, err := a.getUserLimits(*env.AccountId, trx)
			if err != nil {
				return false, err
			}

			if scopedBwc, found := ul.scopes[sdk.BackendMode(shr.BackendMode)]; found {
				latestScopedJe, err := a.isBandwidthClassLimitedForAccount(*env.AccountId, scopedBwc, trx)
				if err != nil {
					return false, err
				}
				if latestScopedJe != nil {
					return false, nil
				}
			} else {
				for _, bwc := range ul.bandwidth {
					latestJe, err := a.isBandwidthClassLimitedForAccount(*env.AccountId, bwc, trx)
					if err != nil {
						return false, err
					}
					if latestJe != nil {
						return false, nil
					}
				}
			}

			rc := ul.resource
			if scopeRc, found := ul.scopes[sdk.BackendMode(shr.BackendMode)]; found {
				rc = scopeRc
			}
			if rc.GetShareFrontends() > store.Unlimited {
				fes, err := a.str.FindFrontendsForPrivateShare(shr.Id, trx)
				if err != nil {
					return false, err
				}
				if len(fes)+1 > rc.GetShareFrontends() {
					logrus.Infof("account '#%d' over frontends per share limit '%d'", *env.AccountId, rc.GetShareFrontends())
					return false, nil
				}
			}
		} else {
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
				logrus.Debugf("not enforcing for usage with no share token: %v", usage.String())
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

	shr, err := a.str.FindShareWithTokenEvenIfDeleted(u.ShareToken, trx)
	if err != nil {
		return err
	}

	ul, err := a.getUserLimits(int(u.AccountId), trx)
	if err != nil {
		return err
	}

	exceededBwc, rxBytes, txBytes, err := a.anyBandwidthLimitExceeded(acct, u, ul.toBandwidthArray(sdk.BackendMode(shr.BackendMode)))
	if err != nil {
		return errors.Wrap(err, "error checking limit classes")
	}

	if exceededBwc != nil {
		latestJe, err := a.isBandwidthClassLimitedForAccount(int(u.AccountId), exceededBwc, trx)
		if err != nil {
			return err
		}
		if latestJe == nil {
			je := &store.BandwidthLimitJournalEntry{
				AccountId: int(u.AccountId),
				RxBytes:   rxBytes,
				TxBytes:   txBytes,
				Action:    exceededBwc.GetLimitAction(),
			}
			if !exceededBwc.IsGlobal() {
				lcId := exceededBwc.GetLimitClassId()
				je.LimitClassId = &lcId
			}
			if _, err := a.str.CreateBandwidthLimitJournalEntry(je, trx); err != nil {
				return err
			}
			acct, err := a.str.GetAccount(int(u.AccountId), trx)
			if err != nil {
				return err
			}
			switch exceededBwc.GetLimitAction() {
			case store.LimitLimitAction:
				for _, limitAction := range a.limitActions {
					if err := limitAction.HandleAccount(acct, rxBytes, txBytes, exceededBwc, ul, trx); err != nil {
						return errors.Wrapf(err, "%v", reflect.TypeOf(limitAction).String())
					}
				}

			case store.WarningLimitAction:
				for _, warningAction := range a.warningActions {
					if err := warningAction.HandleAccount(acct, rxBytes, txBytes, exceededBwc, ul, trx); err != nil {
						return errors.Wrapf(err, "%v", reflect.TypeOf(warningAction).String())
					}
				}
			}
			if err := trx.Commit(); err != nil {
				return err
			}
		} else {
			logrus.Debugf("limit '%v' already applied for '%v' (at: %v)", exceededBwc, acct.Email, latestJe.CreatedAt)
		}
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

	if bwjes, err := a.str.FindAllBandwidthLimitJournal(trx); err == nil {
		accounts := make(map[int]*store.Account)
		uls := make(map[int]*userLimits)
		accountPeriods := make(map[int]map[int]*periodBwValues)

		for _, bwje := range bwjes {
			if _, found := accounts[bwje.AccountId]; !found {
				if acct, err := a.str.GetAccount(bwje.AccountId, trx); err == nil {
					accounts[bwje.AccountId] = acct
					ul, err := a.getUserLimits(acct.Id, trx)
					if err != nil {
						return errors.Wrapf(err, "error getting user limits for '%v'", acct.Email)
					}
					uls[bwje.AccountId] = ul
					accountPeriods[bwje.AccountId] = make(map[int]*periodBwValues)
				} else {
					return err
				}
			}

			var bwc store.BandwidthClass
			if bwje.LimitClassId == nil {
				globalBwcs := newConfigBandwidthClasses(a.cfg.Bandwidth)
				if bwje.Action == store.WarningLimitAction {
					bwc = globalBwcs[0]
				} else {
					bwc = globalBwcs[1]
				}
			} else {
				lc, err := a.str.GetLimitClass(*bwje.LimitClassId, trx)
				if err != nil {
					return err
				}
				bwc = lc
			}

			if periods, accountFound := accountPeriods[bwje.AccountId]; accountFound {
				if _, periodFound := periods[bwc.GetPeriodMinutes()]; !periodFound {
					rx, tx, err := a.ifx.totalRxTxForAccount(int64(bwje.AccountId), time.Duration(bwc.GetPeriodMinutes())*time.Minute)
					if err != nil {
						return err
					}
					periods[bwc.GetPeriodMinutes()] = &periodBwValues{rx: rx, tx: tx}
					accountPeriods[bwje.AccountId] = periods
				}
			} else {
				return errors.New("accountPeriods corrupted")
			}

			used := accountPeriods[bwje.AccountId][bwc.GetPeriodMinutes()]
			if !a.transferBytesExceeded(used.rx, used.tx, bwc) {
				if bwc.GetLimitAction() == store.LimitLimitAction {
					logrus.Infof("relaxing limit '%v' for '%v'", bwc.String(), accounts[bwje.AccountId].Email)
					for _, action := range a.relaxActions {
						if err := action.HandleAccount(accounts[bwje.AccountId], used.rx, used.tx, bwc, uls[bwje.AccountId], trx); err != nil {
							return errors.Wrapf(err, "%v", reflect.TypeOf(action).String())
						}
					}
				} else {
					logrus.Infof("relaxing warning '%v' for '%v'", bwc.String(), accounts[bwje.AccountId].Email)
				}
				if bwc.IsGlobal() {
					if err := a.str.DeleteBandwidthLimitJournalEntryForGlobal(bwje.AccountId, trx); err == nil {
						commit = true
					} else {
						logrus.Errorf("error deleting global bandwidth limit journal entry for '%v': %v", accounts[bwje.AccountId].Email, err)
					}
				} else {
					if err := a.str.DeleteBandwidthLimitJournalEntryForLimitClass(bwje.AccountId, *bwje.LimitClassId, trx); err == nil {
						commit = true
					} else {
						logrus.Errorf("error deleting bandwidth limit journal entry for '%v': %v", accounts[bwje.AccountId].Email, err)
					}
				}
			} else {
				logrus.Infof("'%v' still over limit: '%v' with rx: %v, tx: %v, total: %v", accounts[bwje.AccountId].Email, bwc, util.BytesToSize(used.rx), util.BytesToSize(used.tx), util.BytesToSize(used.rx+used.tx))
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

func (a *Agent) isBandwidthClassLimitedForAccount(acctId int, bwc store.BandwidthClass, trx *sqlx.Tx) (*store.BandwidthLimitJournalEntry, error) {
	if bwc.IsGlobal() {
		if empty, err := a.str.IsBandwidthLimitJournalEmptyForGlobal(acctId, trx); err == nil && !empty {
			je, err := a.str.FindLatestBandwidthLimitJournalForGlobal(acctId, trx)
			if err != nil {
				return nil, err
			}
			if je.Action == bwc.GetLimitAction() {
				logrus.Debugf("account '#%d' over bandwidth for global bandwidth class '%v'", acctId, bwc)
				return je, nil
			}
		} else if err != nil {
			return nil, err
		}
	} else {
		if empty, err := a.str.IsBandwidthLimitJournalEmptyForLimitClass(acctId, bwc.GetLimitClassId(), trx); err == nil && !empty {
			je, err := a.str.FindLatestBandwidthLimitJournalForLimitClass(acctId, bwc.GetLimitClassId(), trx)
			if err != nil {
				return nil, err
			}
			if je.Action == bwc.GetLimitAction() {
				logrus.Debugf("account '#%d' over bandwidth for limit class '%v'", acctId, bwc)
				return je, nil
			}
		} else if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (a *Agent) anyBandwidthLimitExceeded(acct *store.Account, u *metrics.Usage, bwcs []store.BandwidthClass) (store.BandwidthClass, int64, int64, error) {
	periodBw := make(map[int]periodBwValues)

	var selectedLc store.BandwidthClass
	var rxBytes int64
	var txBytes int64

	for _, bwc := range bwcs {
		if _, found := periodBw[bwc.GetPeriodMinutes()]; !found {
			rx, tx, err := a.ifx.totalRxTxForAccount(u.AccountId, time.Minute*time.Duration(bwc.GetPeriodMinutes()))
			if err != nil {
				return nil, 0, 0, errors.Wrapf(err, "error getting rx/tx for account '%v'", acct.Email)
			}
			periodBw[bwc.GetPeriodMinutes()] = periodBwValues{rx: rx, tx: tx}
		}
		period := periodBw[bwc.GetPeriodMinutes()]

		if a.transferBytesExceeded(period.rx, period.tx, bwc) {
			selectedLc = bwc
			rxBytes = period.rx
			txBytes = period.tx
		} else {
			logrus.Debugf("'%v' limit ok '%v' with rx: %v, tx: %v, total: %v", acct.Email, bwc, util.BytesToSize(period.rx), util.BytesToSize(period.tx), util.BytesToSize(period.rx+period.tx))
		}
	}

	if selectedLc != nil {
		logrus.Infof("'%v' exceeded limit '%v' with rx: %v, tx: %v, total: %v", acct.Email, selectedLc, util.BytesToSize(rxBytes), util.BytesToSize(txBytes), util.BytesToSize(rxBytes+txBytes))
	}

	return selectedLc, rxBytes, txBytes, nil
}

func (a *Agent) transferBytesExceeded(rx, tx int64, bwc store.BandwidthClass) bool {
	if bwc.GetTxBytes() != store.Unlimited && tx >= bwc.GetTxBytes() {
		return true
	}
	if bwc.GetRxBytes() != store.Unlimited && rx >= bwc.GetRxBytes() {
		return true
	}
	if bwc.GetTotalBytes() != store.Unlimited && tx+rx >= bwc.GetTotalBytes() {
		return true
	}
	return false
}

type periodBwValues struct {
	rx int64
	tx int64
}
