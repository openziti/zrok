package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type NamespaceFrontendMapping struct {
	Model
	NamespaceId int
	FrontendId  int
	IsDefault   bool
}

func (str *Store) CreateNamespaceFrontendMapping(nsId, feId int, isDefault bool, tx *sqlx.Tx) (int, error) {
	stmt, err := tx.Prepare("insert into namespace_frontend_mappings (namespace_id, frontend_id, is_default) values ($1, $2, $3) returning id")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing namespace frontend mapping insert statement")
	}
	var id int
	if err := stmt.QueryRow(nsId, feId, isDefault).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "error executing namespace frontend mapping insert statement")
	}
	return id, nil
}

func (str *Store) FindFrontendsForNamespace(namespaceId int, tx *sqlx.Tx) ([]*Frontend, error) {
	rows, err := tx.Queryx(`
		select f.* from frontends f 
		inner join namespace_frontend_mappings nfm on f.id = nfm.frontend_id 
		where nfm.namespace_id = $1 and not f.deleted and not nfm.deleted`, namespaceId)
	if err != nil {
		return nil, errors.Wrap(err, "error selecting frontends for namespace")
	}
	var frontends []*Frontend
	for rows.Next() {
		fe := &Frontend{}
		if err := rows.StructScan(fe); err != nil {
			return nil, errors.Wrap(err, "error scanning frontend")
		}
		frontends = append(frontends, fe)
	}
	return frontends, nil
}

func (str *Store) FindNamespacesForFrontend(frontendId int, tx *sqlx.Tx) ([]*Namespace, error) {
	rows, err := tx.Queryx(`
		select n.* from namespaces n 
		inner join namespace_frontend_mappings nfm on n.id = nfm.namespace_id 
		where nfm.frontend_id = $1 and not n.deleted and not nfm.deleted`, frontendId)
	if err != nil {
		return nil, errors.Wrap(err, "error selecting namespaces for frontend")
	}
	var namespaces []*Namespace
	for rows.Next() {
		ns := &Namespace{}
		if err := rows.StructScan(ns); err != nil {
			return nil, errors.Wrap(err, "error scanning namespace")
		}
		namespaces = append(namespaces, ns)
	}
	return namespaces, nil
}

func (str *Store) FindNamespaceFrontendMappingsForNamespace(namespaceId int, tx *sqlx.Tx) ([]*NamespaceFrontendMapping, error) {
	rows, err := tx.Queryx(`
		select nfm.* from namespace_frontend_mappings nfm 
		where nfm.namespace_id = $1 and not nfm.deleted`, namespaceId)
	if err != nil {
		return nil, errors.Wrap(err, "error selecting namespace frontend mappings for namespace")
	}
	var mappings []*NamespaceFrontendMapping
	for rows.Next() {
		mapping := &NamespaceFrontendMapping{}
		if err := rows.StructScan(mapping); err != nil {
			return nil, errors.Wrap(err, "error scanning namespace frontend mapping")
		}
		mappings = append(mappings, mapping)
	}
	return mappings, nil
}

func (str *Store) FindNamespaceFrontendMappingsForFrontend(frontendId int, tx *sqlx.Tx) ([]*NamespaceFrontendMapping, error) {
	rows, err := tx.Queryx(`
		select nfm.* from namespace_frontend_mappings nfm 
		where nfm.frontend_id = $1 and not nfm.deleted`, frontendId)
	if err != nil {
		return nil, errors.Wrap(err, "error selecting namespace frontend mappings for frontend")
	}
	var mappings []*NamespaceFrontendMapping
	for rows.Next() {
		mapping := &NamespaceFrontendMapping{}
		if err := rows.StructScan(mapping); err != nil {
			return nil, errors.Wrap(err, "error scanning namespace frontend mapping")
		}
		mappings = append(mappings, mapping)
	}
	return mappings, nil
}

func (str *Store) DeleteNamespaceFrontendMapping(nsId, feId int, tx *sqlx.Tx) error {
	stmt, err := tx.Prepare("update namespace_frontend_mappings set updated_at = current_timestamp, deleted = true where namespace_id = $1 and frontend_id = $2")
	if err != nil {
		return errors.Wrap(err, "error preparing namespace frontend mapping delete statement")
	}
	_, err = stmt.Exec(nsId, feId)
	if err != nil {
		return errors.Wrap(err, "error executing namespace frontend mapping delete statement")
	}
	return nil
}