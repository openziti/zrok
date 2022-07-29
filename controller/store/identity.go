package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type Identity struct {
	Model
	AccountId int
	ZitiId    string
	Active    bool
}

func (self *Store) CreateIdentity(accountId int, i *Identity, tx *sqlx.Tx) (int, error) {
	stmt, err := tx.Prepare("insert into identities (account_id, ziti_id, active) values (?, ?, true)")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing identities insert statement")
	}
	res, err := stmt.Exec(accountId, i.ZitiId)
	if err != nil {
		return 0, errors.Wrap(err, "error executing identities insert statement")
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, errors.Wrap(err, "error retrieving last identities insert id")
	}
	return int(id), nil
}

func (self *Store) GetIdentity(id int, tx *sqlx.Tx) (*Identity, error) {
	i := &Identity{}
	if err := tx.QueryRowx("select * from identities where id = ?", id).StructScan(i); err != nil {
		return nil, errors.Wrap(err, "error selecting identity by id")
	}
	return i, nil
}

func (self *Store) FindIdentitiesForAccount(accountId int, tx *sqlx.Tx) ([]*Identity, error) {
	rows, err := tx.Queryx("select identities.* from identities where account_id = ?", accountId)
	if err != nil {
		return nil, errors.Wrap(err, "error selecting identities by account id")
	}
	var is []*Identity
	for rows.Next() {
		i := &Identity{}
		if err := rows.StructScan(i); err != nil {
			return nil, errors.Wrap(err, "error scanning identity")
		}
		is = append(is, i)
	}
	return is, nil
}

func (self *Store) DeleteIdentity(id int, tx *sqlx.Tx) error {
	stmt, err := tx.Prepare("delete from identities where id = ?")
	if err != nil {
		return errors.Wrap(err, "error preparing identities delete statement")
	}
	_, err = stmt.Exec(id)
	if err != nil {
		return errors.Wrap(err, "error executing identities delete statement")
	}
	return nil
}
