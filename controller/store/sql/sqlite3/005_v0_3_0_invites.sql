-- +migrate Up

--
-- invites
---

create table invites (
  id                    serial              primary key,
  token                 varchar(32)         not null unique,
  token_status          varchar(1024)       not null unique,
  created_at            timestamp           not null default(current_timestamp),
  updated_at            timestamp           not null default(current_timestamp),

  constraint chk_token check(token <> ''),
  constraint chk_status check(token_status <> '')
);