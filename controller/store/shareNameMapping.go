package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type ShareNameMapping struct {
	Model
	ShareId int
	NameId  int
}

type ShareNameMappingWithShare struct {
	ShareNameMapping
	ShareToken   string `db:"share_token"`
	ShareDeleted bool   `db:"share_deleted"`
}

type ShareNameCleanupDetail struct {
	MappingId        int    `db:"mapping_id"`
	NameId           int    `db:"name_id"`
	Name             string `db:"name"`
	NameDeleted      bool   `db:"name_deleted"`
	Reserved         bool   `db:"reserved"`
	NamespaceID      int    `db:"namespace_id"`
	NamespaceName    string `db:"namespace_name"`
	NamespaceDeleted bool   `db:"namespace_deleted"`
}

func (str *Store) CreateShareNameMapping(snm *ShareNameMapping, trx *sqlx.Tx) (int, error) {
	stmt, err := trx.Prepare("insert into share_name_mappings (share_id, name_id) values ($1, $2) returning id")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing share name mapping insert statement")
	}
	var id int
	if err := stmt.QueryRow(snm.ShareId, snm.NameId).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "error executing share name mapping insert statement")
	}
	return id, nil
}

func (str *Store) GetShareNameMapping(id int, trx *sqlx.Tx) (*ShareNameMapping, error) {
	snm := &ShareNameMapping{}
	if err := trx.QueryRowx("select * from share_name_mappings where id = $1 and not deleted", id).StructScan(snm); err != nil {
		return nil, errors.Wrap(err, "error selecting share name mapping by id")
	}
	return snm, nil
}

func (str *Store) FindShareNameMappingsByShareId(shareId int, trx *sqlx.Tx) ([]*ShareNameMapping, error) {
	rows, err := trx.Queryx("select * from share_name_mappings where share_id = $1 and not deleted", shareId)
	if err != nil {
		return nil, errors.Wrap(err, "error finding share name mappings by share id")
	}
	var mappings []*ShareNameMapping
	for rows.Next() {
		snm := &ShareNameMapping{}
		if err := rows.StructScan(&snm); err != nil {
			return nil, errors.Wrap(err, "error scanning share name mapping")
		}
		mappings = append(mappings, snm)
	}
	return mappings, nil
}

func (str *Store) FindShareNameMappingsByNameId(nameId int, trx *sqlx.Tx) ([]*ShareNameMapping, error) {
	rows, err := trx.Queryx("select * from share_name_mappings where name_id = $1 and not deleted", nameId)
	if err != nil {
		return nil, errors.Wrap(err, "error finding share name mappings by name id")
	}
	var mappings []*ShareNameMapping
	for rows.Next() {
		snm := &ShareNameMapping{}
		if err := rows.StructScan(&snm); err != nil {
			return nil, errors.Wrap(err, "error scanning share name mapping")
		}
		mappings = append(mappings, snm)
	}
	return mappings, nil
}

func (str *Store) FindShareNameMappingsByNameIdWithShare(nameId int, trx *sqlx.Tx) ([]*ShareNameMappingWithShare, error) {
	sql := `select snm.id, snm.created_at, snm.updated_at, snm.deleted, snm.share_id, snm.name_id,
	               s.token as share_token, s.deleted as share_deleted
	        from share_name_mappings snm
	        join shares s on snm.share_id = s.id
	        where snm.name_id = $1 and not snm.deleted
	        order by snm.id`

	rows, err := trx.Queryx(sql, nameId)
	if err != nil {
		return nil, errors.Wrap(err, "error finding share name mappings with share by name id")
	}

	var mappings []*ShareNameMappingWithShare
	for rows.Next() {
		snm := &ShareNameMappingWithShare{}
		if err := rows.StructScan(snm); err != nil {
			return nil, errors.Wrap(err, "error scanning share name mapping with share")
		}
		mappings = append(mappings, snm)
	}

	return mappings, nil
}

