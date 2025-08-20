-- +migrate Up

--
-- controllers
--
create table controllers (
    id                  serial              primary key,
    address             varchar(128)        not null unique,
    created_at          timestamp           not null default(current_timestamp)
);

--
-- mappings
--
create table mappings (
    id                  serial              primary key,
    mapping             varchar(256)        not null unique,
    version             int                 not null,
    share_token         varchar(32)         not null,
    created_at          timestamp           not null default(current_timestamp)
);