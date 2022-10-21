-- +migrate Up

--
-- accounts
--
create table accounts (
  id                    serial              primary key,
  email                 varchar(1024)       not null unique,
  password              char(128)           not null,
  token                 varchar(32)         not null unique,
  created_at            timestamp           not null default(current_timestamp),
  updated_at            timestamp           not null default(current_timestamp),

  constraint chk_email check (email <> ''),
  constraint chk_password check (password <> ''),
  constraint chk_token check(token <> '')
);

--
-- account_requests
--
create table account_requests (
  id                    serial              primary key,
  token                 varchar(32)         not null unique,
  email                 varchar(1024)       not null unique,
  source_address        varchar(64)         not null,
  created_at            timestamp           not null default(current_timestamp),
  updated_at            timestamp           not null default(current_timestamp)
);

--
-- environments
--
create table environments (
  id                    serial              primary key,
  account_id            integer             constraint fk_accounts_identities references accounts on delete cascade,
  description           text,
  host                  varchar(256),
  address               varchar(64),
  z_id                  varchar(32)         not null unique,
  created_at            timestamp           not null default(current_timestamp),
  updated_at            timestamp           not null default(current_timestamp),

  constraint chk_z_id check (z_id <> '')
);

--
-- services
--
create table services (
  id                    serial              primary key,
  environment_id        integer             constraint fk_environments_services references environments on delete cascade,
  z_id                  varchar(32)         not null unique,
  name                  varchar(32)         not null unique,
  frontend              varchar(1024),
  backend               varchar(1024),
  created_at            timestamp           not null default(current_timestamp),
  updated_at            timestamp           not null default(current_timestamp),

  constraint chk_z_id check (z_id <> ''),
  constraint chk_name check (name <> '')
);