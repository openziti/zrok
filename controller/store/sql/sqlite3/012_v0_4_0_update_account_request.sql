-- +migrate Up

alter table account_requests rename to account_requests_old;
create table account_requests (
  id                    integer             primary key,
  token                 string              not null unique,
  email                 string              not null,
  source_address        string              not null,
  created_at            datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
  updated_at            datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
  deleted               boolean             not null default(false)
);

insert into account_requests select * from account_requests_old;
drop table account_requests_old;