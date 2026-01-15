-- +migrate Up

-- MFA configuration for accounts
create table account_mfa (
    id                    serial primary key,
    account_id            integer not null unique references accounts(id),
    totp_secret           varchar(128) not null,
    enabled               boolean not null default(false),
    created_at            timestamptz not null default(current_timestamp),
    updated_at            timestamptz not null default(current_timestamp)
);

-- Recovery codes (hashed)
create table mfa_recovery_codes (
    id                    serial primary key,
    account_mfa_id        integer not null references account_mfa(id) on delete cascade,
    code_hash             varchar(128) not null,
    used                  boolean not null default(false),
    used_at               timestamptz,
    created_at            timestamptz not null default(current_timestamp)
);

-- Pending MFA sessions for two-step login
create table mfa_pending_auth (
    id                    serial primary key,
    account_id            integer not null references accounts(id),
    pending_token         varchar(64) not null unique,
    expires_at            timestamptz not null,
    created_at            timestamptz not null default(current_timestamp)
);

-- MFA challenge tokens for step-up authentication
create table mfa_challenge_tokens (
    id                    serial primary key,
    account_id            integer not null references accounts(id),
    challenge_token       varchar(64) not null unique,
    expires_at            timestamptz not null,
    created_at            timestamptz not null default(current_timestamp)
);

-- Index for efficient cleanup of expired tokens
create index idx_mfa_pending_auth_expires_at on mfa_pending_auth(expires_at);
create index idx_mfa_challenge_tokens_expires_at on mfa_challenge_tokens(expires_at);

-- +migrate Down

drop index if exists idx_mfa_challenge_tokens_expires_at;
drop index if exists idx_mfa_pending_auth_expires_at;
drop table if exists mfa_challenge_tokens;
drop table if exists mfa_pending_auth;
drop table if exists mfa_recovery_codes;
drop table if exists account_mfa;
