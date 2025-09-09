package store

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type Mapping struct {
	Mapping    string
	Version    int64
	ShareToken string
	CreatedAt  time.Time `db:"created_at"`
}

func (str *Store) CreateMapping(m *Mapping, tx *sqlx.Tx) error {
	stmt, err := tx.Prepare("insert into mappings (mapping, version, share_token) values ($1, $2, $3)")
	if err != nil {
		return errors.Wrap(err, "error preparing mappings insert statement")
	}
	if _, err := stmt.Exec(m.Mapping, m.Version, m.ShareToken); err != nil {
		return errors.Wrap(err, "error executing mappings insert statement")
	}
	return nil
}

func (str *Store) GetMapping(mapping string, tx *sqlx.Tx) (*Mapping, error) {
	m := &Mapping{}
	if err := tx.QueryRowx("select * from mappings where mapping = $1 order by version desc limit 1", mapping).StructScan(m); err != nil {
		return nil, errors.Wrap(err, "error selecting mapping with highest version")
	}
	return m, nil
}

func (str *Store) FindMappingsByShareToken(shareToken string, tx *sqlx.Tx) ([]*Mapping, error) {
	rows, err := tx.Queryx("select * from mappings where share_token = $1 order by version desc", shareToken)
	if err != nil {
		return nil, errors.Wrap(err, "error selecting mappings by share token")
	}
	var mappings []*Mapping
	for rows.Next() {
		m := &Mapping{}
		if err := rows.StructScan(m); err != nil {
			return nil, errors.Wrap(err, "error scanning mapping")
		}
		mappings = append(mappings, m)
	}
	return mappings, nil
}

func (str *Store) DeleteMapping(mapping string, version int64, tx *sqlx.Tx) error {
	stmt, err := tx.Prepare("delete from mappings where mapping = $1 and version = $2")
	if err != nil {
		return errors.Wrap(err, "error preparing mappings delete statement")
	}
	_, err = stmt.Exec(mapping, version)
	if err != nil {
		return errors.Wrap(err, "error executing mappings delete statement")
	}
	return nil
}

func (str *Store) DeleteMappings(mapping string, tx *sqlx.Tx) error {
	stmt, err := tx.Prepare("delete from mappings where mapping = $1")
	if err != nil {
		return errors.Wrap(err, "error preparing mappings delete statement")
	}
	_, err = stmt.Exec(mapping)
	if err != nil {
		return errors.Wrap(err, "error executing mappings delete statement")
	}
	return nil
}
