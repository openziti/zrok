package limits

import (
	"github.com/jmoiron/sqlx"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/sdk/golang/sdk"
)

type UserLimits struct {
	resource  store.ResourceCountClass
	bandwidth store.BandwidthClass
	scopes    map[sdk.BackendMode]store.BandwidthClass
}

type AccountAction interface {
	HandleAccount(a *store.Account, rxBytes, txBytes int64, limit store.BandwidthClass, trx *sqlx.Tx) error
}

type EnvironmentAction interface {
	HandleEnvironment(e *store.Environment, rxBytes, txBytes int64, limit store.BandwidthClass, trx *sqlx.Tx) error
}

type ShareAction interface {
	HandleShare(s *store.Share, rxBytes, txBytes int64, limit store.BandwidthClass, trx *sqlx.Tx) error
}
