-- +migrate Up

ALTER TABLE account_requests DROP CONSTRAINT account_requests_email_key;