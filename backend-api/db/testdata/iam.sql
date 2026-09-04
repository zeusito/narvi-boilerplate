insert into organizations(id, name, slug, kind, state, created_at, updated_at)
values ('01a02086-04a2-75a7-ba24-12d5872b8c49', 'Acme Corp', 'acme', 'management', 'active', now(), now());

insert into identities(id, email, first_name, last_name, state, created_at, updated_at)
values ('01a02086-04a2-75a7-ba24-1616b586c403', 'admin@example.com', 'Super', 'Admin', 'active', now(), now());

insert into organization_memberships(organization_id, identity_id, member_role, created_at, updated_at)
values ('01a02086-04a2-75a7-ba24-12d5872b8c49', '01a02086-04a2-75a7-ba24-1616b586c403', 'owner', now(), now());

-- User with no memberships but an active pending invitation
insert into identities(id, email, first_name, last_name, state, created_at, updated_at)
values ('01a02086-04a2-75a7-ba24-1616b586c404', 'invited@example.com', 'Invited', 'User', 'active', now(), now());

insert into invitations(id, kind, email, first_name, last_name, target_id, inviter_id, member_role, state, expires_at, created_at)
values ('inv_test1', 'organization_member', 'invited@example.com', 'Invited', 'User', '01a02086-04a2-75a7-ba24-12d5872b8c49', '01a02086-04a2-75a7-ba24-1616b586c403', 'member', 'pending', now() + interval '1 day', now());
