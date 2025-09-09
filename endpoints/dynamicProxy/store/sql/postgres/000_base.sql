-- +migrate Up

--
-- mappings
--
create table mappings (
    mapping             varchar(256)        not null,
    version             bigint              not null,
    share_token         varchar(32)         not null,
    created_at          timestamp           not null default(current_timestamp),
    primary key (mapping, version)
);