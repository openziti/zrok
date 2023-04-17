-- +migrate Up

alter type backend_mode rename value 'dav' to 'tunnel';