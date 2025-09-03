package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type Name struct {
	Model
	NamespaceId int
	Name        string
	AccountId   int
	Reserved    bool
}

func (str *Store) CreateName(an *Name, tx *sqlx.Tx) (int, error) {
	stmt, err := tx.Prepare("insert into names (namespace_id, name, account_id) values ($1, $2, $3) returning id")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing name insert statement")
	}
	var id int
	if err := stmt.QueryRow(an.NamespaceId, an.Name, an.AccountId).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "error executing name insert statement")
	}
	return id, nil
}

func (str *Store) GetName(id int, tx *sqlx.Tx) (*Name, error) {
	an := &Name{}
	if err := tx.QueryRowx("select * from names where id = $1 and not deleted", id).StructScan(an); err != nil {
		return nil, errors.Wrap(err, "error selecting name by id")
	}
	return an, nil
}

func (str *Store) FindNameByNamespaceAndName(namespaceId int, name string, tx *sqlx.Tx) (*Name, error) {
	an := &Name{}
	if err := tx.QueryRowx("select * from names where namespace_id = $1 and name = $2 and not deleted", namespaceId, name).StructScan(an); err != nil {
		return nil, errors.Wrap(err, "error selecting name by namespace and name")
	}
	return an, nil
}

func (str *Store) FindNamesForNamespace(namespaceId int, tx *sqlx.Tx) ([]*Name, error) {
	rows, err := tx.Queryx("select * from names where namespace_id = $1 and not deleted order by name", namespaceId)
	if err != nil {
		return nil, errors.Wrap(err, "error finding names for namespace")
	}
	var names []*Name
	for rows.Next() {
		an := &Name{}
		if err := rows.StructScan(&an); err != nil {
			return nil, errors.Wrap(err, "error scanning name")
		}
		names = append(names, an)
	}
	return names, nil
}

func (str *Store) FindNamesForAccount(accountId int, tx *sqlx.Tx) ([]*Name, error) {
	rows, err := tx.Queryx("select * from names where account_id = $1 and not deleted order by name", accountId)
	if err != nil {
		return nil, errors.Wrap(err, "error finding names for account")
	}
	var names []*Name
	for rows.Next() {
		an := &Name{}
		if err := rows.StructScan(&an); err != nil {
			return nil, errors.Wrap(err, "error scanning name")
		}
		names = append(names, an)
	}
	return names, nil
}

func (str *Store) FindNamesForAccountAndNamespace(accountId, namespaceId int, tx *sqlx.Tx) ([]*Name, error) {
	rows, err := tx.Queryx("select * from names where account_id = $1 and namespace_id = $2 and not deleted order by name", accountId, namespaceId)
	if err != nil {
		return nil, errors.Wrap(err, "error finding names for account and namespace")
	}
	var names []*Name
	for rows.Next() {
		an := &Name{}
		if err := rows.StructScan(&an); err != nil {
			return nil, errors.Wrap(err, "error scanning name")
		}
		names = append(names, an)
	}
	return names, nil
}

func (str *Store) CheckNameAvailability(namespaceId int, name string, tx *sqlx.Tx) (bool, error) {
	var count int
	if err := tx.QueryRow("select count(*) from names where namespace_id = $1 and name = $2 and not deleted", namespaceId, name).Scan(&count); err != nil {
		return false, errors.Wrap(err, "error checking name availability")
	}
	return count == 0, nil
}

func (str *Store) DeleteName(id int, tx *sqlx.Tx) error {
	stmt, err := tx.Prepare("update names set updated_at = current_timestamp, deleted = true where id = $1")
	if err != nil {
		return errors.Wrap(err, "error preparing name delete statement")
	}
	_, err = stmt.Exec(id)
	if err != nil {
		return errors.Wrap(err, "error executing name delete statement")
	}
	return nil
}
