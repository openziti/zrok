-- +migrate Up

alter table services add column share_mode string default 'public';
alter table services add column backend_mode string default 'proxy';

alter table services rename to services_old;
create table services (
  id                    integer             primary key,
  environment_id        integer             constraint fk_environments_services references environments on delete cascade,
  z_id                  string              not null unique,
  name                  string              not null unique,
  frontend              string,
  backend               string,
  created_at            datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
  updated_at            datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
  share_mode            string              not null,
  backend_mode          string              not null

  constraint chk_z_id check (z_id <> ''),
  constraint chk_name check (name <> ''),
  constraint chk_share_mode check (share_mode == 'public' || share_mode == 'private'),
  constraint chk_backend_mode check (backend_mode == 'proxy' || backend_mode == 'web' || backend_mode == 'dav')
);
insert into services select * from services_old;
drop table services_old;
