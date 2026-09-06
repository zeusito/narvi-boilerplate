package iam

import (
	"time"

	"github.com/uptrace/bun"
)

type IdentityState string

const (
	IdentityStateActive    IdentityState = "active"
	IdentityStateSuspended IdentityState = "suspended"
	IdentityStateDeleted   IdentityState = "deleted"
	IdentityStateBanned    IdentityState = "banned"
)

type Identity struct {
	bun.BaseModel `bun:"table:identities,alias:i"`

	ID              string        `bun:"id,pk"`
	Email           string        `bun:"email,notnull"`
	FirstName       string        `bun:"first_name,notnull"`
	LastName        string        `bun:"last_name,notnull"`
	State           IdentityState `bun:"state,notnull"`
	EmailVerifiedAt *time.Time    `bun:"email_verified_at"`
	Observations    string        `bun:"observations,notnull"`
	CreatedAt       time.Time     `bun:"created_at,notnull"`
	UpdatedAt       time.Time     `bun:"updated_at,notnull"`
}
