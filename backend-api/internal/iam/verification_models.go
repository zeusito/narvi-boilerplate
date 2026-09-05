package iam

import (
	"time"

	"github.com/uptrace/bun"
)

const (
	VerificationKindEmailOTP = "email_otp"
)

type Verification struct {
	bun.BaseModel `bun:"table:verifications,alias:v"`

	ID         string    `bun:"id,pk"`
	IdentityID string    `bun:"identity_id,notnull"`
	Kind       string    `bun:"kind,notnull"`
	HashedCode string    `bun:"hashed_code,notnull"`
	Attempts   int       `bun:"attempts,notnull"`
	ExpiresAt  time.Time `bun:"expires_at,notnull"`
	CreatedAt  time.Time `bun:"created_at,notnull"`
}
