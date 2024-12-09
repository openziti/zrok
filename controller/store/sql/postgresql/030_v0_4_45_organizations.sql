-- +migrate Up

create table organizations (
    id                  serial                  primary key,

    token               varchar(32)             not null,
    description         varchar(128),

    created_at          timestamptz             not null default(current_timestamp),
    updated_at          timestamptz             not null default(current_timestamp),
    deleted             boolean                 not null default(false)
);

create index organizations_token_idx on organizations(token);

create table organization_members (
    id                  serial                  primary key,

    organization_id     integer                 references organizations(id) not null,
    account_id          integer                 references accounts(id) not null,
    admin               boolean                 not null default(false),

    created_at          timestamptz             not null default(current_timestamp),
    updated_at          timestamptz             not null default(current_timestamp)
);

create index organization_members_account_id_idx on organization_members(account_id);