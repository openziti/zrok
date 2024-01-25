-- +migrate Up

-- remove the old unique index (users might need multiple password resets)
ALTER TABLE password_reset_requests DROP CONSTRAINT password_reset_requests_account_id_key;

-- add new constraint which doesnt mind having multiple resets for account ids
ALTER TABLE password_reset_requests ADD CONSTRAINT password_reset_requests_account_id_key FOREIGN KEY (account_id) REFERENCES accounts (id) ON DELETE CASCADE;
