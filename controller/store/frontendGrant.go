package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

func (str *Store) IsFrontendGrantedToAccount(acctId, frontendId int, trx *sqlx.Tx) (bool, error) {
	stmt, err := trx.Prepare("select count(0) from frontend_grants where account_id = $1 AND frontend_id = $2")
	if err != nil {
		return false, errors.Wrap(err, "error preparing frontend_grants select statement")
	}
	var count int
	if err := stmt.QueryRow(acctId, frontendId).Scan(&count); err != nil {
		return false, errors.Wrap(err, "error querying frontend_grants count")
	}
	return count > 0, nil
}
