-- +migrate Up

create table secrets (
    id                  serial                  primary key,

    share_id            integer                 not null references shares(id),
    key                 varchar(255)            not null,
    value               text                    not null,

    created_at          timestamptz             not null default(current_timestamp),
    updated_at          timestamptz             not null default(current_timestamp),
    deleted             boolean                 not null default(false)
);

create index secrets_share_id_idx on secrets(share_id);
create unique index secrets_share_id_key_idx on secrets(share_id, key);