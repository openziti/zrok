package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type NamespaceGrant struct {
	Model
	NamespaceId int
	AccountId   int
}

func (str *Store) CreateNamespaceGrant(ng *NamespaceGrant, tx *sqlx.Tx) (int, error) {
	stmt, err := tx.Prepare("insert into namespace_grants (namespace_id, account_id) values ($1, $2) returning id")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing namespace grant insert statement")
	}
	var id int
	if err := stmt.QueryRow(ng.NamespaceId, ng.AccountId).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "error executing namespace grant insert statement")
	}
	return id, nil
}

func (str *Store) GetNamespaceGrant(id int, tx *sqlx.Tx) (*NamespaceGrant, error) {
	ng := &NamespaceGrant{}
	if err := tx.QueryRowx("select * from namespace_grants where id = $1 and not deleted", id).StructScan(ng); err != nil {
		return nil, errors.Wrap(err, "error selecting namespace grant by id")
	}
	return ng, nil
}

func (str *Store) FindNamespaceGrantsForAccount(accountId int, tx *sqlx.Tx) ([]*NamespaceGrant, error) {
	rows, err := tx.Queryx("select * from namespace_grants where account_id = $1 and not deleted", accountId)
	if err != nil {
		return nil, errors.Wrap(err, "error finding namespace grants for account")
	}
	var grants []*NamespaceGrant
	for rows.Next() {
		ng := &NamespaceGrant{}
		if err := rows.StructScan(&ng); err != nil {
			return nil, errors.Wrap(err, "error scanning namespace grant")
		}
		grants = append(grants, ng)
	}
	return grants, nil
}

func (str *Store) FindNamespaceGrantsForNamespace(namespaceId int, tx *sqlx.Tx) ([]*NamespaceGrant, error) {
	rows, err := tx.Queryx("select * from namespace_grants where namespace_id = $1 and not deleted", namespaceId)
	if err != nil {
		return nil, errors.Wrap(err, "error finding namespace grants for namespace")
	}
	var grants []*NamespaceGrant
	for rows.Next() {
		ng := &NamespaceGrant{}
		if err := rows.StructScan(&ng); err != nil {
			return nil, errors.Wrap(err, "error scanning namespace grant")
		}
		grants = append(grants, ng)
	}
	return grants, nil
}

func (str *Store) CheckNamespaceGrant(namespaceId, accountId int, tx *sqlx.Tx) (bool, error) {
	var count int
	if err := tx.QueryRow("select count(*) from namespace_grants where namespace_id = $1 and account_id = $2 and not deleted", namespaceId, accountId).Scan(&count); err != nil {
		return false, errors.Wrap(err, "error checking namespace grant")
	}
	return count > 0, nil
}

func (str *Store) DeleteNamespaceGrant(id int, tx *sqlx.Tx) error {
	stmt, err := tx.Prepare("update namespace_grants set updated_at = current_timestamp, deleted = true where id = $1")
	if err != nil {
		return errors.Wrap(err, "error preparing namespace grant delete statement")
	}
	_, err = stmt.Exec(id)
	if err != nil {
		return errors.Wrap(err, "error executing namespace grant delete statement")
	}
	return nil
}