package limits

import (
	"github.com/jmoiron/sqlx"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/sdk/golang/sdk"
)

type userLimits struct {
	resource  store.ResourceCountClass
	bandwidth store.BandwidthClass
	scopes    map[sdk.BackendMode]store.BandwidthClass
}

func (a *Agent) getUserLimits(acctId int, trx *sqlx.Tx) (*userLimits, error) {
	_ = newConfigBandwidthClasses(a.cfg.Bandwidth)
	userLimits := &userLimits{}
	return userLimits, nil
}
