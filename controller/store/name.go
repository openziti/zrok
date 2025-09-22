package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type Name struct {
	Model
	NamespaceId int
	Name        string
	AccountId   int
	Reserved    bool
}

type NameWithShareToken struct {
	Name
	ShareToken *string
}

type NameWithNamespace struct {
	Name
	NamespaceName string
}

func (str *Store) CreateName(an *Name, trx *sqlx.Tx) (int, error) {
	stmt, err := trx.Prepare("insert into names (namespace_id, name, account_id, reserved) values ($1, $2, $3, $4) returning id")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing name insert statement")
	}
	var id int
	if err := stmt.QueryRow(an.NamespaceId, an.Name, an.AccountId, an.Reserved).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "error executing name insert statement")
	}
	return id, nil
}

func (str *Store) GetName(id int, trx *sqlx.Tx) (*Name, error) {
	an := &Name{}
	if err := trx.QueryRowx("select * from names where id = $1 and not deleted", id).StructScan(an); err != nil {
		return nil, errors.Wrap(err, "error selecting name by id")
	}
	return an, nil
}

func (str *Store) FindNameByNamespaceAndName(namespaceId int, name string, trx *sqlx.Tx) (*Name, error) {
	an := &Name{}
	if err := trx.QueryRowx("select * from names where namespace_id = $1 and name = $2 and not deleted", namespaceId, name).StructScan(an); err != nil {
		return nil, errors.Wrap(err, "error selecting name by namespace and name")
	}
	return an, nil
}

func (str *Store) FindNamesForNamespace(namespaceId int, trx *sqlx.Tx) ([]*Name, error) {
	rows, err := trx.Queryx("select * from names where namespace_id = $1 and not deleted order by name", namespaceId)
	if err != nil {
		return nil, errors.Wrap(err, "error finding names for namespace")
	}
	var names []*Name
	for rows.Next() {
		an := &Name{}
		if err := rows.StructScan(&an); err != nil {
			return nil, errors.Wrap(err, "error scanning name")
		}
		names = append(names, an)
	}
	return names, nil
}

func (str *Store) FindNamesForAccount(accountId int, trx *sqlx.Tx) ([]*Name, error) {
	rows, err := trx.Queryx("select * from names where account_id = $1 and not deleted order by name", accountId)
	if err != nil {
		return nil, errors.Wrap(err, "error finding names for account")
	}
	var names []*Name
	for rows.Next() {
		an := &Name{}
		if err := rows.StructScan(&an); err != nil {
			return nil, errors.Wrap(err, "error scanning name")
		}
		names = append(names, an)
	}
	return names, nil
}

func (str *Store) FindNamesForAccountAndNamespace(accountId, namespaceId int, trx *sqlx.Tx) ([]*Name, error) {
	rows, err := trx.Queryx("select * from names where account_id = $1 and namespace_id = $2 and not deleted order by name", accountId, namespaceId)
	if err != nil {
		return nil, errors.Wrap(err, "error finding names for account and namespace")
	}
	var names []*Name
	for rows.Next() {
		an := &Name{}
		if err := rows.StructScan(&an); err != nil {
			return nil, errors.Wrap(err, "error scanning name")
		}
		names = append(names, an)
	}
	return names, nil
}

func (str *Store) CheckNameAvailability(namespaceId int, name string, trx *sqlx.Tx) (bool, error) {
	var count int
	if err := trx.QueryRow("select count(*) from names where namespace_id = $1 and name = $2 and not deleted", namespaceId, name).Scan(&count); err != nil {
		return false, errors.Wrap(err, "error checking name availability")
	}
	return count == 0, nil
}

func (str *Store) FindNamesWithShareTokensForAccountAndNamespace(accountId, namespaceId int, trx *sqlx.Tx) ([]*NameWithShareToken, error) {
	sql := `select n.id, n.created_at, n.updated_at, n.deleted, n.namespace_id, n.name, n.account_id, n.reserved, s.token as share_token
			from names n
			left join share_name_mappings snm on n.id = snm.name_id and not snm.deleted
			left join shares s on snm.share_id = s.id and not s.deleted
			where n.account_id = $1 and n.namespace_id = $2 and not n.deleted
			order by n.name`

	rows, err := trx.Queryx(sql, accountId, namespaceId)
	if err != nil {
		return nil, errors.Wrap(err, "error finding names with share tokens for account and namespace")
	}

	var names []*NameWithShareToken
	for rows.Next() {
		nwst := &NameWithShareToken{}
		if err := rows.Scan(&nwst.Name.Id, &nwst.Name.CreatedAt, &nwst.Name.UpdatedAt, &nwst.Name.Deleted, &nwst.Name.NamespaceId, &nwst.Name.Name, &nwst.Name.AccountId, &nwst.Name.Reserved, &nwst.ShareToken); err != nil {
			return nil, errors.Wrap(err, "error scanning name with share token")
		}
		names = append(names, nwst)
	}

	return names, nil
}

func (str *Store) FindNamesForShare(shareId int, trx *sqlx.Tx) ([]*NameWithNamespace, error) {
	sql := `select n.id, n.created_at, n.updated_at, n.deleted, n.namespace_id,
	               n.name, n.account_id, n.reserved,
	               ns.name as namespace_name
	        from share_name_mappings snm
	        join names n on snm.name_id = n.id
	        join namespaces ns on n.namespace_id = ns.id
	        where snm.share_id = $1
	          and not snm.deleted
	          and not n.deleted
	          and not ns.deleted
	        order by ns.name, n.name`

	rows, err := trx.Queryx(sql, shareId)
	if err != nil {
		return nil, errors.Wrap(err, "error finding names for share")
	}

	var names []*NameWithNamespace
	for rows.Next() {
		nwn := &NameWithNamespace{}
		if err := rows.Scan(&nwn.Name.Id, &nwn.Name.CreatedAt, &nwn.Name.UpdatedAt,
			&nwn.Name.Deleted, &nwn.Name.NamespaceId, &nwn.Name.Name,
			&nwn.Name.AccountId, &nwn.Name.Reserved, &nwn.NamespaceName); err != nil {
			return nil, errors.Wrap(err, "error scanning name with namespace")
		}
		names = append(names, nwn)
	}

	return names, nil
}

func (str *Store) UpdateName(name *Name, trx *sqlx.Tx) error {
	stmt, err := trx.Prepare("update names set updated_at = current_timestamp, reserved = $1 where id = $2")
	if err != nil {
		return errors.Wrap(err, "error preparing name update statement")
	}
	_, err = stmt.Exec(name.Reserved, name.Id)
	if err != nil {
		return errors.Wrap(err, "error executing name update statement")
	}
	return nil
}

func (str *Store) DeleteName(id int, trx *sqlx.Tx) error {
	stmt, err := trx.Prepare("update names set updated_at = current_timestamp, deleted = true where id = $1")
	if err != nil {
		return errors.Wrap(err, "error preparing name delete statement")
	}
	_, err = stmt.Exec(id)
	if err != nil {
		return errors.Wrap(err, "error executing name delete statement")
	}
	return nil
}
