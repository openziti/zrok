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

// DeleteAllocatedNames soft-deletes the live auto-allocated names among ids, leaving reserved names, and
// returns how many it deleted.
func (str *Store) DeleteAllocatedNames(ids []int, trx *sqlx.Tx) (int64, error) {
	return str.deleteNamesWhere("not deleted and not reserved", ids, trx)
}

func (str *Store) deleteNamesWhere(guard string, ids []int, trx *sqlx.Tx) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	query, args, err := sqlx.In("update names set updated_at = current_timestamp, deleted = true where id in (?) and "+guard, ids)
	if err != nil {
		return 0, errors.Wrap(err, "error building names delete statement")
	}
	res, err := trx.Exec(trx.Rebind(query), args...)
	if err != nil {
		return 0, errors.Wrap(err, "error executing names delete statement")
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "error reading names deleted")
	}
	return n, nil
}

// FindShareNameMappingsForAccountNamesWithShare lists the live share name mappings on the account's live
// names, with the state of each mapping's share.
func (str *Store) FindShareNameMappingsForAccountNamesWithShare(accountId int, trx *sqlx.Tx) ([]*ShareNameMappingWithShare, error) {
	sql := `select snm.id, snm.created_at, snm.updated_at, snm.deleted, snm.share_id, snm.name_id,
	               s.token as share_token, s.deleted as share_deleted
	        from share_name_mappings snm
	        join names n on snm.name_id = n.id
	        join shares s on snm.share_id = s.id
	        where n.account_id = $1 and not n.deleted and not snm.deleted
	        order by snm.id`
	rows, err := trx.Queryx(sql, accountId)
	if err != nil {
		return nil, errors.Wrap(err, "error finding share name mappings for account names")
	}
	defer func() { _ = rows.Close() }()
	var mappings []*ShareNameMappingWithShare
	for rows.Next() {
		snm := &ShareNameMappingWithShare{}
		if err := rows.StructScan(snm); err != nil {
			return nil, errors.Wrap(err, "error scanning share name mapping for account names")
		}
		mappings = append(mappings, snm)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating share name mappings for account names")
	}
	return mappings, nil
}

// DeleteNamesForAccount soft-deletes every live name of the account, reserved or not, and returns how many
// it deleted.
func (str *Store) DeleteNamesForAccount(accountId int, trx *sqlx.Tx) (int64, error) {
	res, err := trx.Exec("update names set updated_at = current_timestamp, deleted = true where account_id = $1 and not deleted", accountId)
	if err != nil {
		return 0, errors.Wrap(err, "error executing names delete for account statement")
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "error reading names deleted for account")
	}
	return n, nil
}

// NameRepairDetail is a live name found by the store repair sweep.
type NameRepairDetail struct {
	Id             int    `db:"id"`
	Name           string `db:"name"`
	Reserved       bool   `db:"reserved"`
	AccountId      int    `db:"account_id"`
	NamespaceToken string `db:"namespace_token"`
}

// a live name whose account is deleted.
const namesOfDeletedAccountsWhere = `where not n.deleted
	  and exists (select 1 from accounts a where a.id = n.account_id and a.deleted)`

// a live auto-allocated name that no live share name mapping holds.
const unmappedAllocatedNamesWhere = `where not n.deleted
	  and not n.reserved
	  and not exists (select 1 from share_name_mappings snm where snm.name_id = n.id and not snm.deleted)`

const nameRepairDetailSelect = `select n.id, n.name, n.reserved, n.account_id, ns.token as namespace_token
	from names n
	join namespaces ns on n.namespace_id = ns.id
	`

// CountNamesOfDeletedAccounts counts the live names whose account is deleted.
func (str *Store) CountNamesOfDeletedAccounts(trx *sqlx.Tx) (int, error) {
	return str.countNames(namesOfDeletedAccountsWhere, "names of deleted accounts", trx)
}

// FindNamesOfDeletedAccounts lists, lowest id first, up to limit live names whose account is deleted.
func (str *Store) FindNamesOfDeletedAccounts(limit int, trx *sqlx.Tx) ([]*NameRepairDetail, error) {
	return str.findNameRepairDetails(nameRepairDetailSelect+namesOfDeletedAccountsWhere+" order by n.id limit $1", "names of deleted accounts", limit, trx)
}

// CountUnmappedAllocatedNames counts the live auto-allocated names that no live mapping holds.
func (str *Store) CountUnmappedAllocatedNames(trx *sqlx.Tx) (int, error) {
	return str.countNames(unmappedAllocatedNamesWhere, "unmapped allocated names", trx)
}

// FindUnmappedAllocatedNames lists, lowest id first, up to limit live auto-allocated names that no live
// mapping holds.
func (str *Store) FindUnmappedAllocatedNames(limit int, trx *sqlx.Tx) ([]*NameRepairDetail, error) {
	return str.findNameRepairDetails(nameRepairDetailSelect+unmappedAllocatedNamesWhere+" order by n.id limit $1", "unmapped allocated names", limit, trx)
}

func (str *Store) countNames(where, what string, trx *sqlx.Tx) (int, error) {
	var count int
	if err := trx.QueryRow("select count(*) from names n " + where).Scan(&count); err != nil {
		return 0, errors.Wrapf(err, "error selecting count of %s", what)
	}
	return count, nil
}

func (str *Store) findNameRepairDetails(sql, what string, limit int, trx *sqlx.Tx) ([]*NameRepairDetail, error) {
	rows, err := trx.Queryx(sql, limit)
	if err != nil {
		return nil, errors.Wrapf(err, "error finding %s", what)
	}
	defer func() { _ = rows.Close() }()
	var names []*NameRepairDetail
	for rows.Next() {
		n := &NameRepairDetail{}
		if err := rows.StructScan(n); err != nil {
			return nil, errors.Wrapf(err, "error scanning %s", what)
		}
		names = append(names, n)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrapf(err, "error iterating %s", what)
	}
	return names, nil
}

// DeleteNames soft-deletes the live names among ids, reserved or not, and returns how many it deleted.
func (str *Store) DeleteNames(ids []int, trx *sqlx.Tx) (int64, error) {
	return str.deleteNamesWhere("not deleted", ids, trx)
}

// DeleteShareNameMappingsForNames soft-deletes the live share name mappings on the names among ids and
// returns how many it deleted.
func (str *Store) DeleteShareNameMappingsForNames(ids []int, trx *sqlx.Tx) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	query, args, err := sqlx.In("update share_name_mappings set updated_at = current_timestamp, deleted = true where name_id in (?) and not deleted", ids)
	if err != nil {
		return 0, errors.Wrap(err, "error building share name mappings delete by name statement")
	}
	res, err := trx.Exec(trx.Rebind(query), args...)
	if err != nil {
		return 0, errors.Wrap(err, "error executing share name mappings delete by name statement")
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "error reading share name mappings deleted by name")
	}
	return n, nil
}

// DeleteUnmappedAllocatedNames soft-deletes the live auto-allocated names among ids that no live mapping
// holds when the statement runs, and returns how many it deleted.
func (str *Store) DeleteUnmappedAllocatedNames(ids []int, trx *sqlx.Tx) (int64, error) {
	return str.deleteNamesWhere("not deleted and not reserved and not exists (select 1 from share_name_mappings snm where snm.name_id = names.id and not snm.deleted)", ids, trx)
}
