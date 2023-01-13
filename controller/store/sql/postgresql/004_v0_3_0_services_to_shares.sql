-- +migrate Up

alter table services rename to shares;
alter sequence services_id_seq1 rename to shares_id_seq1;
alter index services_pkey1 rename to shares_pkey1;
alter index services_token_key rename to shares_token_key;
alter index services_z_id_key1 rename to shares_z_id_key1;
alter table shares rename constraint services_environment_id_fkey to shares_environment_id_fkey;