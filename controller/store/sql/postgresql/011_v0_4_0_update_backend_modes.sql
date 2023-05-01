-- +migrate Up

alter type backend_mode rename value 'dav' to 'tcpTunnel';
alter type backend_mode add value 'udpTunnel';