func (str *Store) FindShareNameMappingByShareIdAndNameId(shareId, nameId int, trx *sqlx.Tx) (*ShareNameMapping, error) {
	snm := &ShareNameMapping{}
	if err := trx.QueryRowx("select * from share_name_mappings where share_id = $1 and name_id = $2 and not deleted", shareId, nameId).StructScan(snm); err != nil {
		return nil, errors.Wrap(err, "error selecting share name mapping by share id and name id")
	}
	return snm, nil
}

func (str *Store) FindShareNameCleanupDetailsByShareId(shareId int, trx *sqlx.Tx) ([]*ShareNameCleanupDetail, error) {
	sql := `select snm.id as mapping_id, snm.name_id,
	               n.name, n.deleted as name_deleted, n.reserved,
	               ns.id as namespace_id, ns.name as namespace_name, ns.deleted as namespace_deleted
	        from share_name_mappings snm
	        join names n on snm.name_id = n.id
	        join namespaces ns on n.namespace_id = ns.id
	        where snm.share_id = $1 and not snm.deleted
	        order by snm.id`

	rows, err := trx.Queryx(sql, shareId)
	if err != nil {
		return nil, errors.Wrap(err, "error finding share name cleanup details by share id")
	}

	var details []*ShareNameCleanupDetail
	for rows.Next() {
		detail := &ShareNameCleanupDetail{}
		if err := rows.StructScan(detail); err != nil {
			return nil, errors.Wrap(err, "error scanning share name cleanup detail")
		}
		details = append(details, detail)
	}

	return details, nil
}

func (str *Store) DeleteShareNameMapping(id int, trx *sqlx.Tx) error {
	stmt, err := trx.Prepare("update share_name_mappings set updated_at = current_timestamp, deleted = true where id = $1")
	if err != nil {
		return errors.Wrap(err, "error preparing share name mapping delete statement")
	}
	_, err = stmt.Exec(id)
	if err != nil {
		return errors.Wrap(err, "error executing share name mapping delete statement")
	}
	return nil
}

// ShareNameMappingRepairDetail is a live share name mapping found by the store repair sweep, with what a
// reviewer needs to spot-check it.
type ShareNameMappingRepairDetail struct {
	MappingId      int    `db:"mapping_id"`
	NameId         int    `db:"name_id"`
	Name           string `db:"name"`
	NameDeleted    bool   `db:"name_deleted"`
	Reserved       bool   `db:"reserved"`
	NamespaceToken string `db:"namespace_token"`
	ShareToken     string `db:"share_token"`
}

// a live mapping whose share has no live row. a mapping whose name is also deleted belongs here, not to
// the deleted-name condition, so the two never count a row twice.
const shareNameMappingsToDeletedSharesWhere = `where not snm.deleted
	  and n.reserved = $1
	  and not exists (select 1 from shares s where s.id = snm.share_id and not s.deleted)`

// a live mapping, whose share is live, to a deleted name.
const shareNameMappingsToDeletedNamesWhere = `where not snm.deleted
	  and n.deleted
	  and exists (select 1 from shares s where s.id = snm.share_id and not s.deleted)`

const shareNameMappingRepairDetailSelect = `select snm.id as mapping_id, snm.name_id,
	       n.name, n.deleted as name_deleted, n.reserved,
	       ns.token as namespace_token, coalesce(s.token, '') as share_token
	from share_name_mappings snm
	join names n on snm.name_id = n.id
	join namespaces ns on n.namespace_id = ns.id
	left join shares s on snm.share_id = s.id
	`

// CountShareNameMappingsToDeletedShares counts the live share name mappings of reserved (or of
// auto-allocated) names whose share is deleted.
func (str *Store) CountShareNameMappingsToDeletedShares(reserved bool, trx *sqlx.Tx) (int, error) {
	var count int
	sql := "select count(*) from share_name_mappings snm join names n on snm.name_id = n.id " + shareNameMappingsToDeletedSharesWhere
	if err := trx.QueryRow(sql, reserved).Scan(&count); err != nil {
		return 0, errors.Wrap(err, "error counting share name mappings to deleted shares")
	}
	return count, nil
}

