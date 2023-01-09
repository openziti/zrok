-- +migrate Up

--
-- invites
---

create table invites (
  id                    integer             primary key,
  token                 string              not null unique,
  created_at            datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
  updated_at            datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),

  constraint chk_token check(token <> '')
);