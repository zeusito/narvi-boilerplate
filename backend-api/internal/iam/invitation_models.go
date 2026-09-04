package iam

import (
	"time"

	"github.com/uptrace/bun"
)

type InvitationKind string

const (
	InvitationKindOrgMember InvitationKind = "org_member"
)

type Invitation struct {
	bun.BaseModel `bun:"table:invitations,alias:inv"`

	ID         string         `bun:"id,pk"`
	Kind       InvitationKind `bun:"kind,notnull"`
	Email      string         `bun:"email,notnull"`
	FirstName  string         `bun:"first_name,notnull"`
	LastName   string         `bun:"last_name,notnull"`
	TargetID   string         `bun:"target_id,notnull"`
	InviterID  string         `bun:"inviter_id,notnull"`
	MemberRole string         `bun:"member_role,notnull"`
	State      string         `bun:"state,notnull"`
	ExpiresAt  time.Time      `bun:"expires_at,notnull"`
	CreatedAt  time.Time      `bun:"created_at,notnull"`
}
