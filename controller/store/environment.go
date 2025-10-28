package store

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type Environment struct {
	Model
	AccountId   *int
	Description string
	Host        string
	Address     string
	ZId         string
	Deleted     bool
}

func (str *Store) CreateEnvironment(accountId int, i *Environment, trx *sqlx.Tx) (int, error) {
	stmt, err := trx.Prepare("insert into environments (account_id, description, host, address, z_id) values ($1, $2, $3, $4, $5) returning id")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing environments insert statement")
	}
	var id int
	if err := stmt.QueryRow(accountId, i.Description, i.Host, i.Address, i.ZId).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "error executing environments insert statement")
	}
	return id, nil
}

func (str *Store) CreateEphemeralEnvironment(i *Environment, trx *sqlx.Tx) (int, error) {
	stmt, err := trx.Prepare("insert into environments (description, host, address, z_id) values ($1, $2, $3, $4) returning id")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing environments (ephemeral) insert statement")
	}
	var id int
	if err := stmt.QueryRow(i.Description, i.Host, i.Address, i.ZId).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "error executing environments (ephemeral) insert statement")
	}
	return id, nil
}

func (str *Store) GetEnvironment(id int, trx *sqlx.Tx) (*Environment, error) {
	i := &Environment{}
	if err := trx.QueryRowx("select * from environments where id = $1", id).StructScan(i); err != nil {
		return nil, errors.Wrap(err, "error selecting environment by id")
	}
	return i, nil
}

func (str *Store) FindEnvironmentsForAccount(accountId int, trx *sqlx.Tx) ([]*Environment, error) {
	rows, err := trx.Queryx("select environments.* from environments where account_id = $1 and not deleted", accountId)
	if err != nil {
		return nil, errors.Wrap(err, "error selecting environments by account id")
	}
	var is []*Environment
	for rows.Next() {
		i := &Environment{}
		if err := rows.StructScan(i); err != nil {
			return nil, errors.Wrap(err, "error scanning environment")
		}
		is = append(is, i)
	}
	return is, nil
}

func (str *Store) FindEnvironmentForAccount(envZId string, accountId int, trx *sqlx.Tx) (*Environment, error) {
	env := &Environment{}
	if err := trx.QueryRowx("select environments.* from environments where z_id = $1 and account_id = $2 and not deleted", envZId, accountId).StructScan(env); err != nil {
		return nil, errors.Wrap(err, "error finding environment by z_id and account_id")
	}
	return env, nil
}

func (str *Store) DeleteEnvironment(id int, trx *sqlx.Tx) error {
	stmt, err := trx.Prepare("update environments set updated_at = current_timestamp, deleted = true where id = $1")
	if err != nil {
		return errors.Wrap(err, "error preparing environments delete statement")
	}
	_, err = stmt.Exec(id)
	if err != nil {
		return errors.Wrap(err, "error executing environments delete statement")
	}
	return nil
}

type EnvironmentFilter struct {
	Description    *string
	Host           *string
	Address        *string
	ShareCount     *string
	AccessCount    *string
	CreatedAfter   *time.Time
	CreatedBefore  *time.Time
	UpdatedAfter   *time.Time
	UpdatedBefore  *time.Time
	HasShares      *bool
	HasAccesses    *bool
}

type EnvironmentWithCounts struct {
	Environment
	ShareCount  int
	AccessCount int
}

