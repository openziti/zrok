-- +migrate Up

create table skip_interstitial_grants (
    id                  integer                 primary key,

    account_id          integer                 references accounts (id) not null,

    created_at          datetime                not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
    updated_at          datetime                not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
    deleted             boolean                 not null default(false)
);

create index skip_interstitial_grants_id_idx on skip_interstitial_grants (account_id);