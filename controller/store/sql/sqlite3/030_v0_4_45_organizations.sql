-- +migrate Up

create table organizations (
    id                  integer                 primary key,

    token               varchar(32)             not null,
    description         varchar(128),

    created_at          datetime                not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
    updated_at          datetime                not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
    deleted             boolean                 not null default(false)
);

create index organization_token_idx on organizations(token);

create table organization_members (
    id                  integer                 primary key,

    organization_id     integer                 references organizations(id) not null,
    account_id          integer                 references accounts(id) not null,
    admin               boolean                 not null default(false),

    created_at          datetime                not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
    updated_at          datetime                not null default(strftime('%Y-%m-%d %H:%M:%f', 'now'))
);

create index organization_members_account_id_idx on organization_members(account_id);