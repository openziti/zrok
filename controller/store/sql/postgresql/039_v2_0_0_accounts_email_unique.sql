-- +migrate Up

-- drop the existing unique constraint on the email column
alter table accounts drop constraint if exists accounts_email_key;

-- create partial unique index on email excluding deleted records
create unique index uk_accounts_email on accounts(email) where not deleted;

-- +migrate Down

-- drop the partial unique index
drop index if exists uk_accounts_email;

-- recreate the original unique constraint
alter table accounts add constraint accounts_email_key unique (email);
