-- +migrate Up

--
-- namespaces
--
create table namespaces (
  id                    integer             primary key,
  name                  varchar(255)        not null unique,
  description           text,
  created_at            datetime            not null default(current_timestamp),
  updated_at            datetime            not null default(current_timestamp),
  deleted               boolean             not null default(false),

  constraint chk_name check (name <> '')
);

--
-- namespace_grants
--
create table namespace_grants (
  id                    integer             primary key,
  namespace_id          integer             not null constraint fk_namespace_grants_namespaces references namespaces on delete cascade,
  account_id            integer             not null constraint fk_namespace_grants_accounts references accounts on delete cascade,
  created_at            datetime            not null default(current_timestamp),
  updated_at            datetime            not null default(current_timestamp),
  deleted               boolean             not null default(false),

  constraint uk_namespace_grants unique (namespace_id, account_id)
);

--
-- allocated_names
--
create table allocated_names (
  id                    integer             primary key,
  namespace_id          integer             not null constraint fk_allocated_names_namespaces references namespaces on delete cascade,
  name                  varchar(255)        not null,
  account_id            integer             not null constraint fk_allocated_names_accounts references accounts on delete cascade,
  created_at            datetime            not null default(current_timestamp),
  updated_at            datetime            not null default(current_timestamp),
  deleted               boolean             not null default(false),

  constraint uk_allocated_names unique (namespace_id, name),
  constraint chk_allocated_name check (name <> '')
);

-- +migrate Down

drop table if exists allocated_names;
drop table if exists namespace_grants;
drop table if exists namespaces;