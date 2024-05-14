package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/pkg/errors"
)

type LimitClass struct {
	Model
	LimitScope    LimitScope
	LimitAction   LimitAction
	ShareMode     sdk.ShareMode
	BackendMode   sdk.BackendMode
	PeriodMinutes int
	RxBytes       int64
	TxBytes       int64
	TotalBytes    int64
}

func (str *Store) CreateLimitClass(lc *LimitClass, trx *sqlx.Tx) (int, error) {
	stmt, err := trx.Prepare("insert into limit_classes (limit_scope, limit_action, share_mode, backend_mode, period_minutes, rx_bytes, tx_bytes, total_bytes) values ($1, $2, $3, $4, $5, $6, $7, $8) returning id")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing limit_classes insert statement")
	}
	var id int
	if err := stmt.QueryRow(lc.LimitScope, lc.LimitAction, lc.ShareMode, lc.BackendMode, lc.PeriodMinutes).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "error executing limit_classes insert statement")
	}
	return id, nil
}
