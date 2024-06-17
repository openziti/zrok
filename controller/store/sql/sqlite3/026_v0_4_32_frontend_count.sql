-- +migrate Up

alter table limit_classes add column frontends int not null default (-1);