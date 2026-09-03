-- migrate:up
create table if not exists organizations (
    id varchar(50) not null,
    name varchar(255) not null,
    slug varchar(50) not null default '',
    kind varchar(50) not null default '',
    logo varchar(500) not null default '',
    state varchar(50) not null default 'active',
    observations text not null default '',
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    primary key (id),
    CONSTRAINT check_org_state CHECK (state IN ('active', 'suspended', 'deleted'))
);

create unique index if not exists idx_orgs_slug on organizations(lower(slug));

create table if not exists organization_memberships (
    organization_id varchar(50) not null,
    identity_id varchar(50) not null,
    roles varchar(50) not null default '',
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    primary key (organization_id, identity_id),
    foreign key (organization_id) references organizations(id) on delete cascade,
    foreign key (identity_id) references identities(id) on delete cascade
);

create or replace view organization_members_view as (
    SELECT i.id AS identity_id,
       i.email AS identity_email,
       (i.email_verified_at IS NOT NULL) AS identity_email_verified,
       i.first_name AS identity_first_name,
       i.last_name AS identity_last_name,
       i.state AS identity_state,
       i.created_at AS identity_created_at,
       i.updated_at AS identity_updated_at,
       COALESCE(o.id, '') AS organization_id,
       COALESCE(o.name, '') AS organization_name,
       COALESCE(o.slug, '') AS organization_slug,
       COALESCE(o.logo, '') AS organization_logo,
       COALESCE(o.kind, '') AS organization_kind,
       COALESCE(o.state, '') AS organization_state,
       COALESCE(om.roles, '') AS organization_role,
       COALESCE(om.created_at, '0001-01-01'::timestamptz) AS membership_created_at,
       COALESCE(om.updated_at, '0001-01-01'::timestamptz) AS membership_updated_at
    FROM identities i
    LEFT JOIN organization_memberships om
       ON om.identity_id = i.id
    LEFT JOIN organizations o
       ON o.id = om.organization_id
);


-- migrate:down
drop view organization_members_view;
drop table organization_memberships;
drop table organizations;
