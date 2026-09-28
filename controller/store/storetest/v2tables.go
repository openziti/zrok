// Package storetest holds fixtures for tests that run against the controller store.
package storetest

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// V2MappingsDDL creates the v2 line's name and frontend mapping tables, as the v2 sqlite migrations
// (034_v2_0_0_namespaces.sql, 037_v2_0_0_frontend_mappings.sql) define them, in a v1-migrated sqlite store.
var V2MappingsDDL = []string{
	`create table namespaces (
	  id                    integer             primary key,
	  token                 varchar(64)         not null,
	  name                  varchar(255)        not null,
	  description           text,
	  open                  boolean             not null default(false),
	  created_at            datetime            not null default(current_timestamp),
	  updated_at            datetime            not null default(current_timestamp),
	  deleted               boolean             not null default(false),
	  constraint chk_name check (name <> '')
	)`,
	`create table names (
	  id                    integer             primary key,
	  namespace_id          integer             not null constraint fk_names_namespaces references namespaces on delete cascade,
	  account_id            integer             not null constraint fk_names_accounts references accounts on delete cascade,
	  name                  varchar(255)        not null,
	  reserved              boolean             not null default(false),
	  created_at            datetime            not null default(current_timestamp),
	  updated_at            datetime            not null default(current_timestamp),
	  deleted               boolean             not null default(false),
	  constraint chk_name check (name <> '')
	)`,
	`create unique index uk_names on names(namespace_id, name) where not deleted`,
	`create table share_name_mappings (
	  id                    integer             primary key,
	  share_id              integer             not null constraint fk_share_name_mappings_shares references shares on delete cascade,
	  name_id               integer             not null constraint fk_share_name_mappings_names references names on delete cascade,
	  created_at            datetime            not null default(current_timestamp),
	  updated_at            datetime            not null default(current_timestamp),
	  deleted               boolean             not null default(false)
	)`,
	`create unique index uk_share_name_mappings_name on share_name_mappings(name_id) where not deleted`,
	`create table frontend_mappings (
	  id                      integer             primary key autoincrement,
	  frontend_token          string              not null,
	  name                    string              not null,
	  share_token             string              not null,
	  created_at              datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
	  unique (frontend_token, name)
	)`,
	`create index frontend_mappings_share_token_idx on frontend_mappings (share_token)`,
}

// CreateV2Tables runs V2MappingsDDL in the transaction.
func CreateV2Tables(trx *sqlx.Tx) error {
	for _, ddl := range V2MappingsDDL {
		if _, err := trx.Exec(ddl); err != nil {
			return errors.Wrap(err, "error creating v2 mapping tables")
		}
	}
	return nil
}

// V2Fixture is a share published under a reserved and an auto-allocated name, with frontend mappings for its token.
type V2Fixture struct {
	ReservedNameId int
	AutoNameId     int
}

// MapShare creates a namespace, a reserved and an auto-allocated name mapped to the share, and two frontend
// mappings carrying the share token.
func MapShare(trx *sqlx.Tx, acctId, shrId int, shrToken, prefix string) (*V2Fixture, error) {
	var nsId int
	if err := trx.QueryRow("insert into namespaces (token, name) values ($1, $2) returning id", prefix+"-ns", prefix+"-ns").Scan(&nsId); err != nil {
		return nil, errors.Wrap(err, "error inserting namespace")
	}
	f := &V2Fixture{}
	if err := trx.QueryRow("insert into names (namespace_id, account_id, name, reserved) values ($1, $2, $3, true) returning id", nsId, acctId, prefix+"-reserved").Scan(&f.ReservedNameId); err != nil {
		return nil, errors.Wrap(err, "error inserting reserved name")
	}
	if err := trx.QueryRow("insert into names (namespace_id, account_id, name, reserved) values ($1, $2, $3, false) returning id", nsId, acctId, prefix+"-auto").Scan(&f.AutoNameId); err != nil {
		return nil, errors.Wrap(err, "error inserting auto-allocated name")
	}
	for _, nameId := range []int{f.ReservedNameId, f.AutoNameId} {
		if _, err := trx.Exec("insert into share_name_mappings (share_id, name_id) values ($1, $2)", shrId, nameId); err != nil {
			return nil, errors.Wrap(err, "error inserting share name mapping")
		}
	}
	for _, name := range []string{prefix + "-reserved", prefix + "-auto"} {
		if _, err := trx.Exec("insert into frontend_mappings (frontend_token, name, share_token) values ($1, $2, $3)", prefix+"-fe", name, shrToken); err != nil {
			return nil, errors.Wrap(err, "error inserting frontend mapping")
		}
	}
	return f, nil
}

// LiveMappings counts the share's live name mappings and the frontend mappings carrying its token.
func LiveMappings(trx *sqlx.Tx, shrId int, shrToken string) (names int, frontends int, err error) {
	if err := trx.QueryRow("select count(*) from share_name_mappings where share_id = $1 and not deleted", shrId).Scan(&names); err != nil {
		return 0, 0, errors.Wrap(err, "error counting share name mappings")
	}
	if err := trx.QueryRow("select count(*) from frontend_mappings where share_token = $1", shrToken).Scan(&frontends); err != nil {
		return 0, 0, errors.Wrap(err, "error counting frontend mappings")
	}
	return names, frontends, nil
}

// NameDeleted reports whether the name row is soft-deleted.
func NameDeleted(trx *sqlx.Tx, nameId int) (bool, error) {
	var deleted bool
	if err := trx.QueryRow("select deleted from names where id = $1", nameId).Scan(&deleted); err != nil {
		return false, errors.Wrap(err, "error selecting name")
	}
	return deleted, nil
}
