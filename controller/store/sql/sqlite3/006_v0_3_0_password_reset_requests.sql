-- +migrate Up

--
-- password_reset_requests
---

create table password_reset_requests (
  id                    integer             primary key,
  token                 string              not null unique,
  created_at            datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
  updated_at            datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
  account_id            integer             not null unique constraint fk_accounts_password_reset_requests references accounts on delete cascade,

  constraint chk_token check(token <> '')
);