-- +migrate Up

-- first, drop foreign key constraints from dependent tables
alter table frontend_grants drop constraint if exists frontend_grants_frontend_id_fkey;
alter table namespace_frontend_mappings drop constraint if exists fk_namespace_frontend_mappings_frontends;

-- recreate frontends table with new structure
alter table frontends rename to frontends_old;
create table frontends (
    id                    serial              primary key,
    environment_id        integer             references environments(id),
    token                 varchar(32)         not null unique,
    z_id                  varchar(32)         not null,
    public_name           varchar(64)         unique,
    url_template          varchar(1024),
    dynamic               boolean             not null default(false),
    bind_address          varchar(128),
    reserved              boolean             not null default(false),
    permission_mode       permission_mode_type not null default('open'),
    description           text,
    created_at            timestamp           not null default(current_timestamp),
    updated_at            timestamp           not null default(current_timestamp),
    deleted               boolean             not null default(false)
);
insert into frontends (id, environment_id, token, z_id, public_name, url_template, bind_address, reserved, permission_mode, description, created_at, updated_at, deleted) 
select id, environment_id, token, z_id, public_name, url_template, bind_address, reserved, permission_mode, description, created_at, updated_at, deleted from frontends_old;
drop table frontends_old;

-- recreate indexes
create index frontends_environment_id_idx on frontends (environment_id);

-- recreate foreign key constraints
alter table frontend_grants add constraint frontend_grants_frontend_id_fkey foreign key (frontend_id) references frontends(id);
alter table namespace_frontend_mappings add constraint fk_namespace_frontend_mappings_frontends foreign key (frontend_id) references frontends(id) on delete cascade;

-- +migrate Down

-- drop foreign key constraints
alter table frontend_grants drop constraint if exists frontend_grants_frontend_id_fkey;
alter table namespace_frontend_mappings drop constraint if exists fk_namespace_frontend_mappings_frontends;

-- recreate original table structure
alter table frontends rename to frontends_old;
create table frontends (
    id                    serial              primary key,
    environment_id        integer             references environments(id),
    token                 varchar(32)         not null unique,
    z_id                  varchar(32)         not null,
    public_name           varchar(64)         unique,
    url_template          varchar(1024),
    reserved              boolean             not null default(false),
    created_at            timestamp           not null default(current_timestamp),
    updated_at            timestamp           not null default(current_timestamp),
    deleted               boolean             not null default(false),
    bind_address          varchar(128),
    permission_mode       permission_mode_type not null default('open'),
    description           text
);
insert into frontends select id, environment_id, token, z_id, public_name, url_template, reserved, created_at, updated_at, deleted, bind_address, permission_mode, description from frontends_old;
drop table frontends_old;

-- recreate indexes and constraints
create index frontends_environment_id_idx on frontends (environment_id);
alter table frontend_grants add constraint frontend_grants_frontend_id_fkey foreign key (frontend_id) references frontends(id);
alter table namespace_frontend_mappings add constraint fk_namespace_frontend_mappings_frontends foreign key (frontend_id) references frontends(id) on delete cascade;