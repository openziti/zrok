-- +migrate Up

create table agent_enrollments (
    id                  integer                 primary key,

    environment_id      integer                 not null references environments(id),
    token               varchar(32)             not null unique,

    created_at          datetime                not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
    updated_at          datetime                not null default(strftime('%Y-%m-%d %H:%M:%f', 'now')),
    deleted             boolean                 not null default(false)
);

create index agent_enrollments_environment_id_idx on agent_enrollments(environment_id);