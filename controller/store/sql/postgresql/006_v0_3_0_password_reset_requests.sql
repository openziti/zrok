-- +migrate up

--
-- password_reset_requests
---

create table password_reset_requests (
  id                    serial                primary key,
  token                 varchar(32)           not null unique,
  created_at            timestamptz           not null default(current_timestamp),
  updated_at            timestamptz           not null default(current_timestamp),
  account_id            integer               not null unique constraint fk_accounts_password_reset_requests references accounts on delete cascade,

  constraint chk_token check(token <> '')
);