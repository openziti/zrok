-- +migrate Up

alter table accounts add column limitless boolean not null default(false);