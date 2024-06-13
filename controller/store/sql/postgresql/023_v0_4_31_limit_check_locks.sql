-- +migrate Up

create table limit_check_locks (
    id                  serial                  primary key,
    account_id          integer                 not null references accounts (id) unique,
    updated_at          timestamptz             not null default(current_timestamp)
);