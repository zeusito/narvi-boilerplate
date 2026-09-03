package iam

import (
	"time"

	"github.com/uptrace/bun"
)

type Identity struct {
	bun.BaseModel `bun:"table:identities,alias:i"`

	ID                  string     `bun:"id,pk"`
	Email               string     `bun:"email,notnull"`
	FirstName           string     `bun:"first_name,notnull"`
	LastName            string     `bun:"last_name,notnull"`
	State               string     `bun:"state,notnull"`
	EmailVerifiedAt     *time.Time `bun:"email_verified_at"`
	FailedLoginAttempts int        `bun:"failed_login_attempts,notnull"`
	LockExpiresAt       time.Time  `bun:"lock_expires_at,notnull"`
	Observations        string     `bun:"observations,notnull"`
	CreatedAt           time.Time  `bun:"created_at,notnull"`
	UpdatedAt           time.Time  `bun:"updated_at,notnull"`
}

type IdentityMembership struct {
	OrganizationID string    `bun:"organization_id"`
	Role           string    `bun:"role"`
	CreatedAt      time.Time `bun:"created_at"`
}
