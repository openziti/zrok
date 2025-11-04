-- +migrate Up

-- drop the existing unique constraint on public_name
alter table frontends drop constraint if exists frontends_public_name_key;
alter table frontends drop constraint if exists frontends_public_name_key1;

-- create partial unique index on public_name excluding deleted records
create unique index uk_frontends_public_name on frontends(public_name) where not deleted;

-- +migrate Down

-- drop the partial unique index
drop index if exists uk_frontends_public_name;

-- recreate the original unique constraint on public_name
alter table frontends add constraint frontends_public_name_key unique (public_name);
