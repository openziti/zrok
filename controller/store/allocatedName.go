package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type AllocatedName struct {
	Model
	NamespaceId int
	Name        string
	AccountId   int
}

func (str *Store) CreateAllocatedName(an *AllocatedName, tx *sqlx.Tx) (int, error) {
	stmt, err := tx.Prepare("insert into allocated_names (namespace_id, name, account_id) values ($1, $2, $3) returning id")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing allocated name insert statement")
	}
	var id int
	if err := stmt.QueryRow(an.NamespaceId, an.Name, an.AccountId).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "error executing allocated name insert statement")
	}
	return id, nil
}

func (str *Store) GetAllocatedName(id int, tx *sqlx.Tx) (*AllocatedName, error) {
	an := &AllocatedName{}
	if err := tx.QueryRowx("select * from allocated_names where id = $1 and not deleted", id).StructScan(an); err != nil {
		return nil, errors.Wrap(err, "error selecting allocated name by id")
	}
	return an, nil
}

func (str *Store) FindAllocatedNameByNamespaceAndName(namespaceId int, name string, tx *sqlx.Tx) (*AllocatedName, error) {
	an := &AllocatedName{}
	if err := tx.QueryRowx("select * from allocated_names where namespace_id = $1 and name = $2 and not deleted", namespaceId, name).StructScan(an); err != nil {
		return nil, errors.Wrap(err, "error selecting allocated name by namespace and name")
	}
	return an, nil
}

func (str *Store) FindAllocatedNamesForNamespace(namespaceId int, tx *sqlx.Tx) ([]*AllocatedName, error) {
	rows, err := tx.Queryx("select * from allocated_names where namespace_id = $1 and not deleted order by name", namespaceId)
	if err != nil {
		return nil, errors.Wrap(err, "error finding allocated names for namespace")
	}
	var names []*AllocatedName
	for rows.Next() {
		an := &AllocatedName{}
		if err := rows.StructScan(&an); err != nil {
			return nil, errors.Wrap(err, "error scanning allocated name")
		}
		names = append(names, an)
	}
	return names, nil
}

func (str *Store) FindAllocatedNamesForAccount(accountId int, tx *sqlx.Tx) ([]*AllocatedName, error) {
	rows, err := tx.Queryx("select * from allocated_names where account_id = $1 and not deleted order by name", accountId)
	if err != nil {
		return nil, errors.Wrap(err, "error finding allocated names for account")
	}
	var names []*AllocatedName
	for rows.Next() {
		an := &AllocatedName{}
		if err := rows.StructScan(&an); err != nil {
			return nil, errors.Wrap(err, "error scanning allocated name")
		}
		names = append(names, an)
	}
	return names, nil
}

func (str *Store) FindAllocatedNamesForAccountAndNamespace(accountId, namespaceId int, tx *sqlx.Tx) ([]*AllocatedName, error) {
	rows, err := tx.Queryx("select * from allocated_names where account_id = $1 and namespace_id = $2 and not deleted order by name", accountId, namespaceId)
	if err != nil {
		return nil, errors.Wrap(err, "error finding allocated names for account and namespace")
	}
	var names []*AllocatedName
	for rows.Next() {
		an := &AllocatedName{}
		if err := rows.StructScan(&an); err != nil {
			return nil, errors.Wrap(err, "error scanning allocated name")
		}
		names = append(names, an)
	}
	return names, nil
}

func (str *Store) CheckNameAvailability(namespaceId int, name string, tx *sqlx.Tx) (bool, error) {
	var count int
	if err := tx.QueryRow("select count(*) from allocated_names where namespace_id = $1 and name = $2 and not deleted", namespaceId, name).Scan(&count); err != nil {
		return false, errors.Wrap(err, "error checking name availability")
	}
	return count == 0, nil
}

func (str *Store) DeleteAllocatedName(id int, tx *sqlx.Tx) error {
	stmt, err := tx.Prepare("update allocated_names set updated_at = current_timestamp, deleted = true where id = $1")
	if err != nil {
		return errors.Wrap(err, "error preparing allocated name delete statement")
	}
	_, err = stmt.Exec(id)
	if err != nil {
		return errors.Wrap(err, "error executing allocated name delete statement")
	}
	return nil
}
