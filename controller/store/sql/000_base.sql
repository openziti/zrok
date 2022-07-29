-- +migrate Up

--
-- accounts
--
create table accounts (
  id            integer             primary key,
  username      string              not null unique,
  password      string              not null,
  token         string              not null unique,
  created_at    datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
  updated_at    datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),

  constraint chk_username check (username <> ''),
  constraint chk_password check (username <> ''),
  constraint chk_token check(token <> '')
);

--
-- services
--
create table services (
  id            integer             primary key,
  name          string              not null unique,
  created_at    datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
  updated_at    datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
  
  constraint chk_name check (name <> '')
);