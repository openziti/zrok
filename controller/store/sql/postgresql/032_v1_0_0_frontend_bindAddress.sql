-- +migrate Up

alter table frontends add column bind_address varchar(128);