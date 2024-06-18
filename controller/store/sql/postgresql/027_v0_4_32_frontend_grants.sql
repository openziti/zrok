-- +migrate Up

alter table frontends add column permission_mode permission_mode_type not null default('open');

create table frontend_grants (
    id                  serial                  primary key,

    account_id          integer                 references accounts (id) not null,
    frontend_id         integer                 references frontends (id) not null,

    created_at          timestamptz             not null default(current_timestamp),
    updated_at          timestamptz             not null default(current_timestamp),
    deleted             boolean                 not null default(false)
);

create index frontend_grants_account_id_idx on frontend_grants (account_id);
create index frontend_grants_frontend_id_idx on frontend_grants (frontend_id);