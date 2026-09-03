-- migrate:up
insert into organizations(id, name, slug, kind, state, created_at, updated_at)
values ('01a02086-04a2-75a7-ba24-12d5872b8c49', 'Admin Org', 'cautauw9', 'management', 'active', now(), now());

insert into identities(id, email, first_name, last_name, state, created_at, updated_at)
values ('01a02086-04a2-75a7-ba24-1616b586c403', 'admin@example.com', 'Super', 'Admin', 'active', now(), now());

insert into organization_memberships(organization_id, identity_id, roles, created_at, updated_at)
values ('01a02086-04a2-75a7-ba24-12d5872b8c49', '01a02086-04a2-75a7-ba24-1616b586c403', 'owner', now(), now());

-- migrate:down
truncate table organization_memberships;
truncate table identities;
truncate table organizations;
