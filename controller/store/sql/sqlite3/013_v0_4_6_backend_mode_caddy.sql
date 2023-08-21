-- +migrate Up

alter table shares rename to shares_old;
create table shares (
    id                        integer             primary key,
    environment_id            integer             constraint fk_environments_shares references environments on delete cascade,
    z_id                      string              not null unique,
    token                     string              not null unique,
    share_mode                string              not null,
    backend_mode              string              not null,
    frontend_selection        string,
    frontend_endpoint         string,
    backend_proxy_endpoint    string,
    reserved                  boolean             not null default(false),
    created_at                datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
    updated_at                datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')), deleted boolean not null default(false),

    constraint chk_z_id check (z_id <> ''),
    constraint chk_token check (token <> ''),
    constraint chk_share_mode check (share_mode == 'public' or share_mode == 'private'),
    constraint chk_backend_mode check (backend_mode == 'proxy' or backend_mode == 'web' or backend_mode == 'tcpTunnel' or backend_mode == 'udpTunnel' or backend_mode == 'caddy')
);
insert into shares select * from shares_old;
drop table shares_old;

alter table frontends rename to frontends_old;
create table frontends (
   id                    integer             primary key,
   environment_id        integer             references environments(id),
   token                 varchar(32)         not null unique,
   z_id                  varchar(32)         not null,
   public_name           varchar(64)         unique,
   url_template          varchar(1024),
   reserved              boolean             not null default(false),
   created_at            datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
   updated_at            datetime            not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
   deleted               boolean             not null default(false),
   private_share_id      integer             references shares(id)
);
insert into frontends select * from frontends_old;
drop table frontends_old;

alter table share_limit_journal rename to share_limit_journal_old;
create table share_limit_journal (
    id                  integer                 primary key,
    share_id            integer                 references shares(id),
    rx_bytes            bigint                  not null,
    tx_bytes            bigint                  not null,
    action              limit_action_type       not null,
    created_at          datetime                not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
    updated_at          datetime                not null default(strftime('%Y-%m-%d %H:%M:%f', 'now'))
);
insert into share_limit_journal select * from share_limit_journal_old;
drop table share_limit_journal_old;