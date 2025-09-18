package store

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type FrontendMapping struct {
	FrontendToken string    `db:"frontend_token"`
	Name          string    `db:"name"`
	Version       int64     `db:"version"`
	ShareToken    string    `db:"share_token"`
	CreatedAt     time.Time `db:"created_at"`
}

func (str *Store) CreateFrontendMapping(fm *FrontendMapping, trx *sqlx.Tx) error {
	stmt, err := trx.Prepare("insert into frontend_mappings (frontend_token, name, version, share_token) values ($1, $2, $3, $4)")
	if err != nil {
		return errors.Wrap(err, "error preparing frontend_mappings insert statement")
	}
	if _, err := stmt.Exec(fm.FrontendToken, fm.Name, fm.Version, fm.ShareToken); err != nil {
		return errors.Wrap(err, "error executing frontend_mappings insert statement")
	}
	return nil
}

func (str *Store) FindFrontendMapping(frontendToken, name string, version int64, trx *sqlx.Tx) (*FrontendMapping, error) {
	fm := &FrontendMapping{}
	if err := trx.QueryRowx("select * from frontend_mappings where frontend_token = $1 and name = $2 and version = $3", frontendToken, name, version).StructScan(fm); err != nil {
		return nil, errors.Wrap(err, "error selecting frontend mapping by frontend_token, name and version")
	}
	return fm, nil
}

func (str *Store) FindFrontendMappingsByFrontendTokenAndName(frontendToken, name string, trx *sqlx.Tx) ([]*FrontendMapping, error) {
	rows, err := trx.Queryx("select * from frontend_mappings where frontend_token = $1 and name = $2 order by version desc", frontendToken, name)
	if err != nil {
		return nil, errors.Wrap(err, "error selecting frontend mappings by frontend_token and name")
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

func (str *Store) FindLatestFrontendMapping(frontendToken, name string, trx *sqlx.Tx) (*FrontendMapping, error) {
	fm := &FrontendMapping{}
	if err := trx.QueryRowx("select * from frontend_mappings where frontend_token = $1 and name = $2 order by version desc limit 1", frontendToken, name).StructScan(fm); err != nil {
		return nil, errors.Wrap(err, "error selecting latest frontend mapping by frontend_token and name")
	}
	return fm, nil
}

func (str *Store) DeleteFrontendMapping(frontendToken, name string, version int64, trx *sqlx.Tx) error {
	stmt, err := trx.Prepare("delete from frontend_mappings where frontend_token = $1 and name = $2 and version = $3")
	if err != nil {
		return errors.Wrap(err, "error preparing frontend_mappings delete statement")
	}
	if _, err := stmt.Exec(frontendToken, name, version); err != nil {
		return errors.Wrap(err, "error executing frontend_mappings delete statement")
	}
	return nil
}

func (str *Store) FindFrontendMappingsByFrontendToken(frontendToken string, trx *sqlx.Tx) ([]*FrontendMapping, error) {
	rows, err := trx.Queryx("select * from frontend_mappings where frontend_token = $1 order by name, version desc", frontendToken)
	if err != nil {
		return nil, errors.Wrap(err, "error selecting frontend mappings by frontend_token")
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

func (str *Store) DeleteFrontendMappingsByFrontendTokenAndName(frontendToken, name string, trx *sqlx.Tx) error {
	stmt, err := trx.Prepare("delete from frontend_mappings where frontend_token = $1 and name = $2")
	if err != nil {
		return errors.Wrap(err, "error preparing frontend_mappings delete by frontend_token and name statement")
	}
	if _, err := stmt.Exec(frontendToken, name); err != nil {
		return errors.Wrap(err, "error executing frontend_mappings delete by frontend_token and name statement")
	}
	return nil
}

func (str *Store) DeleteFrontendMappingsByFrontendToken(frontendToken string, trx *sqlx.Tx) error {
	stmt, err := trx.Prepare("delete from frontend_mappings where frontend_token = $1")
	if err != nil {
		return errors.Wrap(err, "error preparing frontend_mappings delete by frontend_token statement")
	}
	if _, err := stmt.Exec(frontendToken); err != nil {
		return errors.Wrap(err, "error executing frontend_mappings delete by frontend_token statement")
	}
	return nil
}

func (str *Store) FindFrontendMappingsWithVersionOrHigher(frontendToken, name string, version int64, trx *sqlx.Tx) ([]*FrontendMapping, error) {
	rows, err := trx.Queryx("select * from frontend_mappings where frontend_token = $1 and name = $2 and version >= $3 order by version asc", frontendToken, name, version)
	if err != nil {
		return nil, errors.Wrap(err, "error selecting frontend mappings with version or higher")
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

func (str *Store) FindFrontendMappingsByFrontendTokenWithVersionOrHigher(frontendToken string, version int64, trx *sqlx.Tx) ([]*FrontendMapping, error) {
	rows, err := trx.Queryx("select * from frontend_mappings where frontend_token = $1 and version >= $2 order by name, version asc", frontendToken, version)
	if err != nil {
		return nil, errors.Wrap(err, "error selecting frontend mappings by frontend_token with version or higher")
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