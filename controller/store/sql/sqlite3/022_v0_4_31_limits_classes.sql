-- +migrate Up

create table limit_classes (
    id                  serial                  primary key,
    limit_scope         string                  not null default ('account'),
    limit_action        string                  not null default ('limit'),
    share_mode          string,
    backend_mode        string,
    period_minutes      integer                 not null default (1440),
    rx_bytes            bigint                  not null default (-1),
    tx_bytes            bigint                  not null default (-1),
    total_bytes         bigint                  not null default (-1),
    created_at          datetime                not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
    updated_at          datetime                not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
    deleted             boolean                 not null default(false)
);

create table applied_limit_classes (
    id                  serial                  primary key,
    account_id          integer                 not null references accounts (id),
    limit_class_id      integer                 not null references limit_classes (id),
    created_at          datetime                not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
    updated_at          datetime                not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
    deleted             boolean                 not null default(false)
);