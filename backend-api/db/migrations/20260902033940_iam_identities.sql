-- migrate:up
create table if not exists identities (
    id varchar(50) not null,
    email varchar(255) not null default '',
    first_name varchar(255) not null default '',
    last_name varchar(255) not null default '',
    state varchar(50) not null default 'active',
    email_verified_at timestamptz null,
    failed_login_attempts int not null default 0,
    lock_expires_at timestamptz not null default now(),
    observations text not null default '',
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    primary key (id),
    CONSTRAINT check_identity_state CHECK (state IN ('active', 'suspended', 'banned', 'deleted'))
);

create unique index if not exists idx_identity_email on identities(lower(email));

create table if not exists identity_accounts (
    id varchar(50) not null,
    identity_id varchar(50) not null,
    kind varchar(50) not null default '',
    provider_id varchar(255) not null default '',
    password_hash varchar(255) not null default '',
    created_at timestamptz not null default now(),
    primary key (id),
    foreign key (identity_id) references identities(id) on delete cascade
);

create unique index if not exists idx_identity_accounts_provider on identity_accounts(lower(kind), lower(provider_id));

create table if not exists verifications (
    id varchar(100) not null,
    identity_id varchar(50) not null,
    kind varchar(50) not null default '',
    attempts int not null default 0,
    expires_at timestamptz not null default now() + interval '10 minutes',
    created_at timestamptz not null default now(),
    primary key (id),
    foreign key (identity_id) references identities(id) on delete cascade
);

-- migrate:down
drop table if exists verifications;
drop table if exists identity_accounts;
drop table if exists identities;
