-- +migrate Up

-- environments.account_id should allow NULL; environments with NULL account_id are "ephemeral"
alter table environments rename to environments_old;
create table environments (
  id                    integer             primary key,
  account_id            integer             references accounts(id) on delete cascade,
  description           string,
  host                  string,
  address               string,
  z_id                  string              not null unique,
  created_at            datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
  updated_at            datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),

  constraint chk_z_id check (z_id <> '')
);
insert into environments select * from environments_old;
drop table environments_old;

create table frontends (
   id                    integer             primary key,
   environment_id        integer             not null references environments(id),
   name                  varchar(32)         not null unique,
   z_id                  varchar(32)         not null unique,
   public_name           varchar(64)         unique,
   created_at            datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
   updated_at            datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now'))
);
