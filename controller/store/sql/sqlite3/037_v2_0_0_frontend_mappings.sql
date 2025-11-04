-- +migrate Up

--
-- frontend_mappings
--
create table frontend_mappings (
    id                      integer             primary key autoincrement,
    frontend_token          string              not null,
    name                    string              not null,
    share_token             string              not null,
    created_at              datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
    unique (frontend_token, name)
);

create index frontend_mappings_share_token_idx on frontend_mappings (share_token);

-- +migrate Down

drop index if exists frontend_mappings_share_token_idx;
drop table if exists frontend_mappings;