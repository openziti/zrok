-- +migrate Up

drop table account_limit_journal;
drop table environment_limit_journal;
drop table share_limit_journal;

create table bandwidth_limit_journal (
    id                  serial                  primary key,
    account_id          integer                 references accounts (id) not null,
    limit_class         integer                 references limit_classes,
    action              string                  not null,
    rx_bytes            bigint                  not null,
    tx_bytes            bigint                  not null,
    created_at          timestamptz             not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
    updated_at          timestamptz             not null default(strftime('%Y-%m-%d %H:%M:%f', 'now'))
);

create index bandwidth_limit_journal_account_id_idx on bandwidth_limit_journal (account_id);