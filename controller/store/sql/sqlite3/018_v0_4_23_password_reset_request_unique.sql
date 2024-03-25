-- +migrate Up

alter table password_reset_requests rename to password_reset_requests_old;

CREATE TABLE password_reset_requests (
  id                    integer             primary key,
  token                 string              not null unique,
  created_at            datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
  updated_at            datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
  account_id            integer             not null constraint fk_accounts_password_reset_requests references accounts,
  deleted               boolean             not null default(false),

  constraint chk_token check(token <> '')
);

insert into password_reset_requests select * from password_reset_requests_old;
drop table password_reset_requests_old;