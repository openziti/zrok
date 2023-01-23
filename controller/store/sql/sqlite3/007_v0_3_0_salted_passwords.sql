-- +migrate Up

alter table accounts add column salt string;