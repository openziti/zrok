-- +migrate Up

--
-- namespaces
--
create table namespaces (
  id                    integer             primary key,
  token                 varchar(32)         not null,
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

  constraint chk_allocated_name check (name <> '')
);

create unique index uk_allocated_names on allocated_names(namespace_id, name) where not deleted;

-- +migrate Down

drop index if exists uk_allocated_names;
drop index if exists uk_namespace_grants;
drop index if exists uk_namespace_name;
drop index if exists uk_namespace_token;
drop table if exists allocated_names;
drop table if exists namespace_grants;
drop table if exists namespaces;