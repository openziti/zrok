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

-- recreate frontends table with new structure
alter table frontends rename to frontends_old;
create table frontends (
    id                    integer             primary key,
    environment_id        integer             references environments(id),
    token                 varchar(32)         not null unique,
    z_id                  varchar(32)         not null,
    public_name           varchar(64)         unique,
    url_template          varchar(1024),
    dynamic               boolean             not null default(false),
    bind_address          varchar(128),
    reserved              boolean             not null default(false),
    permission_mode       string              not null default('open'),
    description           text,
    created_at            datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
    updated_at            datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
    deleted               boolean             not null default(false)
);
insert into frontends (id, environment_id, token, z_id, public_name, url_template, bind_address, reserved, permission_mode, description, created_at, updated_at, deleted) 
select id, environment_id, token, z_id, public_name, url_template, bind_address, reserved, permission_mode, description, created_at, updated_at, deleted from frontends_old;
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

-- +migrate Down

-- recreate dependent tables without foreign key constraints
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

-- recreate original frontends table structure
alter table frontends rename to frontends_old;
create table frontends (
    id                    integer             primary key,
    environment_id        integer             references environments(id),
    token                 varchar(32)         not null unique,
    z_id                  varchar(32)         not null,
    public_name           varchar(64)         unique,
    url_template          varchar(1024),
    reserved              boolean             not null default(false),
    created_at            datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
    updated_at            datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
    deleted               boolean             not null default(false),
    bind_address          varchar(128),
    permission_mode       string              not null default('open'),
    description           text
);
insert into frontends select id, environment_id, token, z_id, public_name, url_template, reserved, created_at, updated_at, deleted, bind_address, permission_mode, description from frontends_old;
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