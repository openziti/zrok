-- +migrate Up

alter table frontends add column private_share_id references shares(id);
