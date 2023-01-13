-- +migrate Up

create table shares (
      id                        integer             primary key,
      environment_id            integer             constraint fk_environments_shares references environments on delete cascade,
      z_id                      string              not null unique,
      token                     string              not null unique,
      share_mode                string              not null,
      backend_mode              string              not null,
      frontend_selection        string,
      frontend_endpoint         string,
      backend_proxy_endpoint    string,
      reserved                  boolean             not null default(false),
      created_at                datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
      updated_at                datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),

      constraint chk_z_id check (z_id <> ''),
      constraint chk_token check (token <> ''),
      constraint chk_share_mode check (share_mode == 'public' or share_mode == 'private'),
      constraint chk_backend_mode check (backend_mode == 'proxy' or backend_mode == 'web' or backend_mode == 'dav')
);

insert into shares select * from services;

drop table services;
