-- +migrate Up

alter table accounts alter column created_at type timestamptz;
alter table accounts alter column updated_at type timestamptz;

alter table account_requests alter column created_at type timestamptz;
alter table account_requests alter column updated_at type timestamptz;

alter table environments alter column created_at type timestamptz;
alter table environments alter column updated_at type timestamptz;

alter table services alter column created_at type timestamptz;
alter table services alter column updated_at type timestamptz;