package limits

import (
	"github.com/jmoiron/sqlx"
	"github.com/openziti/zrok/controller/store"
)

type AccountAction interface {
	HandleAccount(a *store.Account, rxBytes, txBytes int64, limit *BandwidthPerPeriod, trx *sqlx.Tx) error
}

type EnvironmentAction interface {
	HandleEnvironment(e *store.Environment, rxBytes, txBytes int64, limit *BandwidthPerPeriod, trx *sqlx.Tx) error
}

type ShareAction interface {
	HandleShare(s *store.Share, rxBytes, txBytes int64, limit *BandwidthPerPeriod, trx *sqlx.Tx) error
}