// CountAllocatedNamesHeldByDeletedShares counts the live auto-allocated names held by a live mapping to a
// deleted share: the names a repair of those mappings releases.
func (str *Store) CountAllocatedNamesHeldByDeletedShares(trx *sqlx.Tx) (int, error) {
	var count int
	sql := "select count(*) from share_name_mappings snm join names n on snm.name_id = n.id " + shareNameMappingsToDeletedSharesWhere + " and not n.deleted"
	if err := trx.QueryRow(sql, false).Scan(&count); err != nil {
		return 0, errors.Wrap(err, "error counting allocated names held by deleted shares")
	}
	return count, nil
}

// FindShareNameMappingsToDeletedShares lists, lowest id first, up to limit live share name mappings of
// reserved (or of auto-allocated) names whose share is deleted.
func (str *Store) FindShareNameMappingsToDeletedShares(reserved bool, limit int, trx *sqlx.Tx) ([]*ShareNameMappingRepairDetail, error) {
	sql := shareNameMappingRepairDetailSelect + shareNameMappingsToDeletedSharesWhere + " order by snm.id limit $2"
	return str.findShareNameMappingRepairDetails(sql, trx, reserved, limit)
}

// CountShareNameMappingsToDeletedNames counts the live share name mappings of live shares whose name is
// deleted.
func (str *Store) CountShareNameMappingsToDeletedNames(trx *sqlx.Tx) (int, error) {
	var count int
	sql := "select count(*) from share_name_mappings snm join names n on snm.name_id = n.id " + shareNameMappingsToDeletedNamesWhere
	if err := trx.QueryRow(sql).Scan(&count); err != nil {
		return 0, errors.Wrap(err, "error counting share name mappings to deleted names")
	}
	return count, nil
}

// FindShareNameMappingsToDeletedNames lists, lowest id first, up to limit live share name mappings of
// live shares whose name is deleted.
func (str *Store) FindShareNameMappingsToDeletedNames(limit int, trx *sqlx.Tx) ([]*ShareNameMappingRepairDetail, error) {
	sql := shareNameMappingRepairDetailSelect + shareNameMappingsToDeletedNamesWhere + " order by snm.id limit $1"
	return str.findShareNameMappingRepairDetails(sql, trx, limit)
}

func (str *Store) findShareNameMappingRepairDetails(sql string, trx *sqlx.Tx, args ...interface{}) ([]*ShareNameMappingRepairDetail, error) {
	rows, err := trx.Queryx(sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "error finding share name mappings to repair")
	}
	defer func() { _ = rows.Close() }()
	var details []*ShareNameMappingRepairDetail
	for rows.Next() {
		detail := &ShareNameMappingRepairDetail{}
		if err := rows.StructScan(detail); err != nil {
			return nil, errors.Wrap(err, "error scanning share name mapping to repair")
		}
		details = append(details, detail)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating share name mappings to repair")
	}
	return details, nil
}

// DeleteShareNameMappings soft-deletes the live share name mappings among ids and returns how many it
// deleted.
func (str *Store) DeleteShareNameMappings(ids []int, trx *sqlx.Tx) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	query, args, err := sqlx.In("update share_name_mappings set updated_at = current_timestamp, deleted = true where id in (?) and not deleted", ids)
	if err != nil {
		return 0, errors.Wrap(err, "error building share name mappings delete statement")
	}
	res, err := trx.Exec(trx.Rebind(query), args...)
	if err != nil {
		return 0, errors.Wrap(err, "error executing share name mappings delete statement")
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "error reading share name mappings deleted")
	}
	return n, nil
}
