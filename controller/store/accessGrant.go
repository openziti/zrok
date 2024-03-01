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

func (str *Store) CreateAccessGrant(shareId, accountId, tx *sqlx.Tx) (int, error) {
	stmt, err := tx.Prepare("insert into access_grants (share_id, account_id) values ($1, $2) returning id")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing access_grant insert statement")
	}
	var id int
	if err := stmt.QueryRow(shareId, accountId).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "error executing access_grant insert statement")
	}
	return id, nil
}

func (str *Store) FindAccessGrantsForAccount(accountId int, tx *sqlx.Tx) ([]*AccessGrant, error) {
	rows, err := tx.Queryx("select access_grants.* from access_grants where account_id = $1 and not deleted", accountId)
	if err != nil {
		return nil, errors.Wrap(err, "error selecting access_grants by account_id")
	}
	var ags []*AccessGrant
	for rows.Next() {
		ag := &AccessGrant{}
		if err := rows.StructScan(ag); err != nil {
			return nil, errors.Wrap(err, "error scanning access_grant")
		}
		ags = append(ags, ag)
	}
	return ags, nil
}

func (str *Store) DeleteAccessGrant(id int, tx *sqlx.Tx) error {
	stmt, err := tx.Prepare("update access_grants set updated_at = current_timestamp, deleted = true where id = $1")
	if err != nil {
		return errors.Wrap(err, "error preparing access_grants delete statement")
	}
	_, err = stmt.Exec(id)
	if err != nil {
		return errors.Wrap(err, "error executing access_grants delete statement")
	}
	return nil
}
