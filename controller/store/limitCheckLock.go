package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

func (str *Store) LimitCheckLock(acctId int, trx *sqlx.Tx) error {
	stmt, err := trx.Prepare("insert into limit_check_locks (account_id) values ($1) on conflict (account_id) do update set updated_at = current_timestamp")
	if err != nil {
		return errors.Wrap(err, "error preparing upsert on limit_check_locks")
	}
	if _, err := stmt.Exec(acctId); err != nil {
		return errors.Wrap(err, "error executing upsert on limit_check_locks")
	}
	return nil
}
