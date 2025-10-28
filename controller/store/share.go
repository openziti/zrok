package store

import (
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type Share struct {
	Model
	EnvironmentId        int
	ZId                  string
	Token                string
	ShareMode            string
	BackendMode          string
	FrontendSelection    *string
	FrontendEndpoint     *string
	BackendProxyEndpoint *string
	PermissionMode       PermissionMode
}

func (str *Store) CreateShare(envId int, shr *Share, trx *sqlx.Tx) (int, error) {
	stmt, err := trx.Unsafe().Prepare("insert into shares (environment_id, z_id, token, share_mode, backend_mode, frontend_selection, frontend_endpoint, backend_proxy_endpoint, permission_mode) values ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing shares insert statement")
	}
	var id int
	if err := stmt.QueryRow(envId, shr.ZId, shr.Token, shr.ShareMode, shr.BackendMode, shr.FrontendSelection, shr.FrontendEndpoint, shr.BackendProxyEndpoint, shr.PermissionMode).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "error executing shares insert statement")
	}
	return id, nil
}

func (str *Store) GetShare(id int, trx *sqlx.Tx) (*Share, error) {
	shr := &Share{}
	if err := trx.Unsafe().QueryRowx("select * from shares where id = $1", id).StructScan(shr); err != nil {
		return nil, errors.Wrap(err, "error selecting share by id")
	}
	return shr, nil
}

func (str *Store) FindAllShares(trx *sqlx.Tx) ([]*Share, error) {
	rows, err := trx.Unsafe().Queryx("select * from shares where not deleted order by id")
	if err != nil {
		return nil, errors.Wrap(err, "error selecting all shares")
	}
	var shrs []*Share
	for rows.Next() {
		shr := &Share{}
		if err := rows.StructScan(shr); err != nil {
			return nil, errors.Wrap(err, "error scanning share")
		}
		shrs = append(shrs, shr)
	}
	return shrs, nil
}

func (str *Store) FindAllSharesForAccount(accountId int, trx *sqlx.Tx) ([]*Share, error) {
	sql := "select shares.* from shares, environments" +
		" where shares.environment_id = environments.id" +
		" and environments.account_id = $1" +
		" and not shares.deleted"
	rows, err := trx.Unsafe().Queryx(sql, accountId)
	if err != nil {
		return nil, errors.Wrap(err, "error selecting all shares for account")
	}
	var shrs []*Share
	for rows.Next() {
		shr := &Share{}
		if err := rows.StructScan(shr); err != nil {
			return nil, errors.Wrap(err, "error scanning share")
		}
		shrs = append(shrs, shr)
	}
	return shrs, nil
}

func (str *Store) FindShareWithToken(shrToken string, trx *sqlx.Tx) (*Share, error) {
	shr := &Share{}
	if err := trx.Unsafe().QueryRowx("select * from shares where token = $1 and not deleted", shrToken).StructScan(shr); err != nil {
		return nil, errors.Wrap(err, "error selecting share by token")
	}
	return shr, nil
}

func (str *Store) FindShareWithTokenEvenIfDeleted(shrToken string, trx *sqlx.Tx) (*Share, error) {
	shr := &Share{}
	if err := trx.Unsafe().QueryRowx("select * from shares where token = $1", shrToken).StructScan(shr); err != nil {
		return nil, errors.Wrap(err, "error selecting share by token, even if deleted")
	}
	return shr, nil
}

func (str *Store) ShareWithTokenExists(shrToken string, trx *sqlx.Tx) (bool, error) {
	count := 0
	if err := trx.Unsafe().QueryRowx("select count(0) from shares where token = $1 and not deleted", shrToken).Scan(&count); err != nil {
		return true, errors.Wrap(err, "error selecting share count by token")
	}
	return count > 0, nil
}

func (str *Store) FindShareWithZIdAndDeleted(zId string, trx *sqlx.Tx) (*Share, error) {
	shr := &Share{}
	if err := trx.Unsafe().QueryRowx("select * from shares where z_id = $1", zId).StructScan(shr); err != nil {
		return nil, errors.Wrap(err, "error selecting share by z_id")
	}
	return shr, nil
}

func (str *Store) FindSharesForEnvironment(envId int, trx *sqlx.Tx) ([]*Share, error) {
	rows, err := trx.Unsafe().Queryx("select shares.* from shares where environment_id = $1 and not deleted", envId)
	if err != nil {
		return nil, errors.Wrap(err, "error selecting shares by environment id")
	}
	var shrs []*Share
	for rows.Next() {
		shr := &Share{}
		if err := rows.StructScan(shr); err != nil {
			return nil, errors.Wrap(err, "error scanning share")
		}
		shrs = append(shrs, shr)
	}
	return shrs, nil
}

