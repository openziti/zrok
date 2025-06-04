-- +migrate Up

create table agent_enrollments (
    id                  serial                  primary key,

    environment_id      integer                 not null references environments(id),
    token               varchar(32)             not null unique,

    created_at          timestamptz             not null default(current_timestamp),
    updated_at          timestamptz             not null default(current_timestamp),
    deleted             boolean                 not null default(false)
);

create index agent_enrollments_environment_id_idx on agent_enrollments(environment_id);