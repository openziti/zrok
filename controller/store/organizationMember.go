package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

func (str *Store) AddAccountToOrganization(acctId, orgId int, trx *sqlx.Tx) error {
	stmt, err := trx.Prepare("insert into organization_members (organization_id, account_id) values ($1, $2)")
	if err != nil {
		return errors.Wrap(err, "error preparing organization_members insert statement")
	}
	_, err = stmt.Exec(acctId, orgId)
	if err != nil {
		return errors.Wrap(err, "error executing organization_members insert statement")
	}
	return nil
}

func (str *Store) IsAccountInOrganization(acctId, orgId int, trx *sqlx.Tx) (bool, error) {
	stmt, err := trx.Prepare("select count(0) from organization_members where organization_id = $1 and account_id = $2")
	if err != nil {
		return false, errors.Wrap(err, "error preparing organization_members count statement")
	}
	var count int
	if err := stmt.QueryRow(acctId, orgId).Scan(&count); err != nil {
		return false, errors.Wrap(err, "error executing organization_members count statement")
	}
	return count > 0, nil
}

func (str *Store) IsAccountAdminOfOrganization(acctId, orgId int, trx *sqlx.Tx) (bool, error) {
	stmt, err := trx.Prepare("select count(0) from organization_members where organization_id = $1 and account_id = $2 and admin")
	if err != nil {
		return false, errors.Wrap(err, "error preparing organization_members count statement")
	}
	var count int
	if err := stmt.QueryRow(acctId, orgId).Scan(&count); err != nil {
		return false, errors.Wrap(err, "error executing organization_members count statement")
	}
	return count > 0, nil
}

func (str *Store) RemoveAccountFromOrganization(acctId, orgId int, trx *sqlx.Tx) error {
	stmt, err := trx.Prepare("delete from organization_members where organization_id = $1 and account_id = $2")
	if err != nil {
		return errors.Wrap(err, "error preparing organization_members delete statement")
	}
	_, err = stmt.Exec(acctId, orgId)
	if err != nil {
		return errors.Wrap(err, "error executing organization_members delete statement")
	}
	return nil
}
