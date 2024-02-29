-- +migrate Up

alter table shares add column permission_mode string not null default('open');

create table access_grants (
    id                  serial                  primary key,
    share_id            integer                 references shares(id),
    account_id          integer                 references accounts(id),
    created_at          datetime                not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
    updated_at          datetime                not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
    deleted             boolean                 not null default(false)
);