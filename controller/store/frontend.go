package store

import (
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type Frontend struct {
	Model
	EnvironmentId  *int
	PrivateShareId *int
	Token          string
	ZId            string
	PublicName     *string
	UrlTemplate    *string
	Dynamic        bool
	BindAddress    *string
	Reserved       bool
	PermissionMode PermissionMode
	Description    *string
}

func (str *Store) CreateFrontend(envId int, f *Frontend, trx *sqlx.Tx) (int, error) {
	stmt, err := trx.Prepare("insert into frontends (environment_id, private_share_id, token, z_id, public_name, url_template, dynamic, bind_address, reserved, permission_mode, description) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) returning id")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing frontends insert statement")
	}
	var id int
	if err := stmt.QueryRow(envId, f.PrivateShareId, f.Token, f.ZId, f.PublicName, f.UrlTemplate, f.Dynamic, f.BindAddress, f.Reserved, f.PermissionMode, f.Description).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "error executing frontends insert statement")
	}
	return id, nil
}

func (str *Store) CreateGlobalFrontend(f *Frontend, trx *sqlx.Tx) (int, error) {
	stmt, err := trx.Prepare("insert into frontends (token, z_id, public_name, url_template, dynamic, reserved, permission_mode, description) values ($1, $2, $3, $4, $5, $6, $7, $8) returning id")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing global frontends insert statement")
	}
	var id int
	if err := stmt.QueryRow(f.Token, f.ZId, f.PublicName, f.UrlTemplate, f.Dynamic, f.Reserved, f.PermissionMode, f.Description).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "error executing global frontends insert statement")
	}
	return id, nil
}

func (str *Store) GetFrontend(id int, trx *sqlx.Tx) (*Frontend, error) {
	i := &Frontend{}
	if err := trx.QueryRowx("select * from frontends where id = $1", id).StructScan(i); err != nil {
		return nil, errors.Wrap(err, "error selecting frontend by id")
	}
	return i, nil
}

func (str *Store) FindFrontendWithToken(token string, trx *sqlx.Tx) (*Frontend, error) {
	i := &Frontend{}
	if err := trx.QueryRowx("select frontends.* from frontends where token = $1 and not deleted", token).StructScan(i); err != nil {
		return nil, errors.Wrap(err, "error selecting frontend by name")
	}
	return i, nil
}

func (str *Store) FindFrontendWithTokenAndEnvironment(token string, envId int, trx *sqlx.Tx) (*Frontend, error) {
	i := &Frontend{}
	if err := trx.QueryRowx("select frontends.* from frontends where token = $1 and environment_id = $2 and not deleted", token, envId).StructScan(i); err != nil {
		return nil, errors.Wrap(err, "error selecting frontend by token and environment")
	}
	return i, nil
}

func (str *Store) FindFrontendWithZId(zId string, trx *sqlx.Tx) (*Frontend, error) {
	i := &Frontend{}
	if err := trx.QueryRowx("select frontends.* from frontends where z_id = $1 and not deleted", zId).StructScan(i); err != nil {
		return nil, errors.Wrap(err, "error selecting frontend by ziti id")
	}
	return i, nil
}

func (str *Store) FindFrontendPubliclyNamed(publicName string, trx *sqlx.Tx) (*Frontend, error) {
	i := &Frontend{}
	if err := trx.QueryRowx("select frontends.* from frontends where public_name = $1 and not deleted", publicName).StructScan(i); err != nil {
		return nil, errors.Wrap(err, "error selecting frontend by public_name")
	}
	return i, nil
}

func (str *Store) FindFrontendsForEnvironment(envId int, trx *sqlx.Tx) ([]*Frontend, error) {
	rows, err := trx.Queryx("select frontends.* from frontends where environment_id = $1 and not deleted", envId)
	if err != nil {
		return nil, errors.Wrap(err, "error selecting frontends by environment_id")
	}
	var is []*Frontend
	for rows.Next() {
		i := &Frontend{}
		if err := rows.StructScan(i); err != nil {
			return nil, errors.Wrap(err, "error scanning frontend")
		}
		is = append(is, i)
	}
	return is, nil
}

