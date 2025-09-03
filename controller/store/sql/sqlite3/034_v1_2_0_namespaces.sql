-- +migrate Up

--
-- namespaces
--
create table namespaces (
  id                    integer             primary key,
  token                 varchar(64)         not null,
  name                  varchar(255)        not null,
  description           text,
  open                  boolean             not null default(false),
  created_at            datetime            not null default(current_timestamp),
  updated_at            datetime            not null default(current_timestamp),
  deleted               boolean             not null default(false),

  constraint chk_name check (name <> '')
);

create unique index uk_namespace_token on namespaces(token) where not deleted;
create unique index uk_namespace_name on namespaces(name) where not deleted;

--
-- namespace_grants
--
create table namespace_grants (
  id                    integer             primary key,
  namespace_id          integer             not null constraint fk_namespace_grants_namespaces references namespaces on delete cascade,
  account_id            integer             not null constraint fk_namespace_grants_accounts references accounts on delete cascade,
  created_at            datetime            not null default(current_timestamp),
  updated_at            datetime            not null default(current_timestamp),
  deleted               boolean             not null default(false)
);

create unique index uk_namespace_grants on namespace_grants(namespace_id, account_id) where not deleted;

--
-- names
--
create table names (
  id                    integer             primary key,
  namespace_id          integer             not null constraint fk_names_namespaces references namespaces on delete cascade,
  account_id            integer             not null constraint fk_names_accounts references accounts on delete cascade,
  name                  varchar(255)        not null,
  reserved              boolean             not null default(false),
  created_at            datetime            not null default(current_timestamp),
  updated_at            datetime            not null default(current_timestamp),
  deleted               boolean             not null default(false),

  constraint chk_name check (name <> '')
);

create unique index uk_names on names(namespace_id, name) where not deleted;

--
-- share_name_mappings
--
create table share_name_mappings (
  id                    integer             primary key,
  share_id              integer             not null constraint fk_share_name_mappings_shares references shares on delete cascade,
  name_id               integer             not null constraint fk_share_name_mappings_names references names on delete cascade,
  created_at            datetime            not null default(current_timestamp),
  updated_at            datetime            not null default(current_timestamp),
  deleted               boolean             not null default(false)
);

create unique index uk_share_name_mappings_name on share_name_mappings(name_id) where not deleted;
create index idx_share_name_mappings_share on share_name_mappings(share_id) where not deleted;

-- +migrate Down

drop index if exists idx_share_name_mappings_share;
drop index if exists uk_share_name_mappings_name;
drop index if exists uk_names;
drop index if exists uk_namespace_grants;
drop index if exists uk_namespace_name;
drop index if exists uk_namespace_token;
drop table if exists share_name_mappings;
drop table if exists names;
drop table if exists namespace_grants;
drop table if exists namespaces;