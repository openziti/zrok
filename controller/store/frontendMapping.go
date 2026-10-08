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

// FindFrontendMappingByFrontendTokenAndNameWithShareState returns the frontend mapping with the state of
// its share: ShareDeleted is false when any live share carries the token, true when only deleted shares
// do, and nil when no share row carries it. a share token is unique only among live shares, so the state
// is decided over every row with the token, never a single joined one.
func (str *Store) FindFrontendMappingByFrontendTokenAndNameWithShareState(frontendToken, name string, trx *sqlx.Tx) (*FrontendMappingWithShareState, error) {
	sql := `select fm.id, fm.frontend_token, fm.name, fm.share_token, fm.created_at,
	               case
	                 when exists (select 1 from shares s where s.token = fm.share_token and not s.deleted) then false
	                 when exists (select 1 from shares s where s.token = fm.share_token) then true
	                 else null
	               end as share_deleted
	        from frontend_mappings fm
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

func (str *Store) FindFrontendMappingsByShareToken(shareToken string, trx *sqlx.Tx) ([]*FrontendMapping, error) {
	rows, err := trx.Queryx("select * from frontend_mappings where share_token = $1 order by id asc", shareToken)
	if err != nil {
		return nil, errors.Wrap(err, "error selecting frontend mappings by share_token")
	}
	defer func() { _ = rows.Close() }()
	var mappings []*FrontendMapping
	for rows.Next() {
		fm := &FrontendMapping{}
		if err := rows.StructScan(fm); err != nil {
			return nil, errors.Wrap(err, "error scanning frontend mapping")
		}
		mappings = append(mappings, fm)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating frontend mappings by share_token")
	}
	return mappings, nil
}

func (str *Store) DeleteFrontendMappingsByShareToken(shareToken string, trx *sqlx.Tx) error {
	stmt, err := trx.Prepare("delete from frontend_mappings where share_token = $1")
	if err != nil {
		return errors.Wrap(err, "error preparing frontend_mappings delete by share_token statement")
	}
	if _, err := stmt.Exec(shareToken); err != nil {
		return errors.Wrap(err, "error executing frontend_mappings delete by share_token statement")
	}
	return nil
}

// FrontendMappingRepairDetail is a frontend mapping found by the store repair sweep. ShareRow is false when
// no share row, live or deleted, carries its share token.
type FrontendMappingRepairDetail struct {
	FrontendMapping
	ShareRow bool `db:"share_row"`
}

// a frontend mapping whose share token belongs to no live share: the share is deleted or never existed.
const frontendMappingsWithoutLiveShareWhere = `where not exists (select 1 from shares s where s.token = fm.share_token and not s.deleted)`

// CountFrontendMappingsWithoutLiveShare counts the frontend mappings whose share token belongs to no live
// share.
func (str *Store) CountFrontendMappingsWithoutLiveShare(trx *sqlx.Tx) (int, error) {
	var count int
	if err := trx.QueryRow("select count(*) from frontend_mappings fm " + frontendMappingsWithoutLiveShareWhere).Scan(&count); err != nil {
		return 0, errors.Wrap(err, "error counting frontend mappings without a live share")
	}
	return count, nil
}

// FindFrontendMappingsWithoutLiveShare lists, lowest id first, up to limit frontend mappings whose share
// token belongs to no live share.
func (str *Store) FindFrontendMappingsWithoutLiveShare(limit int, trx *sqlx.Tx) ([]*FrontendMappingRepairDetail, error) {
	sql := `select fm.id, fm.frontend_token, fm.name, fm.share_token, fm.created_at,
	               exists (select 1 from shares s where s.token = fm.share_token) as share_row
	        from frontend_mappings fm ` + frontendMappingsWithoutLiveShareWhere + ` order by fm.id limit $1`
	rows, err := trx.Queryx(sql, limit)
	if err != nil {
		return nil, errors.Wrap(err, "error finding frontend mappings without a live share")
	}
	defer func() { _ = rows.Close() }()
	var mappings []*FrontendMappingRepairDetail
	for rows.Next() {
		fm := &FrontendMappingRepairDetail{}
		if err := rows.StructScan(fm); err != nil {
			return nil, errors.Wrap(err, "error scanning frontend mapping without a live share")
		}
		mappings = append(mappings, fm)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating frontend mappings without a live share")
	}
	return mappings, nil
}

// DeleteFrontendMappings deletes the frontend mappings among ids and returns how many it deleted.
func (str *Store) DeleteFrontendMappings(ids []int64, trx *sqlx.Tx) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	query, args, err := sqlx.In("delete from frontend_mappings where id in (?)", ids)
	if err != nil {
		return 0, errors.Wrap(err, "error building frontend_mappings delete by id statement")
	}
	res, err := trx.Exec(trx.Rebind(query), args...)
	if err != nil {
		return 0, errors.Wrap(err, "error executing frontend_mappings delete by id statement")
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "error reading frontend_mappings deleted")
	}
	return n, nil
}
