package store

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type FrontendMapping struct {
	Id            int64
	FrontendToken string
	Name          string
	ShareToken    string
	CreatedAt     time.Time
}

type FrontendMappingWithShareState struct {
	FrontendMapping
	ShareDeleted *bool `db:"share_deleted"`
}

func (str *Store) CreateFrontendMapping(fm *FrontendMapping, trx *sqlx.Tx) (int, error) {
	stmt, err := trx.Prepare("insert into frontend_mappings (frontend_token, name, share_token) values ($1, $2, $3) returning id")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing frontend_mappings insert statement")
	}
	var id int
	if err := stmt.QueryRow(fm.FrontendToken, fm.Name, fm.ShareToken).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "error executing frontend_mappings insert statement")
	}
	return id, nil
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

func (str *Store) FindFrontendMappingByFrontendTokenAndNameWithShareState(frontendToken, name string, trx *sqlx.Tx) (*FrontendMappingWithShareState, error) {
	sql := `select fm.id, fm.frontend_token, fm.name, fm.share_token, fm.created_at,
	               s.deleted as share_deleted
	        from frontend_mappings fm
	        left join shares s on fm.share_token = s.token
	        where fm.frontend_token = $1 and fm.name = $2`

	mapping := &FrontendMappingWithShareState{}
	if err := trx.QueryRowx(sql, frontendToken, name).StructScan(mapping); err != nil {
		return nil, errors.Wrap(err, "error selecting frontend mapping by frontend token and name with share state")
	}
	return mapping, nil
}

func (str *Store) FindFrontendMappingsWithHigherId(frontendToken, name string, id int64, trx *sqlx.Tx) ([]*FrontendMapping, error) {
	rows, err := trx.Queryx("select * from frontend_mappings where frontend_token = $1 and name = $2 and id > $3 order by id asc", frontendToken, name, id)
	if err != nil {
		return nil, errors.Wrap(err, "error selecting frontend mappings with id or higher")
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

func (str *Store) FindFrontendMappingsByFrontendTokenWithHigherId(frontendToken string, id int64, trx *sqlx.Tx) ([]*FrontendMapping, error) {
	rows, err := trx.Queryx("select * from frontend_mappings where frontend_token = $1 and id > $2 order by name, id asc", frontendToken, id)
	if err != nil {
		return nil, errors.Wrap(err, "error selecting frontend mappings by frontend_token with id or higher")
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
