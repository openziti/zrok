-- +migrate Up

-- note: this migration fixes the accounts.email unique constraint for PostgreSQL only.
--
-- for SQLite3, the email column retains its original globally unique constraint
-- (including deleted rows) due to the complexity of recreating the accounts table
-- and all dependent tables with foreign key constraints.
--
-- this means that in SQLite3 deployments, email addresses cannot be reused after
-- soft deletion. this is considered an acceptable limitation for SQLite3 deployments,
-- which are typically used for smaller self-hosted setups, development, and testing
-- environments.
--
-- the inline unique constraint defined in migration 000_base.sql remains in effect:
--   email string not null unique
--
-- to properly fix this for SQLite3 would require:
--   1. dropping all foreign key constraints referencing accounts
--   2. recreating the accounts table without the inline unique constraint
--   3. recreating all dependent tables with their foreign key constraints restored
--   4. creating a partial unique index on email where not deleted
--
-- affected tables: environments, password_reset_requests, frontend_grants,
-- skip_interstitial_grants, organization_members, namespace_grants, names,
-- bandwidth_limit_journal, applied_limit_classes

-- no-op for SQLite3
select 1;

-- +migrate Down

-- no-op for SQLite3
select 1;