func (str *Store) FindPublicFrontends(trx *sqlx.Tx) ([]*Frontend, error) {
	rows, err := trx.Queryx("select frontends.* from frontends where environment_id is null and reserved = true and not deleted")
	if err != nil {
		return nil, errors.Wrap(err, "error selecting public frontends")
	}
	var frontends []*Frontend
	for rows.Next() {
		frontend := &Frontend{}
		if err := rows.StructScan(frontend); err != nil {
			return nil, errors.Wrap(err, "error scanning frontend")
		}
		frontends = append(frontends, frontend)
	}
	return frontends, nil
}

func (str *Store) FindOpenPublicFrontends(trx *sqlx.Tx) ([]*Frontend, error) {
	rows, err := trx.Queryx("select frontends.* from frontends where environment_id is null and permission_mode = 'open' and reserved = true and not deleted")
	if err != nil {
		return nil, errors.Wrap(err, "error selecting open public frontends")
	}
	var frontends []*Frontend
	for rows.Next() {
		frontend := &Frontend{}
		if err := rows.StructScan(frontend); err != nil {
			return nil, errors.Wrap(err, "error scanning frontend")
		}
		frontends = append(frontends, frontend)
	}
	return frontends, nil
}

func (str *Store) FindClosedPublicFrontendsGrantedToAccount(accountId int, trx *sqlx.Tx) ([]*Frontend, error) {
	rows, err := trx.Queryx(`
		select frontends.* from frontends 
		inner join frontend_grants on frontends.id = frontend_grants.frontend_id 
		where frontend_grants.account_id = $1 
		and frontends.environment_id is null 
		and frontends.permission_mode = 'closed' 
		and frontends.reserved = true 
		and not frontends.deleted 
		and not frontend_grants.deleted`, accountId)
	if err != nil {
		return nil, errors.Wrap(err, "error selecting closed public frontends granted to account")
	}
	var frontends []*Frontend
	for rows.Next() {
		frontend := &Frontend{}
		if err := rows.StructScan(frontend); err != nil {
			return nil, errors.Wrap(err, "error scanning frontend")
		}
		frontends = append(frontends, frontend)
	}
	return frontends, nil
}

func (str *Store) FindFrontendsForPrivateShare(shrId int, trx *sqlx.Tx) ([]*Frontend, error) {
	rows, err := trx.Queryx("select frontends.* from frontends where private_share_id = $1 and not deleted", shrId)
	if err != nil {
		return nil, errors.Wrap(err, "error selecting frontends by private_share_id")
	}
	var is []*Frontend
	for rows.Next() {
		i := &Frontend{}
		if err := rows.StructScan(i); err != nil {
			return nil, errors.Wrap(err, "error scanning frontend")
		}
		is = append(is, i)
	}
	return is, nil
}

func (str *Store) UpdateFrontend(fe *Frontend, trx *sqlx.Tx) error {
	sql := "update frontends set environment_id = $1, private_share_id = $2, token = $3, z_id = $4, public_name = $5, url_template = $6, dynamic = $7, bind_address = $8, reserved = $9, permission_mode = $10, description = $11, updated_at = current_timestamp where id = $12"
	stmt, err := trx.Prepare(sql)
	if err != nil {
		return errors.Wrap(err, "error preparing frontends update statement")
	}
	_, err = stmt.Exec(fe.EnvironmentId, fe.PrivateShareId, fe.Token, fe.ZId, fe.PublicName, fe.UrlTemplate, fe.Dynamic, fe.BindAddress, fe.Reserved, fe.PermissionMode, fe.Description, fe.Id)
	if err != nil {
		return errors.Wrap(err, "error executing frontends update statement")
	}
	return nil
}

func (str *Store) DeleteFrontend(id int, trx *sqlx.Tx) error {
	stmt, err := trx.Prepare("update frontends set updated_at = current_timestamp, deleted = true where id = $1")
	if err != nil {
		return errors.Wrap(err, "error preparing frontends delete statement")
	}
	_, err = stmt.Exec(id)
	if err != nil {
		return errors.Wrap(err, "error executing frontends delete statement")
	}
	return nil
}

type FrontendFilter struct {
	EnvZId        *string
	ShareToken    *string
	BindAddress   *string
	Description   *string
	Reserved      *bool
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
	UpdatedAfter  *time.Time
	UpdatedBefore *time.Time
}

