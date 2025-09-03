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

func (str *Store) CreateShareNameMapping(snm *ShareNameMapping, tx *sqlx.Tx) (int, error) {
	stmt, err := tx.Prepare("insert into share_name_mappings (share_id, name_id) values ($1, $2) returning id")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing share name mapping insert statement")
	}
	var id int
	if err := stmt.QueryRow(snm.ShareId, snm.NameId).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "error executing share name mapping insert statement")
	}
	return id, nil
}

func (str *Store) GetShareNameMapping(id int, tx *sqlx.Tx) (*ShareNameMapping, error) {
	snm := &ShareNameMapping{}
	if err := tx.QueryRowx("select * from share_name_mappings where id = $1 and not deleted", id).StructScan(snm); err != nil {
		return nil, errors.Wrap(err, "error selecting share name mapping by id")
	}
	return snm, nil
}

func (str *Store) FindShareNameMappingsByShareId(shareId int, tx *sqlx.Tx) ([]*ShareNameMapping, error) {
	rows, err := tx.Queryx("select * from share_name_mappings where share_id = $1 and not deleted", shareId)
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

func (str *Store) FindShareNameMappingsByNameId(nameId int, tx *sqlx.Tx) ([]*ShareNameMapping, error) {
	rows, err := tx.Queryx("select * from share_name_mappings where name_id = $1 and not deleted", nameId)
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

func (str *Store) FindShareNameMappingByShareIdAndNameId(shareId, nameId int, tx *sqlx.Tx) (*ShareNameMapping, error) {
	snm := &ShareNameMapping{}
	if err := tx.QueryRowx("select * from share_name_mappings where share_id = $1 and name_id = $2 and not deleted", shareId, nameId).StructScan(snm); err != nil {
		return nil, errors.Wrap(err, "error selecting share name mapping by share id and name id")
	}
	return snm, nil
}

func (str *Store) DeleteShareNameMapping(id int, tx *sqlx.Tx) error {
	stmt, err := tx.Prepare("update share_name_mappings set updated_at = current_timestamp, deleted = true where id = $1")
	if err != nil {
		return errors.Wrap(err, "error preparing share name mapping delete statement")
	}
	_, err = stmt.Exec(id)
	if err != nil {
		return errors.Wrap(err, "error executing share name mapping delete statement")
	}
	return nil
}