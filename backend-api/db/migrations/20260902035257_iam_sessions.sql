-- migrate:up
create table if not exists identity_sessions (
    id varchar(100) not null,
    identity_id varchar(50) not null,
    organization_id varchar(50) null,
    ip_address varchar(50) not null default '',
    user_agent text not null default '',
    expires_at timestamptz not null default now() + interval '1 hour',
    created_at timestamptz not null default now(),
    primary key (id),
    foreign key (identity_id) references identities(id) on delete cascade,
    foreign key (organization_id) references organizations(id) on delete cascade
);

create index if not exists idx_identity_sessions_identity_id on identity_sessions(identity_id);
create index if not exists idx_identity_sessions_expires_at on identity_sessions(expires_at);

create or replace view session_introspection_view as (
    SELECT
       s.id as session_id,
       s.expires_at as session_expires_at,
       s.identity_id as identity_id,
       i.email as identity_email,
       i.first_name as identity_first_name,
       i.last_name as identity_last_name,
       i.state as identity_state,
       COALESCE(s.organization_id, '') AS organization_id,
       COALESCE(o.name, '') AS organization_name,
       COALESCE(o.slug, '') AS organization_slug,
       COALESCE(o.logo, '') AS organization_logo,
       COALESCE(om.roles, '') AS organization_role
    FROM identity_sessions s
    JOIN identities i
       ON i.id = s.identity_id
    LEFT JOIN organizations o
       ON o.id = s.organization_id
    LEFT JOIN organization_memberships om
       ON om.organization_id = s.organization_id
      AND om.identity_id = s.identity_id
);


-- migrate:down
drop view if exists session_introspection_view;
drop table if exists identity_sessions;