type FrontendWithEnvironment struct {
	Frontend
	EnvZId      *string
	ShareToken  *string
	BackendMode *string
}

func (str *Store) FindFrontendsForAccountWithFilter(accountId int, filter *FrontendFilter, trx *sqlx.Tx) ([]*FrontendWithEnvironment, error) {
	query := `
		select
			frontends.*,
			environments.z_id as env_z_id,
			shares.token as share_token,
			shares.backend_mode as backend_mode
		from frontends
		left join environments on frontends.environment_id = environments.id
		left join shares on frontends.private_share_id = shares.id
		where environments.account_id = $1
		and not frontends.deleted
		and frontends.public_name is null
		and (environments.deleted is null or not environments.deleted)
		and (shares.deleted is null or not shares.deleted)
	`

	args := []interface{}{accountId}
	argIndex := 2

	// text filters
	if filter.EnvZId != nil && *filter.EnvZId != "" {
		query += fmt.Sprintf(" and environments.z_id = $%d", argIndex)
		args = append(args, *filter.EnvZId)
		argIndex++
	}

	if filter.ShareToken != nil && *filter.ShareToken != "" {
		query += fmt.Sprintf(" and shares.token = $%d", argIndex)
		args = append(args, *filter.ShareToken)
		argIndex++
	}

	if filter.BindAddress != nil && *filter.BindAddress != "" {
		query += fmt.Sprintf(" and lower(frontends.bind_address) like $%d", argIndex)
		args = append(args, "%"+strings.ToLower(*filter.BindAddress)+"%")
		argIndex++
	}

	if filter.Description != nil && *filter.Description != "" {
		query += fmt.Sprintf(" and lower(frontends.description) like $%d", argIndex)
		args = append(args, "%"+strings.ToLower(*filter.Description)+"%")
		argIndex++
	}

	// boolean filter
	if filter.Reserved != nil {
		query += fmt.Sprintf(" and frontends.reserved = $%d", argIndex)
		args = append(args, *filter.Reserved)
		argIndex++
	}

	// date filters
	if filter.CreatedAfter != nil {
		query += fmt.Sprintf(" and frontends.created_at >= $%d", argIndex)
		args = append(args, *filter.CreatedAfter)
		argIndex++
	}

	if filter.CreatedBefore != nil {
		query += fmt.Sprintf(" and frontends.created_at <= $%d", argIndex)
		args = append(args, *filter.CreatedBefore)
		argIndex++
	}

	if filter.UpdatedAfter != nil {
		query += fmt.Sprintf(" and frontends.updated_at >= $%d", argIndex)
		args = append(args, *filter.UpdatedAfter)
		argIndex++
	}

	if filter.UpdatedBefore != nil {
		query += fmt.Sprintf(" and frontends.updated_at <= $%d", argIndex)
		args = append(args, *filter.UpdatedBefore)
		argIndex++
	}

	rows, err := trx.Queryx(query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "error selecting frontends with filter")
	}
	defer rows.Close()

	var results []*FrontendWithEnvironment
	for rows.Next() {
		result := &FrontendWithEnvironment{}
		// use MapScan to get all columns including the joined ones
		cols := make(map[string]interface{})
		if err := rows.MapScan(cols); err != nil {
			return nil, errors.Wrap(err, "error scanning frontend row")
		}

		// manually populate Frontend fields from the map
		if id, ok := cols["id"].(int64); ok {
			result.Frontend.Id = int(id)
		}
		if envId, ok := cols["environment_id"].(int64); ok {
			intVal := int(envId)
			result.Frontend.EnvironmentId = &intVal
		}
		if privShareId, ok := cols["private_share_id"].(int64); ok {
			intVal := int(privShareId)
			result.Frontend.PrivateShareId = &intVal
		}
		if token, ok := cols["token"].(string); ok {
			result.Frontend.Token = token
		}
		if zId, ok := cols["z_id"].(string); ok {
			result.Frontend.ZId = zId
		}
		if publicName, ok := cols["public_name"].(string); ok {
			result.Frontend.PublicName = &publicName
		}
		if urlTemplate, ok := cols["url_template"].(string); ok {
			result.Frontend.UrlTemplate = &urlTemplate
		}
		if dynamic, ok := cols["dynamic"].(bool); ok {
			result.Frontend.Dynamic = dynamic
		}
		if bindAddr, ok := cols["bind_address"].(string); ok {
			result.Frontend.BindAddress = &bindAddr
		}
		if reserved, ok := cols["reserved"].(bool); ok {
			result.Frontend.Reserved = reserved
		}
		if permMode, ok := cols["permission_mode"].(string); ok {
			result.Frontend.PermissionMode = PermissionMode(permMode)
		}
		if desc, ok := cols["description"].(string); ok {
			result.Frontend.Description = &desc
		}
		if createdAt, ok := cols["created_at"].(time.Time); ok {
			result.Frontend.CreatedAt = createdAt
		}
		if updatedAt, ok := cols["updated_at"].(time.Time); ok {
			result.Frontend.UpdatedAt = updatedAt
		}
		if deleted, ok := cols["deleted"].(bool); ok {
			result.Frontend.Deleted = deleted
		}

		// extract the additional joined fields
		if envZId, ok := cols["env_z_id"].(string); ok {
			result.EnvZId = &envZId
		}
		if shareToken, ok := cols["share_token"].(string); ok {
			result.ShareToken = &shareToken
		}
		if backendMode, ok := cols["backend_mode"].(string); ok {
			result.BackendMode = &backendMode
		}

		results = append(results, result)
	}

	return results, nil
}

