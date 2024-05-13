-- +migrate Up

create type limit_scope as enum ('account', 'environment', 'share');
create type limit_action as enum ('warning', 'limit');

create table limit_classes (
    id                  serial                  primary key,
    limit_scope         limit_scope             not null default ('account'),
    limit_action        limit_action            not null default ('limit'),
    share_mode          share_mode,
    backend_mode        backend_mode,
    period_minutes      int                     not null default (1440),
    rx_bytes            bigint                  not null default (-1),
    tx_bytes            bigint                  not null default (-1),
    total_bytes         bigint                  not null default (-1),
    created_at          timestamptz             not null default(current_timestamp),
    updated_at          timestamptz             not null default(current_timestamp),
    deleted             boolean                 not null default(false)
);

create table applied_limit_classes (
    id                  serial                  primary key,
    account_id          integer                 not null references accounts (id),
    limit_class_id      integer                 not null references limit_classes (id),
    created_at          timestamptz             not null default(current_timestamp),
    updated_at          timestamptz             not null default(current_timestamp),
    deleted             boolean                 not null default(false)
);