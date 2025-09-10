-- +migrate Up

--
-- frontend_mappings
--
create table frontend_mappings (
    name                    varchar(256)        not null,
    version                 bigint              not null,
    share_token             varchar(32)         not null,
    created_at              timestamptz         not null default(current_timestamp),
    primary key (name, version)
);

create index frontend_mappings_share_token_idx on frontend_mappings (share_token);

-- +migrate Down

drop table if exists frontend_mappings;