-- +migrate Up

--
-- frontend_mappings
--
create table frontend_mappings (
    id                      bigserial           primary key,
    frontend_token          varchar(32)         not null,
    name                    varchar(256)        not null,
    share_token             varchar(32)         not null,
    created_at              timestamptz         not null default(current_timestamp),
    unique (frontend_token, name)
);

create index frontend_mappings_share_token_idx on frontend_mappings (share_token);

-- +migrate Down

drop table if exists frontend_mappings;