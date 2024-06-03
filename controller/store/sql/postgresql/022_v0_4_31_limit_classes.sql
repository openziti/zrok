-- +migrate Up

create type limit_action as enum ('warning', 'limit');

create table limit_classes (
    id                  serial                  primary key,

    share_mode          share_mode,
    backend_mode        backend_mode,

    environments        int                     not null default (-1),
    shares              int                     not null default (-1),
    reserved_shares     int                     not null default (-1),
    unique_names        int                     not null default (-1),
    period_minutes      int                     not null default (1440),
    rx_bytes            bigint                  not null default (-1),
    tx_bytes            bigint                  not null default (-1),
    total_bytes         bigint                  not null default (-1),

    limit_action        limit_action            not null default ('limit'),

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