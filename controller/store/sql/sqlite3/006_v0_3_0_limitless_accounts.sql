-- +migrate Up

alter table accounts rename to accounts_old;

create table accounts (
  id                    integer             primary key,
  email                 string              not null unique,
  password              string              not null,
  token                 string              not null unique,
  limitless             boolean             not null default(false),
  created_at            datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
  updated_at            datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),

  constraint chk_email check (email <> ''),
  constraint chk_password check (password <> ''),
  constraint chk_token check(token <> '')
);

insert into accounts (id, email, password, token, created_at, updated_at)
    select id, email, password, token, created_at, updated_at from accounts_old;

drop table accounts_old;