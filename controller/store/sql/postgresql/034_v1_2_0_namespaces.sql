-- +migrate Up

--
-- namespaces
--
create table namespaces (
  id                    serial              primary key,
  token                 varchar(64)         not null,
  name                  varchar(255)        not null,
  description           text,
  open                  boolean             not null default(false),
  created_at            timestamp           not null default(current_timestamp),
  updated_at            timestamp           not null default(current_timestamp),
  deleted               boolean             not null default(false),

  constraint chk_name check (name <> ''),
  constraint uk_namespace_token unique (token) where not deleted,
  constraint uk_namespace_name unique (name) where not deleted
);

--
-- namespace_grants
--
create table namespace_grants (
  id                    serial              primary key,
  namespace_id          integer             not null constraint fk_namespace_grants_namespaces references namespaces on delete cascade,
  account_id            integer             not null constraint fk_namespace_grants_accounts references accounts on delete cascade,
  created_at            timestamp           not null default(current_timestamp),
  updated_at            timestamp           not null default(current_timestamp),
  deleted               boolean             not null default(false),

  constraint uk_namespace_grants unique (namespace_id, account_id) where not deleted
);

--
-- names
--
create table names (
  id                    serial              primary key,
  namespace_id          integer             not null constraint fk_names_namespaces references namespaces on delete cascade,
  account_id            integer             not null constraint fk_names_accounts references accounts on delete cascade,
  name                  varchar(255)        not null,
  reserved              boolean             not null default(false),
  created_at            timestamp           not null default(current_timestamp),
  updated_at            timestamp           not null default(current_timestamp),
  deleted               boolean             not null default(false),

  constraint uk_names unique (namespace_id, name) where not deleted,
  constraint chk_name check (name <> '')
);

-- +migrate Down

drop table if exists names;
drop table if exists namespace_grants;
drop table if exists namespaces;