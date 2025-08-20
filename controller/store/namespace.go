package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type Namespace struct {
	Model
	Name        string
	Description string
}

func (str *Store) CreateNamespace(ns *Namespace, tx *sqlx.Tx) (int, error) {
	stmt, err := tx.Prepare("insert into namespaces (name, description) values ($1, $2) returning id")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing namespace insert statement")
	}
	var id int
	if err := stmt.QueryRow(ns.Name, ns.Description).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "error executing namespace insert statement")
	}
	return id, nil
}

func (str *Store) GetNamespace(id int, tx *sqlx.Tx) (*Namespace, error) {
	ns := &Namespace{}
	if err := tx.QueryRowx("select * from namespaces where id = $1 and not deleted", id).StructScan(ns); err != nil {
		return nil, errors.Wrap(err, "error selecting namespace by id")
	}
	return ns, nil
}

func (str *Store) FindNamespaceByName(name string, tx *sqlx.Tx) (*Namespace, error) {
	ns := &Namespace{}
	if err := tx.QueryRowx("select * from namespaces where name = $1 and not deleted", name).StructScan(ns); err != nil {
		return nil, errors.Wrap(err, "error selecting namespace by name")
	}
	return ns, nil
}

func (str *Store) FindNamespaces(tx *sqlx.Tx) ([]*Namespace, error) {
	rows, err := tx.Queryx("select * from namespaces where not deleted order by name")
	if err != nil {
		return nil, errors.Wrap(err, "error finding namespaces")
	}
	var namespaces []*Namespace
	for rows.Next() {
		ns := &Namespace{}
		if err := rows.StructScan(&ns); err != nil {
			return nil, errors.Wrap(err, "error scanning namespace")
		}
		namespaces = append(namespaces, ns)
	}
	return namespaces, nil
}

func (str *Store) DeleteNamespace(id int, tx *sqlx.Tx) error {
	stmt, err := tx.Prepare("update namespaces set updated_at = current_timestamp, deleted = true where id = $1")
	if err != nil {
		return errors.Wrap(err, "error preparing namespace delete statement")
	}
	_, err = stmt.Exec(id)
	if err != nil {
		return errors.Wrap(err, "error executing namespace delete statement")
	}
	return nil
}