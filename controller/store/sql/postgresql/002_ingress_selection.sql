-- +migrate Up

create type availability_type as enum ('public', 'private');

create table frontends (
  id                    serial              primary key,
  environment_id        integer             references environments(id),
  z_id                  varchar(32)         not null unique,
  type                  availability_type   not null,
  created_at            timestamptz         not null default(current_timestamp),
  updated_at            timestamptz         not null default(current_timestamp)
);

-- environments.account_id should allow NULL; environments with NULL account_id are "ephemeral"
alter table environments drop constraint fk_accounts_identities;
alter table environments add constraint fk_accounts_id foreign key (account_id) references accounts(id);