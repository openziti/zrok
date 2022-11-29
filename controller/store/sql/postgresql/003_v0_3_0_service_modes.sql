-- +migrate Up

create type share_mode as enum ('public', 'private');
create type backend_mode as enum ('proxy', 'web', 'dav');

alter table services
    add column frontend_selection varchar(64),
    add column share_mode share_mode not null default 'public',
    add column backend_mode backend_mode not null default 'proxy',
    add column reserved boolean not null default false;

alter table services
    alter column share_mode drop default;
alter table services
    alter column backend_mode drop default;

alter table services rename frontend to frontend_endpoint;
alter table services rename backend to backend_proxy_endpoint;

alter table services rename to services_old;

create table services (
  id                        serial              primary key,
  environment_id            integer             not null references environments(id),
  z_id                      varchar(32)         not null unique,
  name                      varchar(32)         not null unique,
  share_mode                share_mode          not null,
  backend_mode              backend_mode        not null,
  frontend_selection        varchar(64),
  frontend_endpoint         varchar(1024),
  backend_proxy_endpoint    varchar(1024),
  reserved                  boolean             not null default(false),
  created_at                timestamptz         not null default(current_timestamp),
  updated_at                timestamptz         not null default(current_timestamp),

  constraint chk_z_id check (z_id <> ''),
  constraint chk_name check (name <> '')
);

insert into services (id, environment_id, z_id, name, share_mode, backend_mode, frontend_selection, frontend_endpoint, backend_proxy_endpoint, created_at, updated_at)
    select id, environment_id, z_id, name, share_mode, backend_mode, frontend_selection, frontend_endpoint, backend_proxy_endpoint, created_at, updated_at from services_old;

drop table services_old;