-- +migrate Up

create table limit_classes (
    id                  integer                  primary key,

    backend_mode        string,

    environments        integer                 not null default (-1),
    shares              integer                 not null default (-1),
    reserved_shares     integer                 not null default (-1),
    unique_names        integer                 not null default (-1),
    period_minutes      integer                 not null default (1440),
    rx_bytes            bigint                  not null default (-1),
    tx_bytes            bigint                  not null default (-1),
    total_bytes         bigint                  not null default (-1),

    limit_action        string                  not null default ('limit'),

    created_at          datetime                not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
    updated_at          datetime                not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
    deleted             boolean                 not null default(false)
);

create table applied_limit_classes (
    id                  integer                 primary key,
    account_id          integer                 not null references accounts (id),
    limit_class_id      integer                 not null references limit_classes (id),
    created_at          datetime                not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
    updated_at          datetime                not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
    deleted             boolean                 not null default(false)
);

create index applied_limit_classes_account_id_idx on applied_limit_classes (account_id);
create index applied_limit_classes_limit_class_id_idx on applied_limit_classes (limit_class_id);