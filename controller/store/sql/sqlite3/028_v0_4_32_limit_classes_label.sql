-- +migrate Up

alter table limit_classes add column label varchar(32);