-- +migrate Up

-- recreate namespace_frontend_mappings without foreign key constraint
alter table namespace_frontend_mappings rename to namespace_frontend_mappings_old;
create table namespace_frontend_mappings (
  id                    integer             primary key,
  namespace_id          integer             not null,
  frontend_id           integer             not null,
  is_default            boolean             not null default(false),
  created_at            timestamp           not null default(current_timestamp),
  updated_at            timestamp           not null default(current_timestamp),
  deleted               boolean             not null default(false)
);
insert into namespace_frontend_mappings select * from namespace_frontend_mappings_old;
drop table namespace_frontend_mappings_old;

-- recreate frontend_grants without foreign key constraint
alter table frontend_grants rename to frontend_grants_old;
create table frontend_grants (
    id                  integer             primary key,
    account_id          integer             not null,
    frontend_id         integer             not null,
    created_at          datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
    updated_at          datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
    deleted             boolean             not null default(false)
);
insert into frontend_grants select * from frontend_grants_old;
drop table frontend_grants_old;

-- recreate frontends table without inline unique constraint on public_name
alter table frontends rename to frontends_old;
create table frontends (
    id                    integer             primary key,
    environment_id        integer             references environments(id),
    token                 varchar(32)         not null unique,
    z_id                  varchar(32)         not null,
    public_name           varchar(64),
    url_template          varchar(1024),
    dynamic               boolean             not null default(false),
    private_share_id      integer,
    bind_address          varchar(128),
    reserved              boolean             not null default(false),
    permission_mode       string              not null default('open'),
    description           text,
    created_at            datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
    updated_at            datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
    deleted               boolean             not null default(false)
);
insert into frontends select * from frontends_old;
drop table frontends_old;

-- recreate dependent tables with proper foreign key constraints
alter table namespace_frontend_mappings rename to namespace_frontend_mappings_old;
create table namespace_frontend_mappings (
  id                    integer             primary key,
  namespace_id          integer             not null constraint fk_namespace_frontend_mappings_namespaces references namespaces on delete cascade,
  frontend_id           integer             not null constraint fk_namespace_frontend_mappings_frontends references frontends on delete cascade,
  is_default            boolean             not null default(false),
  created_at            timestamp           not null default(current_timestamp),
  updated_at            timestamp           not null default(current_timestamp),
  deleted               boolean             not null default(false)
);
insert into namespace_frontend_mappings select * from namespace_frontend_mappings_old;
drop table namespace_frontend_mappings_old;

alter table frontend_grants rename to frontend_grants_old;
create table frontend_grants (
    id                  integer             primary key,
    account_id          integer             references accounts (id) not null,
    frontend_id         integer             references frontends (id) not null,
    created_at          datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
    updated_at          datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
    deleted             boolean             not null default(false)
);
insert into frontend_grants select * from frontend_grants_old;
drop table frontend_grants_old;

-- recreate all indexes
create index frontends_environment_id_idx on frontends (environment_id);
create index frontend_grants_account_id_idx on frontend_grants (account_id);
create index frontend_grants_frontend_id_idx on frontend_grants (frontend_id);
create unique index uk_namespace_frontend_mappings on namespace_frontend_mappings(namespace_id, frontend_id) where not deleted;
create unique index uk_default_namespace_frontend on namespace_frontend_mappings(frontend_id) where is_default = 1 and not deleted;

-- create partial unique index on public_name excluding deleted records
create unique index uk_frontends_public_name on frontends(public_name) where not deleted;

-- +migrate Down

-- drop the partial unique index
drop index if exists uk_frontends_public_name;

-- drop indexes
drop index if exists frontends_environment_id_idx;
drop index if exists frontend_grants_account_id_idx;
drop index if exists frontend_grants_frontend_id_idx;
drop index if exists uk_namespace_frontend_mappings;
drop index if exists uk_default_namespace_frontend;

-- recreate namespace_frontend_mappings without foreign key constraint
alter table namespace_frontend_mappings rename to namespace_frontend_mappings_old;
create table namespace_frontend_mappings (
  id                    integer             primary key,
  namespace_id          integer             not null,
  frontend_id           integer             not null,
  is_default            boolean             not null default(false),
  created_at            timestamp           not null default(current_timestamp),
  updated_at            timestamp           not null default(current_timestamp),
  deleted               boolean             not null default(false)
);
insert into namespace_frontend_mappings select * from namespace_frontend_mappings_old;
drop table namespace_frontend_mappings_old;

-- recreate frontend_grants without foreign key constraint
alter table frontend_grants rename to frontend_grants_old;
create table frontend_grants (
    id                  integer             primary key,
    account_id          integer             not null,
    frontend_id         integer             not null,
    created_at          datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
    updated_at          datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
    deleted             boolean             not null default(false)
);
insert into frontend_grants select * from frontend_grants_old;
drop table frontend_grants_old;

-- recreate original frontends table with inline unique constraint
alter table frontends rename to frontends_new;
create table frontends (
    id                    integer             primary key,
    environment_id        integer             references environments(id),
    token                 varchar(32)         not null unique,
    z_id                  varchar(32)         not null,
    public_name           varchar(64)         unique,
    url_template          varchar(1024),
    dynamic               boolean             not null default(false),
    private_share_id      integer,
    bind_address          varchar(128),
    reserved              boolean             not null default(false),
    permission_mode       string              not null default('open'),
    description           text,
    created_at            datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
    updated_at            datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
    deleted               boolean             not null default(false)
);
insert into frontends select * from frontends_new;
drop table frontends_new;

-- recreate dependent tables with foreign key constraints
alter table namespace_frontend_mappings rename to namespace_frontend_mappings_old;
create table namespace_frontend_mappings (
  id                    integer             primary key,
  namespace_id          integer             not null constraint fk_namespace_frontend_mappings_namespaces references namespaces on delete cascade,
  frontend_id           integer             not null constraint fk_namespace_frontend_mappings_frontends references frontends on delete cascade,
  is_default            boolean             not null default(false),
  created_at            timestamp           not null default(current_timestamp),
  updated_at            timestamp           not null default(current_timestamp),
  deleted               boolean             not null default(false)
);
insert into namespace_frontend_mappings select * from namespace_frontend_mappings_old;
drop table namespace_frontend_mappings_old;

alter table frontend_grants rename to frontend_grants_old;
create table frontend_grants (
    id                  integer             primary key,
    account_id          integer             references accounts (id) not null,
    frontend_id         integer             references frontends (id) not null,
    created_at          datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
    updated_at          datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
    deleted             boolean             not null default(false)
);
insert into frontend_grants select * from frontend_grants_old;
drop table frontend_grants_old;

-- recreate indexes
create index frontends_environment_id_idx on frontends (environment_id);
create index frontend_grants_account_id_idx on frontend_grants (account_id);
create index frontend_grants_frontend_id_idx on frontend_grants (frontend_id);
create unique index uk_namespace_frontend_mappings on namespace_frontend_mappings(namespace_id, frontend_id) where not deleted;
create unique index uk_default_namespace_frontend on namespace_frontend_mappings(frontend_id) where is_default = 1 and not deleted;
