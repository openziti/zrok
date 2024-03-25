-- +migrate Up

-- remove the old unique index (which did not respect the deleted flag)
ALTER TABLE shares DROP CONSTRAINT shares_token_key;

-- add a new unique index which only constrains uniqueness for not-deleted rows
CREATE UNIQUE INDEX shares_token_idx ON shares(token) WHERE deleted is false;
