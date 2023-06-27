-- +migrate Up

alter table account_requests                add column deleted boolean not null default(false);
alter table accounts                        add column deleted boolean not null default(false);
alter table environments                    add column deleted boolean not null default(false);
alter table frontends                       add column deleted boolean not null default(false);
alter table invite_tokens                   add column deleted boolean not null default(false);
alter table password_reset_requests         add column deleted boolean not null default(false);
alter table shares                          add column deleted boolean not null default(false);