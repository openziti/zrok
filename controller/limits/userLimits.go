package limits

import (
	"github.com/jmoiron/sqlx"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type userLimits struct {
	resource  store.ResourceCountClass
	bandwidth []store.BandwidthClass
	scopes    map[sdk.BackendMode]store.BandwidthClass
}

func (ul *userLimits) toBandwidthArray(backendMode sdk.BackendMode) []store.BandwidthClass {
	if scopedBwc, found := ul.scopes[backendMode]; found {
		out := make([]store.BandwidthClass, 0)
		for _, bwc := range ul.bandwidth {
			out = append(out, bwc)
		}
		out = append(out, scopedBwc)
		return out
	}
	return ul.bandwidth
}

func (a *Agent) getUserLimits(acctId int, trx *sqlx.Tx) (*userLimits, error) {
	resource := newConfigResourceCountClass(a.cfg)
	cfgBwcs := newConfigBandwidthClasses(a.cfg.Bandwidth)
	bwWarning := cfgBwcs[0]
	bwLimit := cfgBwcs[1]
	scopes := map[sdk.BackendMode]store.BandwidthClass{}

	alcs, err := a.str.FindAppliedLimitClassesForAccount(acctId, trx)
	if err != nil {
		return nil, errors.Wrapf(err, "error finding applied limit classes for account '%d'", acctId)
	}
	for _, alc := range alcs {
		if a.isResourceCountClass(alc) {
			resource = alc
		} else if a.isUnscopedBandwidthClass(alc) {
			if alc.LimitAction == store.WarningLimitAction {
				bwWarning = alc
			} else {
				bwLimit = alc
			}
		} else if a.isScopedBandwidthClass(alc) {
			scopes[*alc.BackendMode] = alc
		} else {
			logrus.Warnf("unknown type of limit class '%v'", alc)
		}
	}

	userLimits := &userLimits{
		resource:  resource,
		bandwidth: []store.BandwidthClass{bwWarning, bwLimit},
		scopes:    scopes,
	}

	return userLimits, nil
}

func (a *Agent) isResourceCountClass(alc *store.LimitClass) bool {
	if alc.BackendMode != nil {
		return false
	}
	if alc.Environments == store.Unlimited && alc.Shares == store.Unlimited && alc.ReservedShares == store.Unlimited && alc.UniqueNames == store.Unlimited {
		return false
	}
	return true
}

func (a *Agent) isUnscopedBandwidthClass(alc *store.LimitClass) bool {
	if alc.BackendMode != nil {
		return false
	}
	if alc.Environments > store.Unlimited || alc.Shares > store.Unlimited || alc.ReservedShares > store.Unlimited || alc.UniqueNames > store.Unlimited {
		return false
	}
	if alc.PeriodMinutes < 1 {
		return false
	}
	if alc.RxBytes == store.Unlimited && alc.TxBytes == store.Unlimited && alc.TotalBytes == store.Unlimited {
		return false
	}
	return true
}

func (a *Agent) isScopedBandwidthClass(alc *store.LimitClass) bool {
	if alc.BackendMode == nil {
		return false
	}
	if alc.Environments > store.Unlimited || alc.Shares > store.Unlimited || alc.ReservedShares > store.Unlimited || alc.UniqueNames > store.Unlimited {
		return false
	}
	if alc.PeriodMinutes < 1 {
		return false
	}
	if alc.RxBytes == store.Unlimited && alc.TxBytes == store.Unlimited && alc.TotalBytes == store.Unlimited {
		return false
	}
	return true
}
