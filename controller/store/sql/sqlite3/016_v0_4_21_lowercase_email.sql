-- +migrate Up

update accounts set email = lower(email);