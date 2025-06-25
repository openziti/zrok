package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

func (str *Store) IsFrontendGrantedToAccount(frontendId, accountId int, trx *sqlx.Tx) (bool, error) {
	stmt, err := trx.Prepare("select count(0) from frontend_grants where frontend_id = $1 AND account_id = $2 and not deleted")
	if err != nil {
		return false, errors.Wrap(err, "error preparing frontend_grants select statement")
	}
	var count int
	if err := stmt.QueryRow(frontendId, accountId).Scan(&count); err != nil {
		return false, errors.Wrap(err, "error querying frontend_grants count")
	}
	return count > 0, nil
}

func (str *Store) CreateFrontendGrant(frontendId, accountId int, trx *sqlx.Tx) (int, error) {
	stmt, err := trx.Prepare("insert into frontend_grants (frontend_id, account_id) values ($1, $2) returning id")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing frontend_grants insert statement")
	}
	var id int
	if err := stmt.QueryRow(frontendId, accountId).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "error executing frontend_grants insert statement")
	}
	return id, nil
}

func (str *Store) DeleteFrontendGrant(frontendId, accountId int, trx *sqlx.Tx) error {
	stmt, err := trx.Prepare("delete from frontend_grants where frontend_id = $1 and account_id = $2")
	if err != nil {
		return errors.Wrap(err, "error preparing frontend_grants delete for frontend and acount statement")
	}
	_, err = stmt.Exec(frontendId, accountId)
	if err != nil {
		return errors.Wrap(err, "error executing frontend_grants for frontend and account statement")
	}
	return nil
}
