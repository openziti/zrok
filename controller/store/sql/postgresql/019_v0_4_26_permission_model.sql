-- +migrate Up

create type permission_mode_type as enum('open', 'closed');

alter table shares add column permission_mode permission_mode_type not null default('open');

create table access_grants (
    id                  serial                  primary key,
    share_id            integer                 references shares(id),
    account_id          integer                 references accounts(id),
    created_at          timestamptz             not null default(current_timestamp),
    updated_at          timestamptz             not null default(current_timestamp),
    deleted             boolean                 not null default(false)
);