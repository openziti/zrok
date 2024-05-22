-- +migrate Up

alter table shares add column unique_name boolean not null default (false);