-- +migrate Up

create table skip_interstitial_grants (
    id                  serial                  primary key,

    account_id          integer                 references accounts (id) not null,

    created_at          timestamptz             not null default(current_timestamp),
    updated_at          timestamptz             not null default(current_timestamp),
    deleted             boolean                 not null default(false)
);

create index skip_interstitial_grants_id_idx on skip_interstitial_grants (account_id);