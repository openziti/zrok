package store

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/openziti/zrok/util"
	"github.com/pkg/errors"
)

const Unlimited = -1

type BaseLimitClass interface {
	IsGlobal() bool
	GetLimitClassId() int
	String() string
}

type ResourceCountClass interface {
	BaseLimitClass
	GetEnvironments() int
	GetShares() int
	GetReservedShares() int
	GetUniqueNames() int
	GetShareFrontends() int
}

type BandwidthClass interface {
	BaseLimitClass
	IsScoped() bool
	GetBackendMode() sdk.BackendMode
	GetPeriodMinutes() int
	GetRxBytes() int64
	GetTxBytes() int64
	GetTotalBytes() int64
	GetLimitAction() LimitAction
}

type LimitClass struct {
	Model
	BackendMode    *sdk.BackendMode
	Environments   int
	Shares         int
	ReservedShares int
	UniqueNames    int
	ShareFrontends int
	PeriodMinutes  int
	RxBytes        int64
	TxBytes        int64
	TotalBytes     int64
	LimitAction    LimitAction
}

func (lc LimitClass) IsGlobal() bool {
	return false
}

func (lc LimitClass) IsScoped() bool {
	return lc.BackendMode != nil
}

func (lc LimitClass) GetLimitClassId() int {
	return lc.Id
}

func (lc LimitClass) GetEnvironments() int {
	return lc.Environments
}

func (lc LimitClass) GetShares() int {
	return lc.Shares
}

func (lc LimitClass) GetReservedShares() int {
	return lc.ReservedShares
}

func (lc LimitClass) GetUniqueNames() int {
	return lc.UniqueNames
}

func (lc LimitClass) GetShareFrontends() int {
	return lc.ShareFrontends
}

func (lc LimitClass) GetBackendMode() sdk.BackendMode {
	if lc.BackendMode == nil {
		return ""
	}
	return *lc.BackendMode
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
	out := fmt.Sprintf("LimitClass<#%d", lc.Id)
	if lc.BackendMode != nil {
		out += fmt.Sprintf(", backendMode: '%s'", *lc.BackendMode)
	}
	if lc.Environments > Unlimited {
		out += fmt.Sprintf(", environments: %d", lc.Environments)
	}
	if lc.Shares > Unlimited {
		out += fmt.Sprintf(", shares: %d", lc.Shares)
	}
	if lc.ReservedShares > Unlimited {
		out += fmt.Sprintf(", reservedShares: %d", lc.ReservedShares)
	}
	if lc.UniqueNames > Unlimited {
		out += fmt.Sprintf(", uniqueNames: %d", lc.UniqueNames)
	}
	if lc.ShareFrontends > Unlimited {
		out += fmt.Sprintf(", shareFrontends: %d", lc.ShareFrontends)
	}
	if lc.RxBytes > Unlimited || lc.TxBytes > Unlimited || lc.TotalBytes > Unlimited {
		out += fmt.Sprintf(", periodMinutes: %d", lc.PeriodMinutes)
	}
	if lc.RxBytes > Unlimited {
		out += fmt.Sprintf(", rxBytes: %v", util.BytesToSize(lc.RxBytes))
	}
	if lc.TxBytes > Unlimited {
		out += fmt.Sprintf(", txBytes: %v", util.BytesToSize(lc.TxBytes))
	}
	if lc.TotalBytes > Unlimited {
		out += fmt.Sprintf(", totalBytes: %v", util.BytesToSize(lc.TotalBytes))
	}
	out += fmt.Sprintf(", limitAction: '%v'>", lc.LimitAction)
	return out
}

var _ BandwidthClass = (*LimitClass)(nil)

func (str *Store) CreateLimitClass(lc *LimitClass, trx *sqlx.Tx) (int, error) {
	stmt, err := trx.Prepare("insert into limit_classes (backend_mode, environments, shares, reserved_shares, unique_names, share_frontends, period_minutes, rx_bytes, tx_bytes, total_bytes, limit_action) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) returning id")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing limit_classes insert statement")
	}
	var id int
	if err := stmt.QueryRow(lc.BackendMode, lc.Environments, lc.Shares, lc.ReservedShares, lc.UniqueNames, lc.ShareFrontends, lc.PeriodMinutes, lc.RxBytes, lc.TxBytes, lc.TotalBytes, lc.LimitAction).Scan(&id); err != nil {
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
