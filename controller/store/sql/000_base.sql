-- +migrate Up

--
-- credentials
--
create table credentials (
  id            integer             primary key,
  username      string              not null,
  password      string              not null,
  token         string              not null,
  created_at    datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
  updated_at    datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),

  constraint chk_username check (username <> ''),
  constraint chk_password check (username <> ''),
  constraint chk_token check(token <> '')
);
