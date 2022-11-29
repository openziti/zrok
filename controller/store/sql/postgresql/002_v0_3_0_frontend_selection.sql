-- +migrate Up

create table frontends (
  id                    serial              primary key,
  environment_id        integer             not null references environments(id),
  name                  varchar(32)         not null unique,
  z_id                  varchar(32)         not null unique,
  public_name           varchar(64)         unique,
  reserved              boolean             not null default(false),
  created_at            timestamptz         not null default(current_timestamp),
  updated_at            timestamptz         not null default(current_timestamp)
);

-- environments.account_id should allow NULL; environments with NULL account_id are "ephemeral"
alter table environments drop constraint fk_accounts_identities;
alter table environments add constraint fk_accounts_id foreign key (account_id) references accounts(id);