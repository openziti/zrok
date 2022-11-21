-- +migrate Up

create type share_mode as enum ('public', 'private');
create type backend_mode as enum ('proxy', 'web', 'dav');

alter table services
    add column share_mode share_mode not null default 'public',
    add column backend_mode backend_mode not null default 'proxy';

alter table services
    alter column share_mode drop default;
alter table services
    alter column backend_mode drop default;