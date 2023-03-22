-- +migrate Up

create table account_limit_journal (
    id                  integer                 primary key,
    account_id          integer                 references accounts(id),
    rx_bytes            bigint                  not null,
    tx_bytes            bigint                  not null,
    action              limit_action_type       not null,
    created_at          datetime                not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
    updated_at          datetime                not null default(strftime('%Y-%m-%d %H:%M:%f', 'now'))
);

create table environment_limit_journal (
    id                  integer                 primary key,
    environment_id      integer                 references environments(id),
    rx_bytes            bigint                  not null,
    tx_bytes            bigint                  not null,
    action              limit_action_type       not null,
    created_at          datetime                not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
    updated_at          datetime                not null default(strftime('%Y-%m-%d %H:%M:%f', 'now'))
);

create table share_limit_journal (
    id                  integer                 primary key,
    share_id            integer                 references shares(id),
    rx_bytes            bigint                  not null,
    tx_bytes            bigint                  not null,
    action              limit_action_type       not null,
    created_at          datetime                not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
    updated_at          datetime                not null default(strftime('%Y-%m-%d %H:%M:%f', 'now'))
);