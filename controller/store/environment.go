package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type Environment struct {
	Model
	AccountId   int
	Description string
	Host        string
	Address     string
	ZId         string
}

func (self *Store) CreateEnvironment(accountId int, i *Environment, tx *sqlx.Tx) (int, error) {
	stmt, err := tx.Prepare("insert into environments (account_id, description, host, address, z_id) values ($1, $2, $3, $4, $5) returning id")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing environments insert statement")
	}
	var id int
	if err := stmt.QueryRow(accountId, i.Description, i.Host, i.Address, i.ZId).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "error executing environments insert statement")
	}
	return id, nil
}

func (self *Store) GetEnvironment(id int, tx *sqlx.Tx) (*Environment, error) {
	i := &Environment{}
	if err := tx.QueryRowx("select * from environments where id = $1", id).StructScan(i); err != nil {
		return nil, errors.Wrap(err, "error selecting environment by id")
	}
	return i, nil
}

func (self *Store) FindEnvironmentsForAccount(accountId int, tx *sqlx.Tx) ([]*Environment, error) {
	rows, err := tx.Queryx("select environments.* from environments where account_id = $1", accountId)
	if err != nil {
		return nil, errors.Wrap(err, "error selecting environments by account id")
	}
	var is []*Environment
	for rows.Next() {
		i := &Environment{}
		if err := rows.StructScan(i); err != nil {
			return nil, errors.Wrap(err, "error scanning environment")
		}
		is = append(is, i)
	}
	return is, nil
}

func (self *Store) DeleteEnvironment(id int, tx *sqlx.Tx) error {
	stmt, err := tx.Prepare("delete from environments where id = $1")
	if err != nil {
		return errors.Wrap(err, "error preparing environments delete statement")
	}
	_, err = stmt.Exec(id)
	if err != nil {
		return errors.Wrap(err, "error executing environments delete statement")
	}
	return nil
}
