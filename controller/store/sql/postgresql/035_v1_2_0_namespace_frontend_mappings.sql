-- +migrate Up

--
-- namespace_frontend_mappings
--
create table namespace_frontend_mappings (
  id                    serial              primary key,
  namespace_id          integer             not null constraint fk_namespace_frontend_mappings_namespaces references namespaces on delete cascade,
  frontend_id           integer             not null constraint fk_namespace_frontend_mappings_frontends references frontends on delete cascade,
  created_at            timestamp           not null default(current_timestamp),
  updated_at            timestamp           not null default(current_timestamp),
  deleted               boolean             not null default(false),

  constraint uk_namespace_frontend_mappings unique (namespace_id, frontend_id) where not deleted
);

-- +migrate Down

drop table if exists namespace_frontend_mappings;