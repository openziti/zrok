-- +migrate Up

alter table frontends rename to frontends_old;
alter sequence frontends_id_seq rename to frontends_id_seq_old;

create table frontends (
   id                    serial              primary key,
   environment_id        integer             references environments(id),
   private_share_id      integer             references shares(id),
   token                 varchar(32)         not null unique,
   z_id                  varchar(32)         not null,
   url_template          varchar(1024),
   public_name           varchar(64)         unique,
   reserved              boolean             not null default(false),
   created_at            timestamptz         not null default(current_timestamp),
   updated_at            timestamptz         not null default(current_timestamp),
   deleted               boolean             not null default(false),
);

insert into frontends (id, environment_id, token, z_id, url_template, public_name, reserved, created_at, updated_at, deleted)
    select id, environment_id, token, z_id, url_template, public_name, reserved, created_at, updated_at, deleted from frontends_old;

select setval('frontends_id_seq', (select max(id) from frontends));

drop table frontends_old;

alter index frontends_pkey1 rename to frontends_pkey;
alter index frontends_public_name_key1 to frontends_public_name_key;
alter index frontends_token_key1 to frontends_token_key;

alter table frontends rename constraint frontends_environment_id_fkey1 to frontends_environment_id_fkey;
