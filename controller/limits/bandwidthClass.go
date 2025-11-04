package limits

import (
	"fmt"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/openziti/zrok/util"
)

type configBandwidthClass struct {
	periodInMinutes int
	bw              *Bandwidth
	limitAction     store.LimitAction
}

func newConfigBandwidthClasses(cfg *BandwidthPerPeriod) []store.BandwidthClass {
	return []store.BandwidthClass{
		&configBandwidthClass{
			periodInMinutes: int(cfg.Period.Minutes()),
			bw:              cfg.Warning,
			limitAction:     store.WarningLimitAction,
		},
		&configBandwidthClass{
			periodInMinutes: int(cfg.Period.Minutes()),
			bw:              cfg.Limit,
			limitAction:     store.LimitLimitAction,
		},
	}
}

func (bc *configBandwidthClass) IsGlobal() bool {
	return true
}

func (bc *configBandwidthClass) IsScoped() bool {
	return false
}

func (bc *configBandwidthClass) GetLimitClassId() int {
	return -1
}

func (bc *configBandwidthClass) GetBackendMode() sdk.BackendMode {
	return ""
}

func (bc *configBandwidthClass) GetPeriodMinutes() int {
	return bc.periodInMinutes
}

func (bc *configBandwidthClass) GetRxBytes() int64 {
	return bc.bw.Rx
}

func (bc *configBandwidthClass) GetTxBytes() int64 {
	return bc.bw.Tx
}

func (bc *configBandwidthClass) GetTotalBytes() int64 {
	return bc.bw.Total
}

func (bc *configBandwidthClass) GetLimitAction() store.LimitAction {
	return bc.limitAction
}

func (bc *configBandwidthClass) String() string {
	out := fmt.Sprintf("ConfigClass<periodMinutes: %d", bc.periodInMinutes)
	if bc.bw.Rx > store.Unlimited {
		out += fmt.Sprintf(", rxBytes: %v", util.BytesToSize(bc.bw.Rx))
	}
	if bc.bw.Tx > store.Unlimited {
		out += fmt.Sprintf(", txBytes: %v", util.BytesToSize(bc.bw.Tx))
	}
	if bc.bw.Total > store.Unlimited {
		out += fmt.Sprintf(", totalBytes: %v", util.BytesToSize(bc.bw.Total))
	}
	out += fmt.Sprintf(", limitAction: %s>", bc.limitAction)
	return out
}
