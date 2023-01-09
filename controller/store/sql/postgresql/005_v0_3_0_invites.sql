-- +migrate Up

--
-- invites
---

create table invites (
  id                    serial                primary key,
  token                 varchar(32)           not null unique,
  created_at            timestamptz           not null default(current_timestamp),
  updated_at            timestamptz           not null default(current_timestamp),

  constraint chk_token check(token <> '')
);