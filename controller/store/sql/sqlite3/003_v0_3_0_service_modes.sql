-- +migrate Up

alter table services add column frontend_selection string;
alter table services add column share_mode string not null default 'public';
alter table services add column backend_mode string not null default 'proxy';
alter table services add column reserved boolean not null default false;

alter table services rename to services_old;

create table services (
  id                        integer             primary key,
  environment_id            integer             constraint fk_environments_services references environments on delete cascade,
  z_id                      string              not null unique,
  name                      string              not null unique,
  share_mode                string              not null,
  backend_mode              string              not null,
  frontend_selection        string,
  frontend_endpoint         string,
  backend_proxy_endpoint    string,
  reserved                  boolean             not null default(false),
  created_at                datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
  updated_at                datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),

  constraint chk_z_id check (z_id <> ''),
  constraint chk_name check (name <> ''),
  constraint chk_share_mode check (share_mode == 'public' or share_mode == 'private'),
  constraint chk_backend_mode check (backend_mode == 'proxy' or backend_mode == 'web' or backend_mode == 'dav')
);

insert into services (id, environment_id, z_id, name, share_mode, backend_mode, frontend_selection, frontend_endpoint, backend_proxy_endpoint, created_at, updated_at)
    select id, environment_id, z_id, name, share_mode, backend_mode, frontend_selection, frontend, backend, created_at, updated_at from services_old;

drop table services_old;