// AccessFrontendRepairDetail is a live access frontend whose share is deleted, with the share's token.
type AccessFrontendRepairDetail struct {
	Id            int    `db:"id"`
	Token         string `db:"token"`
	EnvironmentId *int   `db:"environment_id"`
	ShareToken    string `db:"share_token"`
}

// a live access frontend whose private share has no live row.
const accessFrontendsOfDeletedSharesWhere = `where not f.deleted
	  and f.private_share_id is not null
	  and not exists (select 1 from shares s where s.id = f.private_share_id and not s.deleted)`

// CountAccessFrontendsOfDeletedShares counts the live access frontends whose share is deleted.
func (str *Store) CountAccessFrontendsOfDeletedShares(trx *sqlx.Tx) (int, error) {
	var count int
	if err := trx.QueryRow("select count(*) from frontends f " + accessFrontendsOfDeletedSharesWhere).Scan(&count); err != nil {
		return 0, errors.Wrap(err, "error counting access frontends of deleted shares")
	}
	return count, nil
}

// FindAccessFrontendsOfDeletedShares lists, lowest id first, up to limit live access frontends whose share
// is deleted.
func (str *Store) FindAccessFrontendsOfDeletedShares(limit int, trx *sqlx.Tx) ([]*AccessFrontendRepairDetail, error) {
	sql := `select f.id, f.token, f.environment_id, coalesce(s.token, '') as share_token
	        from frontends f
	        left join shares s on f.private_share_id = s.id
	        ` + accessFrontendsOfDeletedSharesWhere + ` order by f.id limit $1`
	rows, err := trx.Queryx(sql, limit)
	if err != nil {
		return nil, errors.Wrap(err, "error finding access frontends of deleted shares")
	}
	defer func() { _ = rows.Close() }()
	var fes []*AccessFrontendRepairDetail
	for rows.Next() {
		fe := &AccessFrontendRepairDetail{}
		if err := rows.StructScan(fe); err != nil {
			return nil, errors.Wrap(err, "error scanning access frontend of deleted share")
		}
		fes = append(fes, fe)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating access frontends of deleted shares")
	}
	return fes, nil
}

// DeleteFrontends soft-deletes the live frontends among ids and returns how many it deleted.
func (str *Store) DeleteFrontends(ids []int, trx *sqlx.Tx) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	query, args, err := sqlx.In("update frontends set updated_at = current_timestamp, deleted = true where id in (?) and not deleted", ids)
	if err != nil {
		return 0, errors.Wrap(err, "error building frontends delete statement")
	}
	res, err := trx.Exec(trx.Rebind(query), args...)
	if err != nil {
		return 0, errors.Wrap(err, "error executing frontends delete statement")
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "error reading frontends deleted")
	}
	return n, nil
}
