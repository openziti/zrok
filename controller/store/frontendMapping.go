package store

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type FrontendMapping struct {
	Name       string    `db:"name"`
	Version    int64     `db:"version"`
	ShareToken string    `db:"share_token"`
	CreatedAt  time.Time `db:"created_at"`
}

func (str *Store) CreateFrontendMapping(fm *FrontendMapping, tx *sqlx.Tx) error {
	stmt, err := tx.Prepare("insert into frontend_mappings (name, version, share_token) values ($1, $2, $3)")
	if err != nil {
		return errors.Wrap(err, "error preparing frontend_mappings insert statement")
	}
	if _, err := stmt.Exec(fm.Name, fm.Version, fm.ShareToken); err != nil {
		return errors.Wrap(err, "error executing frontend_mappings insert statement")
	}
	return nil
}

func (str *Store) FindFrontendMapping(name string, version int64, tx *sqlx.Tx) (*FrontendMapping, error) {
	fm := &FrontendMapping{}
	if err := tx.QueryRowx("select * from frontend_mappings where name = $1 and version = $2", name, version).StructScan(fm); err != nil {
		return nil, errors.Wrap(err, "error selecting frontend mapping by name and version")
	}
	return fm, nil
}

func (str *Store) FindFrontendMappingsByName(name string, tx *sqlx.Tx) ([]*FrontendMapping, error) {
	rows, err := tx.Queryx("select * from frontend_mappings where name = $1 order by version desc", name)
	if err != nil {
		return nil, errors.Wrap(err, "error selecting frontend mappings by name")
	}
	var mappings []*FrontendMapping
	for rows.Next() {
		fm := &FrontendMapping{}
		if err := rows.StructScan(fm); err != nil {
			return nil, errors.Wrap(err, "error scanning frontend mapping")
		}
		mappings = append(mappings, fm)
	}
	return mappings, nil
}

func (str *Store) FindLatestFrontendMapping(name string, tx *sqlx.Tx) (*FrontendMapping, error) {
	fm := &FrontendMapping{}
	if err := tx.QueryRowx("select * from frontend_mappings where name = $1 order by version desc limit 1", name).StructScan(fm); err != nil {
		return nil, errors.Wrap(err, "error selecting latest frontend mapping by name")
	}
	return fm, nil
}

func (str *Store) DeleteFrontendMapping(name string, version int64, tx *sqlx.Tx) error {
	stmt, err := tx.Prepare("delete from frontend_mappings where name = $1 and version = $2")
	if err != nil {
		return errors.Wrap(err, "error preparing frontend_mappings delete statement")
	}
	if _, err := stmt.Exec(name, version); err != nil {
		return errors.Wrap(err, "error executing frontend_mappings delete statement")
	}
	return nil
}

func (str *Store) DeleteFrontendMappingsByName(name string, tx *sqlx.Tx) error {
	stmt, err := tx.Prepare("delete from frontend_mappings where name = $1")
	if err != nil {
		return errors.Wrap(err, "error preparing frontend_mappings delete by name statement")
	}
	if _, err := stmt.Exec(name); err != nil {
		return errors.Wrap(err, "error executing frontend_mappings delete by name statement")
	}
	return nil
}