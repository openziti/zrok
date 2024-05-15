package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

func (str *Store) LimitCheckLock(acctId int, trx *sqlx.Tx) error {
	rows, err := trx.Queryx("select * from limit_check_locks where account_id = $1 for update", acctId)
	if err != nil {
		return errors.Wrap(err, "error preparing limit_check_locks select statement")
	}
	if !rows.Next() {
		return errors.Errorf("no limit_check_locks entry for account_id '%d'", acctId)
	}
	return nil
}
