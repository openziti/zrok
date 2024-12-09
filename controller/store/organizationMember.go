package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

func (str *Store) AddAccountToOrganization(acctId, orgId int, admin bool, trx *sqlx.Tx) error {
	stmt, err := trx.Prepare("insert into organization_members (account_id, organization_id, admin) values ($1, $2, $3)")
	if err != nil {
		return errors.Wrap(err, "error preparing organization_members insert statement")
	}
	_, err = stmt.Exec(acctId, orgId, admin)
	if err != nil {
		return errors.Wrap(err, "error executing organization_members insert statement")
	}
	return nil
}

type OrganizationMember struct {
	Email string
	Admin bool
}

func (str *Store) FindAccountsForOrganization(orgId int, trx *sqlx.Tx) ([]*OrganizationMember, error) {
	rows, err := trx.Queryx("select organization_members.admin, accounts.email from organization_members, accounts where organization_members.organization_id = $1 and organization_members.account_id = accounts.id", orgId)
	if err != nil {
		return nil, errors.Wrap(err, "error querying organization members")
	}
	var members []*OrganizationMember
	for rows.Next() {
		om := &OrganizationMember{}
		if err := rows.StructScan(&om); err != nil {
			return nil, errors.Wrap(err, "error scanning account email")
		}
		members = append(members, om)
	}
	return members, nil
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
	stmt, err := trx.Prepare("delete from organization_members where account_id = $1 and organization_id = $2")
	if err != nil {
		return errors.Wrap(err, "error preparing organization_members delete statement")
	}
	_, err = stmt.Exec(acctId, orgId)
	if err != nil {
		return errors.Wrap(err, "error executing organization_members delete statement")
	}
	return nil
}
