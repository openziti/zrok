-- +migrate up

alter table accounts add column salt string;