-- +migrate Up

--
-- namespace_frontend_mappings
--
create table namespace_frontend_mappings (
  id                    integer             primary key,
  namespace_id          integer             not null constraint fk_namespace_frontend_mappings_namespaces references namespaces on delete cascade,
  frontend_id           integer             not null constraint fk_namespace_frontend_mappings_frontends references frontends on delete cascade,
  created_at            datetime            not null default(current_timestamp),
  updated_at            datetime            not null default(current_timestamp),
  deleted               boolean             not null default(false)
);

create unique index uk_namespace_frontend_mappings on namespace_frontend_mappings(namespace_id, frontend_id) where not deleted;

-- +migrate Down

drop index if exists uk_namespace_frontend_mappings;
drop table if exists namespace_frontend_mappings;