func (str *Store) FindEnvironmentsForAccountWithFilter(accountId int, filter *EnvironmentFilter, trx *sqlx.Tx) ([]*EnvironmentWithCounts, error) {
	query := `
		select
			e.*,
			coalesce(share_counts.count, 0) as share_count,
			coalesce(access_counts.count, 0) as access_count
		from environments e
		left join (
			select environment_id, count(*) as count
			from shares
			where not deleted
			group by environment_id
		) share_counts on e.id = share_counts.environment_id
		left join (
			select environment_id, count(*) as count
			from frontends
			where environment_id is not null and not deleted
			group by environment_id
		) access_counts on e.id = access_counts.environment_id
		where e.account_id = $1 and not e.deleted
	`

	args := []interface{}{accountId}
	argIndex := 2

	// text filters
	if filter.Description != nil && *filter.Description != "" {
		query += fmt.Sprintf(" and lower(e.description) like $%d", argIndex)
		args = append(args, "%"+strings.ToLower(*filter.Description)+"%")
		argIndex++
	}

	if filter.Host != nil && *filter.Host != "" {
		query += fmt.Sprintf(" and lower(e.host) like $%d", argIndex)
		args = append(args, "%"+strings.ToLower(*filter.Host)+"%")
		argIndex++
	}

	if filter.Address != nil && *filter.Address != "" {
		query += fmt.Sprintf(" and e.address = $%d", argIndex)
		args = append(args, *filter.Address)
		argIndex++
	}

	// date filters
	if filter.CreatedAfter != nil {
		query += fmt.Sprintf(" and e.created_at >= $%d", argIndex)
		args = append(args, *filter.CreatedAfter)
		argIndex++
	}

	if filter.CreatedBefore != nil {
		query += fmt.Sprintf(" and e.created_at <= $%d", argIndex)
		args = append(args, *filter.CreatedBefore)
		argIndex++
	}

	if filter.UpdatedAfter != nil {
		query += fmt.Sprintf(" and e.updated_at >= $%d", argIndex)
		args = append(args, *filter.UpdatedAfter)
		argIndex++
	}

	if filter.UpdatedBefore != nil {
		query += fmt.Sprintf(" and e.updated_at <= $%d", argIndex)
		args = append(args, *filter.UpdatedBefore)
		argIndex++
	}

	// boolean filters for shares/accesses
	if filter.HasShares != nil {
		if *filter.HasShares {
			query += " and coalesce(share_counts.count, 0) > 0"
		} else {
			query += " and coalesce(share_counts.count, 0) = 0"
		}
	}

	if filter.HasAccesses != nil {
		if *filter.HasAccesses {
			query += " and coalesce(access_counts.count, 0) > 0"
		} else {
			query += " and coalesce(access_counts.count, 0) = 0"
		}
	}

	// wrap query in a subquery for count filtering
	needsSubquery := filter.ShareCount != nil || filter.AccessCount != nil
	if needsSubquery {
		query = fmt.Sprintf("select * from (%s) as filtered", query)

		if filter.ShareCount != nil && *filter.ShareCount != "" {
			condition, err := parseComparisonFilter(*filter.ShareCount, "share_count")
			if err != nil {
				return nil, errors.Wrap(err, "error parsing shareCount filter")
			}
			query += " where " + condition
		}

		if filter.AccessCount != nil && *filter.AccessCount != "" {
			condition, err := parseComparisonFilter(*filter.AccessCount, "access_count")
			if err != nil {
				return nil, errors.Wrap(err, "error parsing accessCount filter")
			}
			if filter.ShareCount != nil && *filter.ShareCount != "" {
				query += " and " + condition
			} else {
				query += " where " + condition
			}
		}
	}

	rows, err := trx.Queryx(query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "error selecting environments with filter")
	}
	defer rows.Close()

	var results []*EnvironmentWithCounts
	for rows.Next() {
		result := &EnvironmentWithCounts{}
		if err := rows.StructScan(result); err != nil {
			return nil, errors.Wrap(err, "error scanning environment with counts")
		}
		results = append(results, result)
	}

	return results, nil
}

// parseComparisonFilter parses comparison operators like ">0", ">=5", "=10", "<20", "<=15"
func parseComparisonFilter(filter, columnName string) (string, error) {
	filter = strings.TrimSpace(filter)
	if filter == "" {
		return "", errors.New("empty filter")
	}

	// parse operator and value
	var operator string
	var valueStr string

	if strings.HasPrefix(filter, ">=") {
		operator = ">="
		valueStr = strings.TrimSpace(filter[2:])
	} else if strings.HasPrefix(filter, "<=") {
		operator = "<="
		valueStr = strings.TrimSpace(filter[2:])
	} else if strings.HasPrefix(filter, ">") {
		operator = ">"
		valueStr = strings.TrimSpace(filter[1:])
	} else if strings.HasPrefix(filter, "<") {
		operator = "<"
		valueStr = strings.TrimSpace(filter[1:])
	} else if strings.HasPrefix(filter, "=") {
		operator = "="
		valueStr = strings.TrimSpace(filter[1:])
	} else {
		// assume equals if no operator
		operator = "="
		valueStr = filter
	}

	// validate that value is a number
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return "", errors.Wrapf(err, "invalid numeric value: %s", valueStr)
	}

	return fmt.Sprintf("%s %s %d", columnName, operator, value), nil
}
