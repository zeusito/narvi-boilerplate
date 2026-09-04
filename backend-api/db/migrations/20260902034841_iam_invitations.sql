-- migrate:up
create table if not exists invitations (
    id varchar(100) not null,
    kind varchar(50) not null default '',
    email varchar(255) not null default '',
    first_name varchar(255) not null default '',
    last_name varchar(255) not null default '',
    target_id varchar(50) not null,
    inviter_id varchar(50) not null default '',
    member_role varchar(50) not null default '',
    state varchar(50) not null default 'pending',
    expires_at timestamptz not null default now() + interval '1 day',
    created_at timestamptz not null default now(),
    primary key (id),
    foreign key (inviter_id) references identities(id) on delete cascade,
    CONSTRAINT check_invitation_state CHECK (state IN ('pending', 'accepted', 'declined', 'revoked'))
);

CREATE INDEX IF NOT EXISTS idx_invitations_email_state ON invitations(email, state);

-- migrate:down
drop table invitations;
