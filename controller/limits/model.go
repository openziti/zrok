package limits

import (
	"github.com/jmoiron/sqlx"
	"github.com/openziti/zrok/v2/controller/store"
)

type AccountAction interface {
	HandleAccount(a *store.Account, rxBytes, txBytes int64, bwc store.BandwidthClass, ul *userLimits, trx *sqlx.Tx) error
}
