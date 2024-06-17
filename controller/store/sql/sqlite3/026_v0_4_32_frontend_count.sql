-- +migrate Up

alter table limit_classes add column share_frontends int not null default (-1);