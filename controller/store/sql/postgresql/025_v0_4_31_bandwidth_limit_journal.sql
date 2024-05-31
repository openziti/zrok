-- +migrate Up

drop table account_limit_journal;
drop table environment_limit_journal;
drop table share_limit_journal;

drop type limit_action_type;
create type limit_action_type as enum ('warning', 'limit');

create table bandwidth_limit_journal (
    id                  serial                  primary key,
    account_id          integer                 references accounts (id) not null,
    limit_class_id      integer                 references limit_classes (id),
    action              limit_action_type       not null,
    rx_bytes            bigint                  not null,
    tx_bytes            bigint                  not null,
    created_at          timestamptz             not null default(current_timestamp),
    updated_at          timestamptz             not null default(current_timestamp)
);

create index bandwidth_limit_journal_account_id_idx on bandwidth_limit_journal (account_id);