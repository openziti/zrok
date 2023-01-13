-- +migrate Up

alter table accounts rename to accounts_old;

create table accounts (
  id                    serial              primary key,
  email                 varchar(1024)       not null unique,
  password              char(128)           not null,
  token                 varchar(32)         not null unique,
  limitless             boolean             not null default(false),
  created_at            timestamp           not null default(current_timestamp),
  updated_at            timestamp           not null default(current_timestamp),

  constraint chk_email check (email <> ''),
  constraint chk_password check (password <> ''),
  constraint chk_token check(token <> '')
);

insert into accounts(id, email, password, token, created_at, updated_at)
    select id, email, password, token, created_at, updated_at from accounts_old;

alter table environments drop constraint fk_accounts_id;
alter table environments add constraint fk_accounts_id foreign key (account_id) references accounts(id);

drop table accounts_old;