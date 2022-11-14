-- +migrate Up

create table ingress (
  id                    serial              primary key,
  account_id            integer             references accounts(id),
  z_id                  varchar(32)         not null unique,
  created_at            timestamptz         not null default(current_timestamp),
  updated_at            timestamptz         not null default(current_timestamp)
);