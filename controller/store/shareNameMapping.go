package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type ShareNameMapping struct {
	Model
	ShareId int
	NameId  int
}

type ShareNameMappingWithShare struct {
	ShareNameMapping
	ShareToken   string `db:"share_token"`
	ShareDeleted bool   `db:"share_deleted"`
}

type ShareNameCleanupDetail struct {
	MappingId        int    `db:"mapping_id"`
	NameId           int    `db:"name_id"`
	Name             string `db:"name"`
	NameDeleted      bool   `db:"name_deleted"`
	Reserved         bool   `db:"reserved"`
	NamespaceID      int    `db:"namespace_id"`
	NamespaceName    string `db:"namespace_name"`
	NamespaceDeleted bool   `db:"namespace_deleted"`
}

func (str *Store) CreateShareNameMapping(snm *ShareNameMapping, trx *sqlx.Tx) (int, error) {
	stmt, err := trx.Prepare("insert into share_name_mappings (share_id, name_id) values ($1, $2) returning id")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing share name mapping insert statement")
	}
	var id int
	if err := stmt.QueryRow(snm.ShareId, snm.NameId).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "error executing share name mapping insert statement")
	}
	return id, nil
}

func (str *Store) GetShareNameMapping(id int, trx *sqlx.Tx) (*ShareNameMapping, error) {
	snm := &ShareNameMapping{}
	if err := trx.QueryRowx("select * from share_name_mappings where id = $1 and not deleted", id).StructScan(snm); err != nil {
		return nil, errors.Wrap(err, "error selecting share name mapping by id")
	}
	return snm, nil
}

func (str *Store) FindShareNameMappingsByShareId(shareId int, trx *sqlx.Tx) ([]*ShareNameMapping, error) {
	rows, err := trx.Queryx("select * from share_name_mappings where share_id = $1 and not deleted", shareId)
	if err != nil {
		return nil, errors.Wrap(err, "error finding share name mappings by share id")
	}
	var mappings []*ShareNameMapping
	for rows.Next() {
		snm := &ShareNameMapping{}
		if err := rows.StructScan(&snm); err != nil {
			return nil, errors.Wrap(err, "error scanning share name mapping")
		}
		mappings = append(mappings, snm)
	}
	return mappings, nil
}

func (str *Store) FindShareNameMappingsByNameId(nameId int, trx *sqlx.Tx) ([]*ShareNameMapping, error) {
	rows, err := trx.Queryx("select * from share_name_mappings where name_id = $1 and not deleted", nameId)
	if err != nil {
		return nil, errors.Wrap(err, "error finding share name mappings by name id")
	}
	var mappings []*ShareNameMapping
	for rows.Next() {
		snm := &ShareNameMapping{}
		if err := rows.StructScan(&snm); err != nil {
			return nil, errors.Wrap(err, "error scanning share name mapping")
		}
		mappings = append(mappings, snm)
	}
	return mappings, nil
}

func (str *Store) FindShareNameMappingsByNameIdWithShare(nameId int, trx *sqlx.Tx) ([]*ShareNameMappingWithShare, error) {
	sql := `select snm.id, snm.created_at, snm.updated_at, snm.deleted, snm.share_id, snm.name_id,
	               s.token as share_token, s.deleted as share_deleted
	        from share_name_mappings snm
	        join shares s on snm.share_id = s.id
	        where snm.name_id = $1 and not snm.deleted
	        order by snm.id`

	rows, err := trx.Queryx(sql, nameId)
	if err != nil {
		return nil, errors.Wrap(err, "error finding share name mappings with share by name id")
	}

	var mappings []*ShareNameMappingWithShare
	for rows.Next() {
		snm := &ShareNameMappingWithShare{}
		if err := rows.StructScan(snm); err != nil {
			return nil, errors.Wrap(err, "error scanning share name mapping with share")
		}
		mappings = append(mappings, snm)
	}

	return mappings, nil
}

func (str *Store) FindShareNameMappingByShareIdAndNameId(shareId, nameId int, trx *sqlx.Tx) (*ShareNameMapping, error) {
	snm := &ShareNameMapping{}
	if err := trx.QueryRowx("select * from share_name_mappings where share_id = $1 and name_id = $2 and not deleted", shareId, nameId).StructScan(snm); err != nil {
		return nil, errors.Wrap(err, "error selecting share name mapping by share id and name id")
	}
	return snm, nil
}

func (str *Store) FindShareNameCleanupDetailsByShareId(shareId int, trx *sqlx.Tx) ([]*ShareNameCleanupDetail, error) {
	sql := `select snm.id as mapping_id, snm.name_id,
	               n.name, n.deleted as name_deleted, n.reserved,
	               ns.id as namespace_id, ns.name as namespace_name, ns.deleted as namespace_deleted
	        from share_name_mappings snm
	        join names n on snm.name_id = n.id
	        join namespaces ns on n.namespace_id = ns.id
	        where snm.share_id = $1 and not snm.deleted
	        order by snm.id`

	rows, err := trx.Queryx(sql, shareId)
	if err != nil {
		return nil, errors.Wrap(err, "error finding share name cleanup details by share id")
	}

	var details []*ShareNameCleanupDetail
	for rows.Next() {
		detail := &ShareNameCleanupDetail{}
		if err := rows.StructScan(detail); err != nil {
			return nil, errors.Wrap(err, "error scanning share name cleanup detail")
		}
		details = append(details, detail)
	}

	return details, nil
}

func (str *Store) DeleteShareNameMapping(id int, trx *sqlx.Tx) error {
	stmt, err := trx.Prepare("update share_name_mappings set updated_at = current_timestamp, deleted = true where id = $1")
	if err != nil {
		return errors.Wrap(err, "error preparing share name mapping delete statement")
	}
	_, err = stmt.Exec(id)
	if err != nil {
		return errors.Wrap(err, "error executing share name mapping delete statement")
	}
	return nil
}
