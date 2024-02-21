-- +migrate Up

ALTER TABLE accounts ADD COLUMN disabled BOOLEAN NOT NULL DEFAULT(false);