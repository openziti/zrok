-- +migrate Up

create type limit_action_type as enum ('clear', 'warning', 'limit');

create table account_limit_journal (
    id                  serial                  primary key,
    account_id          integer                 references accounts(id),
    rx_bytes            bigint                  not null,
    tx_bytes            bigint                  not null,
    action              limit_action_type       not null,
    created_at          timestamptz             not null default(current_timestamp),
    updated_at          timestamptz             not null default(current_timestamp)
);

create table environment_limit_journal (
    id                  serial                  primary key,
    environment_id      integer                 references environments(id),
    rx_bytes            bigint                  not null,
    tx_bytes            bigint                  not null,
    action              limit_action_type       not null,
    created_at          timestamptz             not null default(current_timestamp),
    updated_at          timestamptz             not null default(current_timestamp)
);

create table share_limit_journal (
    id                  serial                  primary key,
    share_id            integer                 references shares(id),
    rx_bytes            bigint                  not null,
    tx_bytes            bigint                  not null,
    action              limit_action_type       not null,
    created_at          timestamptz             not null default(current_timestamp),
    updated_at          timestamptz             not null default(current_timestamp)
);