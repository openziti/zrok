package store

import (
	"encoding/json"
	"github.com/jmoiron/sqlx"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/pkg/errors"
)

type BandwidthClass interface {
	IsGlobal() bool
	GetLimitClassId() int
	GetShareMode() sdk.ShareMode
	GetBackendMode() sdk.BackendMode
	GetPeriodMinutes() int
	GetRxBytes() int64
	GetTxBytes() int64
	GetTotalBytes() int64
	GetLimitAction() LimitAction
}

type LimitClass struct {
	Model
	ShareMode      sdk.ShareMode
	BackendMode    sdk.BackendMode
	Environments   int
	Shares         int
	ReservedShares int
	UniqueNames    int
	PeriodMinutes  int
	RxBytes        int64
	TxBytes        int64
	TotalBytes     int64
	LimitAction    LimitAction
}

func (lc LimitClass) IsGlobal() bool {
	return false
}

func (lc LimitClass) GetLimitClassId() int {
	return lc.Id
}

func (lc LimitClass) GetShareMode() sdk.ShareMode {
	return lc.ShareMode
}

func (lc LimitClass) GetBackendMode() sdk.BackendMode {
	return lc.BackendMode
}

func (lc LimitClass) GetPeriodMinutes() int {
	return lc.PeriodMinutes
}

func (lc LimitClass) GetRxBytes() int64 {
	return lc.RxBytes
}

func (lc LimitClass) GetTxBytes() int64 {
	return lc.TxBytes
}

func (lc LimitClass) GetTotalBytes() int64 {
	return lc.TotalBytes
}

func (lc LimitClass) GetLimitAction() LimitAction {
	return lc.LimitAction
}

func (lc LimitClass) String() string {
	out, err := json.MarshalIndent(&lc, "", "  ")
	if err != nil {
		return ""

	}
	return string(out)
}

var _ BandwidthClass = (*LimitClass)(nil)

func (str *Store) CreateLimitClass(lc *LimitClass, trx *sqlx.Tx) (int, error) {
	stmt, err := trx.Prepare("insert into limit_classes (share_mode, backend_mode, environments, shares, reserved_shares, unique_names, period_minutes, rx_bytes, tx_bytes, total_bytes, limit_action) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) returning id")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing limit_classes insert statement")
	}
	var id int
	if err := stmt.QueryRow(lc.ShareMode, lc.BackendMode, lc.Environments, lc.Shares, lc.ReservedShares, lc.UniqueNames, lc.PeriodMinutes, lc.RxBytes, lc.TxBytes, lc.TotalBytes, lc.LimitAction).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "error executing limit_classes insert statement")
	}
	return id, nil
}

func (str *Store) GetLimitClass(lcId int, trx *sqlx.Tx) (*LimitClass, error) {
	lc := &LimitClass{}
	if err := trx.QueryRowx("select * from limit_classes where id = $1", lcId).StructScan(lc); err != nil {
		return nil, errors.Wrap(err, "error selecting limit_class by id")
	}
	return lc, nil
}
