package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type AccessGrant struct {
	Model
	ShareId   int
	AccountId int
}

func (str *Store) CreateAccessGrant(shareId, accountId int, tx *sqlx.Tx) (int, error) {
	stmt, err := tx.Prepare("insert into access_grants (share_id, account_id) values ($1, $2) returning id")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing access_grants insert statement")
	}
	var id int
	if err := stmt.QueryRow(shareId, accountId).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "error executing access_grants insert statement")
	}
	return id, nil
}

func (str *Store) IsAccessGrantedToAccountForShare(shrId, acctId int, tx *sqlx.Tx) (int, error) {
	count := 0
	err := tx.QueryRowx("select count(0) from access_grants where share_id = $1 and account_id = $2 and not deleted", shrId, acctId).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "error selecting access_grants by share_id and account_id")
	}
	return count, nil
}

func (str *Store) DeleteAccessGrantsForShare(shrId int, tx *sqlx.Tx) error {
	stmt, err := tx.Prepare("update access_grants set updated_at = current_timestamp, deleted = true where share_id = $1")
	if err != nil {
		return errors.Wrap(err, "error preparing access_grants delete for shares statement")
	}
	_, err = stmt.Exec(shrId)
	if err != nil {
		return errors.Wrap(err, "error executing access_grants delete for shares statement")
	}
	return nil
}

func (str *Store) DeleteAccessGrantsForShareAndAccount(shrId, acctId int, tx *sqlx.Tx) error {
	stmt, err := tx.Prepare("update access_grants set updated_at = current_timestamp, deleted = true where share_id = $1 and account_id = $2")
	if err != nil {
		return errors.Wrap(err, "error preparing access_grants delete for share and account statement")
	}
	_, err = stmt.Exec(shrId, acctId)
	if err != nil {
		return errors.Wrap(err, "error executing access_grants delete for share and account statement")
	}
	return nil
}