func (str *Store) UpdateShare(shr *Share, trx *sqlx.Tx) error {
	sql := "update shares set z_id = $1, token = $2, share_mode = $3, backend_mode = $4, frontend_selection = $5, frontend_endpoint = $6, backend_proxy_endpoint = $7, permission_mode = $8, updated_at = current_timestamp where id = $9"
	stmt, err := trx.Unsafe().Prepare(sql)
	if err != nil {
		return errors.Wrap(err, "error preparing shares update statement")
	}
	_, err = stmt.Exec(shr.ZId, shr.Token, shr.ShareMode, shr.BackendMode, shr.FrontendSelection, shr.FrontendEndpoint, shr.BackendProxyEndpoint, shr.PermissionMode, shr.Id)
	if err != nil {
		return errors.Wrap(err, "error executing shares update statement")
	}
	return nil
}

func (str *Store) DeleteShare(id int, trx *sqlx.Tx) error {
	stmt, err := trx.Unsafe().Prepare("update shares set updated_at = current_timestamp, deleted = true where id = $1")
	if err != nil {
		return errors.Wrap(err, "error preparing shares delete statement")
	}
	_, err = stmt.Exec(id)
	if err != nil {
		return errors.Wrap(err, "error executing shares delete statement")
	}
	return nil
}

type ShareFilter struct {
	EnvZId         *string
	ShareMode      *string
	BackendMode    *string
	ShareToken     *string
	Target         *string
	PermissionMode *string
	CreatedAfter   *time.Time
	CreatedBefore  *time.Time
	UpdatedAfter   *time.Time
	UpdatedBefore  *time.Time
}

func (str *Store) FindSharesForAccountWithFilter(accountId int, filter *ShareFilter, trx *sqlx.Tx) ([]*Share, error) {
	query := `
		select shares.*
		from shares
		inner join environments on shares.environment_id = environments.id
		where environments.account_id = $1
		and not shares.deleted
		and not environments.deleted
	`

	args := []interface{}{accountId}
	argIndex := 2

	// text filters
	if filter.EnvZId != nil && *filter.EnvZId != "" {
		query += fmt.Sprintf(" and environments.z_id = $%d", argIndex)
		args = append(args, *filter.EnvZId)
		argIndex++
	}

	if filter.ShareMode != nil && *filter.ShareMode != "" {
		query += fmt.Sprintf(" and shares.share_mode = $%d", argIndex)
		args = append(args, *filter.ShareMode)
		argIndex++
	}

	if filter.BackendMode != nil && *filter.BackendMode != "" {
		query += fmt.Sprintf(" and shares.backend_mode = $%d", argIndex)
		args = append(args, *filter.BackendMode)
		argIndex++
	}

	if filter.ShareToken != nil && *filter.ShareToken != "" {
		query += fmt.Sprintf(" and lower(shares.token) like $%d", argIndex)
		args = append(args, "%"+strings.ToLower(*filter.ShareToken)+"%")
		argIndex++
	}

	if filter.Target != nil && *filter.Target != "" {
		query += fmt.Sprintf(" and lower(shares.backend_proxy_endpoint) like $%d", argIndex)
		args = append(args, "%"+strings.ToLower(*filter.Target)+"%")
		argIndex++
	}

	if filter.PermissionMode != nil && *filter.PermissionMode != "" {
		query += fmt.Sprintf(" and shares.permission_mode = $%d", argIndex)
		args = append(args, *filter.PermissionMode)
		argIndex++
	}

	// date filters
	if filter.CreatedAfter != nil {
		query += fmt.Sprintf(" and shares.created_at >= $%d", argIndex)
		args = append(args, *filter.CreatedAfter)
		argIndex++
	}

	if filter.CreatedBefore != nil {
		query += fmt.Sprintf(" and shares.created_at <= $%d", argIndex)
		args = append(args, *filter.CreatedBefore)
		argIndex++
	}

	if filter.UpdatedAfter != nil {
		query += fmt.Sprintf(" and shares.updated_at >= $%d", argIndex)
		args = append(args, *filter.UpdatedAfter)
		argIndex++
	}

	if filter.UpdatedBefore != nil {
		query += fmt.Sprintf(" and shares.updated_at <= $%d", argIndex)
		args = append(args, *filter.UpdatedBefore)
		argIndex++
	}

	rows, err := trx.Unsafe().Queryx(query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "error selecting shares with filter")
	}
	defer rows.Close()

	var results []*Share
	for rows.Next() {
		shr := &Share{}
		if err := rows.StructScan(shr); err != nil {
			return nil, errors.Wrap(err, "error scanning share")
		}
		results = append(results, shr)
	}

	return results, nil
}
