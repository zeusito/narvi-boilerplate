package iam

import (
	"time"

	"github.com/uptrace/bun"
)

type IdentitySession struct {
	bun.BaseModel `bun:"table:identity_sessions,alias:s"`

	ID             string    `bun:"id,pk"`
	IdentityID     string    `bun:"identity_id,notnull"`
	OrganizationID string    `bun:"organization_id"`
	IPAddress      string    `bun:"ip_address,notnull"`
	UserAgent      string    `bun:"user_agent,notnull"`
	ExpiresAt      time.Time `bun:"expires_at,notnull"`
	CreatedAt      time.Time `bun:"created_at,notnull"`
}

type SessionIntrospectionView struct {
	bun.BaseModel `bun:"table:session_introspection_view,alias:siv"`

	SessionID         string    `bun:"session_id"`
	SessionExpiresAt  time.Time `bun:"session_expires_at"`
	IdentityID        string    `bun:"identity_id"`
	IdentityEmail     string    `bun:"identity_email"`
	IdentityFirstName string    `bun:"identity_first_name"`
	IdentityLastName  string    `bun:"identity_last_name"`
	IdentityState     string    `bun:"identity_state"`
	OrganizationID    string    `bun:"organization_id"`
	OrganizationName  string    `bun:"organization_name"`
	OrganizationSlug  string    `bun:"organization_slug"`
	OrganizationLogo  string    `bun:"organization_logo"`
	OrganizationRole  string    `bun:"organization_role"`
}
