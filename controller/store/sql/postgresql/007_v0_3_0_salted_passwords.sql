-- +migrate Up

alter table accounts rename to accounts_old;
alter sequence accounts_id_seq rename to accounts_id_seq_old;

create table accounts (
  id                    serial              primary key,
  email                 varchar(1024)       not null unique,
  salt                  varchar(16)         not null default(''),
  password              varchar(64)         not null default(''),
  token                 varchar(32)         not null unique,
  limitless             boolean             not null default(false),
  created_at            timestamp           not null default(current_timestamp),
  updated_at            timestamp           not null default(current_timestamp),

  constraint chk_email check (email <> ''),
  constraint chk_token check(token <> '')
);

insert into accounts(id, email, token, limitless, created_at, updated_at)
    select id, email, token, limitless, created_at, updated_at from accounts_old;

alter table accounts alter column salt drop default;
alter table accounts alter column password drop default;

select setval('accounts_id_seq', (select max(id) from accounts));

alter table environments drop constraint fk_accounts_id;
alter table environments add constraint fk_accounts_id foreign key (account_id) references accounts(id);

alter table password_reset_requests drop constraint fk_accounts_password_reset_requests;
alter table password_reset_requests add constraint fk_accounts_password_reset_requests foreign key (account_id) references accounts(id);

drop table accounts_old;

alter index accounts_pkey1 rename to accounts_pkey;
alter index accounts_email_key1 rename to accounts_email_key;
alter index accounts_token_key1 rename to accounts_token_key;