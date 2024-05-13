-- +migrate Up

create table limits_classes (
    id                  serial                  primary key,
    limit_scope         string                  not null default ('account'),
    limit_action        string                  not null default ('limit'),
    share_mode          string,
    backend_mode        string,
    period_minutes      int                     not null default (1440),
    rx_bytes            bigint                  not null default (-1),
    tx_bytes            bigint                  not null default (-1),
    total_bytes         bigint                  not null default (-1),
    created_at          timestamptz             not null default(current_timestamp),
    updated_at          timestamptz             not null default(current_timestamp),
    deleted             boolean                 not null default(false)
)