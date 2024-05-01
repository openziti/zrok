-- +migrate Up

create index environments_account_id_idx on environments (account_id);
create index shares_token_perf_idx on shares (token);
create index shares_environment_id_idx on shares (environment_id);
create index frontends_environment_id_idx on frontends (environment_id);