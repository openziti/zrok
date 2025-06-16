-- +migrate Up

create table secrets (
    id                  integer                 primary key,

    share_id            integer                 not null references shares(id),
    key                 varchar(255)            not null,
    value               text                    not null,

    created_at          datetime                not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
    updated_at          datetime                not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
    deleted             boolean                 not null default(false)
);

create index secrets_share_id_idx on secrets(share_id);
create unique index secrets_share_id_key_idx on secrets(share_id, key);