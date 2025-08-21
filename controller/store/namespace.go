package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type Namespace struct {
	Model
	Token       string
	Name        string
	Description string
	Open        bool
}

func (str *Store) CreateNamespace(ns *Namespace, tx *sqlx.Tx) (int, error) {
	stmt, err := tx.Prepare("insert into namespaces (token, name, description, open) values ($1, $2, $3, $4) returning id")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing namespace insert statement")
	}
	var id int
	if err := stmt.QueryRow(ns.Token, ns.Name, ns.Description, ns.Open).Scan(&id); err != nil {
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

func (str *Store) FindNamespaceWithName(name string, tx *sqlx.Tx) (*Namespace, error) {
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

func (str *Store) FindNamespaceWithToken(token string, tx *sqlx.Tx) (*Namespace, error) {
	ns := &Namespace{}
	if err := tx.QueryRowx("select * from namespaces where token = $1 and not deleted", token).StructScan(ns); err != nil {
		return nil, errors.Wrap(err, "error selecting namespace by token")
	}
	return ns, nil
}

func (str *Store) UpdateNamespace(ns *Namespace, tx *sqlx.Tx) error {
	stmt, err := tx.Prepare("update namespaces set name = $1, description = $2, open = $3, updated_at = current_timestamp where id = $4")
	if err != nil {
		return errors.Wrap(err, "error preparing namespace update statement")
	}
	_, err = stmt.Exec(ns.Name, ns.Description, ns.Open, ns.Id)
	if err != nil {
		return errors.Wrap(err, "error executing namespace update statement")
	}
	return nil
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
