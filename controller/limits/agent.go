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

		alcs, err := a.str.FindAppliedLimitClassesForAccount(acctId, trx)
		if err != nil {
			return false, err
		}
		maxEnvironments := a.cfg.Environments
		var lcId *int
		for _, alc := range alcs {
			if alc.ShareMode == nil && alc.BackendMode == nil && alc.Environments > maxEnvironments {
				maxEnvironments = alc.Environments
				lcId = &alc.Id
			}
		}

		if lcId == nil {
			if empty, err := a.str.IsBandwidthLimitJournalEmptyForGlobal(acctId, trx); err == nil && !empty {
				lj, err := a.str.FindLatestBandwidthLimitJournalForGlobal(acctId, trx)
				if err != nil {
					return false, err
				}
				if lj.Action == store.LimitLimitAction {
					return false, nil
				}
			}
		} else {
			if empty, err := a.str.IsBandwidthLimitJournalEmptyForLimitClass(acctId, *lcId, trx); err == nil && !empty {
				lj, err := a.str.FindLatestBandwidthLimitJournalForLimitClass(acctId, *lcId, trx)
				if err != nil {
					return false, err
				}
				if lj.Action == store.LimitLimitAction {
					return false, nil
				}
			}
		}

		if maxEnvironments > store.Unlimited {
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

func (a *Agent) CanCreateShare(acctId, envId int, reserved, uniqueName bool, shareMode sdk.ShareMode, backendMode sdk.BackendMode, trx *sqlx.Tx) (bool, error) {
	if a.cfg.Enforcing {
		if err := a.str.LimitCheckLock(acctId, trx); err != nil {
			return false, err
		}

		alcs, err := a.str.FindAppliedLimitClassesForAccount(acctId, trx)
		if err != nil {
			return false, err
		}
		maxShares := a.cfg.Shares
		maxReservedShares := a.cfg.ReservedShares
		maxUniqueNames := a.cfg.UniqueNames
		var lcId *int
		var points = -1
		for _, alc := range alcs {
			if a.bandwidthClassPoints(alc) > points {
				if alc.Shares >= maxShares || alc.ReservedShares >= maxReservedShares || alc.UniqueNames >= maxUniqueNames {
					maxShares = alc.Shares
					maxReservedShares = alc.ReservedShares
					maxUniqueNames = alc.UniqueNames
					lcId = &alc.Id
					points = a.bandwidthClassPoints(alc)
				}
			}
		}

		if lcId == nil {
			if empty, err := a.str.IsBandwidthLimitJournalEmptyForGlobal(acctId, trx); err == nil && !empty {
				lj, err := a.str.FindLatestBandwidthLimitJournalForGlobal(acctId, trx)
				if err != nil {
					return false, err
				}
				if lj.Action == store.LimitLimitAction {
					return false, nil
				}
			}
		} else {
			if empty, err := a.str.IsBandwidthLimitJournalEmptyForLimitClass(acctId, *lcId, trx); err == nil && !empty {
				lj, err := a.str.FindLatestBandwidthLimitJournalForLimitClass(acctId, *lcId, trx)
				if err != nil {
					return false, err
				}
				if lj.Action == store.LimitLimitAction {
					return false, nil
				}
			}
		}

		if maxShares > store.Unlimited || (reserved && maxReservedShares > store.Unlimited) || (reserved && uniqueName && maxUniqueNames > store.Unlimited) {
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
		env, err := a.str.GetEnvironment(shr.EnvironmentId, trx)
		if err != nil {
			return false, err
		}
		if env.AccountId != nil {
			alcs, err := a.str.FindAppliedLimitClassesForAccount(*env.AccountId, trx)
			if err != nil {
				return false, err
			}
			maxShares := a.cfg.Shares
			maxReservedShares := a.cfg.ReservedShares
			maxUniqueNames := a.cfg.UniqueNames
			var lcId *int
			var points = -1
			for _, alc := range alcs {
				if a.bandwidthClassPoints(alc) > points {
					if alc.Shares >= maxShares || alc.ReservedShares >= maxReservedShares || alc.UniqueNames >= maxUniqueNames {
						maxShares = alc.Shares
						maxReservedShares = alc.ReservedShares
						maxUniqueNames = alc.UniqueNames
						lcId = &alc.Id
						points = a.bandwidthClassPoints(alc)
					}
				}
			}

			if lcId == nil {
				if empty, err := a.str.IsBandwidthLimitJournalEmptyForGlobal(*env.AccountId, trx); err == nil && !empty {
					lj, err := a.str.FindLatestBandwidthLimitJournalForGlobal(*env.AccountId, trx)
					if err != nil {
						return false, err
					}
					if lj.Action == store.LimitLimitAction {
						return false, nil
					}
				}
			} else {
				if empty, err := a.str.IsBandwidthLimitJournalEmptyForLimitClass(*env.AccountId, *lcId, trx); err == nil && !empty {
					lj, err := a.str.FindLatestBandwidthLimitJournalForLimitClass(*env.AccountId, *lcId, trx)
					if err != nil {
						return false, err
					}
					if lj.Action == store.LimitLimitAction {
						return false, nil
					}
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

	//shr, err := a.str.FindShareWithTokenEvenIfDeleted(u.ShareToken, trx)
	//if err != nil {
	//	return err
	//}

	ul, err := a.getUserLimits(int(u.AccountId), trx)
	if err != nil {
		return err
	}

	exceededLc, rxBytes, txBytes, err := a.isOverLimitClass(u, ul.bandwidth)
	if err != nil {
		return errors.Wrap(err, "error checking limit classes")
	}

	if exceededLc != nil {
		enforced := false
		var enforcedAt time.Time

		if exceededLc.IsGlobal() {
			if empty, err := a.str.IsBandwidthLimitJournalEmptyForGlobal(int(u.AccountId), trx); err == nil && !empty {
				if latest, err := a.str.FindLatestBandwidthLimitJournalForGlobal(int(u.AccountId), trx); err == nil {
					enforced = latest.Action == exceededLc.GetLimitAction()
					enforcedAt = latest.UpdatedAt
					logrus.Debugf("limit '%v' already applied (enforced: %t)", exceededLc, enforced)
				} else {
					logrus.Errorf("error getting latest global bandwidth journal entry: %v", err)
				}
			} else {
				logrus.Debugf("no bandwidth limit journal entry for '%v'", exceededLc)
			}
		} else {
			if empty, err := a.str.IsBandwidthLimitJournalEmptyForLimitClass(int(u.AccountId), exceededLc.GetLimitClassId(), trx); err == nil && !empty {
				if latest, err := a.str.FindLatestBandwidthLimitJournalForLimitClass(int(u.AccountId), exceededLc.GetLimitClassId(), trx); err == nil {
					enforced = latest.Action == exceededLc.GetLimitAction()
					enforcedAt = latest.UpdatedAt
					logrus.Debugf("limit '%v' already applied (enforced: %t)", exceededLc, enforced)
				} else {
					logrus.Errorf("error getting latest bandwidth limit journal entry for limit class '%d': %v", exceededLc.GetLimitClassId(), err)
				}
			} else {
				logrus.Debugf("no bandwidth limit journal entry for '%v'", exceededLc)
			}
		}

		if !enforced {
			je := &store.BandwidthLimitJournalEntry{
				AccountId: int(u.AccountId),
				RxBytes:   rxBytes,
				TxBytes:   txBytes,
				Action:    exceededLc.GetLimitAction(),
			}
			if !exceededLc.IsGlobal() {
				lcId := exceededLc.GetLimitClassId()
				je.LimitClassId = &lcId
			}
			_, err := a.str.CreateBandwidthLimitJournalEntry(je, trx)

			if err != nil {
				return err
			}
			acct, err := a.str.GetAccount(int(u.AccountId), trx)
			if err != nil {
				return err
			}
			switch exceededLc.GetLimitAction() {
			case store.LimitLimitAction:
				for _, limitAction := range a.limitActions {
					if err := limitAction.HandleAccount(acct, rxBytes, txBytes, exceededLc, trx); err != nil {
						return errors.Wrapf(err, "%v", reflect.TypeOf(limitAction).String())
					}
				}

			case store.WarningLimitAction:
				for _, warningAction := range a.warningActions {
					if err := warningAction.HandleAccount(acct, rxBytes, txBytes, exceededLc, trx); err != nil {
						return errors.Wrapf(err, "%v", reflect.TypeOf(warningAction).String())
					}
				}
			}
			if err := trx.Commit(); err != nil {
				return err
			}
		} else {
			logrus.Debugf("already enforced limit for account '%d' at %v", u.AccountId, enforcedAt)
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
		periodBw := make(map[int]struct {
			rx int64
			tx int64
		})

		accounts := make(map[int]*store.Account)

		for _, bwje := range bwjes {
			if _, found := accounts[bwje.AccountId]; !found {
				if acct, err := a.str.GetAccount(bwje.AccountId, trx); err == nil {
					accounts[bwje.AccountId] = acct
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

			if _, found := periodBw[bwc.GetPeriodMinutes()]; !found {
				rx, tx, err := a.ifx.totalRxTxForAccount(int64(bwje.AccountId), time.Duration(bwc.GetPeriodMinutes())*time.Minute)
				if err != nil {
					return err
				}
				periodBw[bwc.GetPeriodMinutes()] = struct {
					rx int64
					tx int64
				}{
					rx: rx,
					tx: tx,
				}
			}

			used := periodBw[bwc.GetPeriodMinutes()]
			if !a.limitExceeded(used.rx, used.tx, bwc) {
				if bwc.GetLimitAction() == store.LimitLimitAction {
					logrus.Infof("relaxing limit '%v' for '%v'", bwc.String(), accounts[bwje.AccountId].Email)
					for _, action := range a.relaxActions {
						if err := action.HandleAccount(accounts[bwje.AccountId], used.rx, used.tx, bwc, trx); err != nil {
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
				logrus.Infof("account '%v' still over limit: '%v' with rx: %d, tx: %d", accounts[bwje.AccountId].Email, bwc, used.rx, used.tx)
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

func (a *Agent) isOverLimitClass(u *metrics.Usage, bwcs []store.BandwidthClass) (store.BandwidthClass, int64, int64, error) {
	periodBw := make(map[int]struct {
		rx int64
		tx int64
	})

	var selectedLc store.BandwidthClass
	var rxBytes int64
	var txBytes int64

	for _, bwc := range bwcs {
		if _, found := periodBw[bwc.GetPeriodMinutes()]; !found {
			rx, tx, err := a.ifx.totalRxTxForAccount(u.AccountId, time.Minute*time.Duration(bwc.GetPeriodMinutes()))
			if err != nil {
				return nil, 0, 0, errors.Wrapf(err, "error getting rx/tx for account '%d'", u.AccountId)
			}
			periodBw[bwc.GetPeriodMinutes()] = struct {
				rx int64
				tx int64
			}{
				rx: rx,
				tx: tx,
			}
		}
		period := periodBw[bwc.GetPeriodMinutes()]

		if a.limitExceeded(period.rx, period.tx, bwc) {
			selectedLc = bwc
			rxBytes = period.rx
			txBytes = period.tx
		} else {
			logrus.Debugf("limit ok '%v' with rx: %d, tx: %d, total: %d", bwc, period.rx, period.tx, period.rx+period.tx)
		}
	}

	if selectedLc != nil {
		logrus.Infof("exceeded limit '%v' with rx: %d, tx: %d, total: %d", selectedLc, rxBytes, txBytes, rxBytes+txBytes)
	}

	return selectedLc, rxBytes, txBytes, nil
}

func (a *Agent) bandwidthClassPoints(bwc store.BandwidthClass) int {
	points := 0
	if !bwc.IsGlobal() {
		points++
	}
	if bwc.GetLimitAction() == store.WarningLimitAction {
		points++
	}
	if bwc.GetLimitAction() == store.LimitLimitAction {
		points += 2
	}
	if bwc.GetBackendMode() != "" {
		points += 10
	}
	return points
}

func (a *Agent) limitExceeded(rx, tx int64, bwc store.BandwidthClass) bool {
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
