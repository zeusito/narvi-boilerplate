package iam

import (
	"time"

	"github.com/uptrace/bun"
)

type OrganizationState string
type OrganizationKind string
type OrganizationMemberRole string

const (
	OrganizationStateActive    OrganizationState = "active"
	OrganizationStateSuspended OrganizationState = "suspended"
	OrganizationStateDeleted   OrganizationState = "deleted"

	OrganizationKindStandard   OrganizationKind = "standard"
	OrganizationKindManagement OrganizationKind = "management"

	OrganizationMemberRoleOwner  OrganizationMemberRole = "owner"
	OrganizationMemberRoleAdmin  OrganizationMemberRole = "admin"
	OrganizationMemberRoleMember OrganizationMemberRole = "member"
)

type Organization struct {
	bun.BaseModel `bun:"table:organizations,alias:o"`

	ID           string            `bun:"id,pk"`
	Name         string            `bun:"name,notnull"`
	Slug         string            `bun:"slug,notnull"`
	Kind         OrganizationKind  `bun:"kind,notnull"`
	Logo         string            `bun:"logo"`
	State        OrganizationState `bun:"state,notnull"`
	Observations string            `bun:"observations"`
	CreatedAt    time.Time         `bun:"created_at,notnull"`
	UpdatedAt    time.Time         `bun:"updated_at,notnull"`
}

type OrganizationMembership struct {
	bun.BaseModel `bun:"table:organization_memberships,alias:om"`

	OrganizationID string    `bun:"organization_id,notnull"`
	IdentityID     string    `bun:"identity_id,notnull"`
	MemberRole     string    `bun:"member_role,notnull"`
	CreatedAt      time.Time `bun:"created_at,notnull"`
	UpdatedAt      time.Time `bun:"updated_at,notnull"`
}

type OrganizationMembershipView struct {
	bun.BaseModel `bun:"table:organization_members_view,alias:omv"`

	IdentityID          string                 `bun:"identity_id,notnull"`
	IdentityEmail       string                 `bun:"identity_email,notnull"`
	IdentityFirstName   string                 `bun:"identity_first_name,notnull"`
	IdentityLastName    string                 `bun:"identity_last_name,notnull"`
	IdentityState       IdentityState          `bun:"identity_state,notnull"`
	IdentityCreatedAt   time.Time              `bun:"identity_created_at,notnull"`
	IdentityUpdatedAt   time.Time              `bun:"identity_updated_at,notnull"`
	OrganizationID      string                 `bun:"organization_id,notnull"`
	OrganizationName    string                 `bun:"organization_name,notnull"`
	OrganizationSlug    string                 `bun:"organization_slug,notnull"`
	OrganizationLogo    string                 `bun:"organization_logo"`
	OrganizationKind    OrganizationKind       `bun:"organization_kind,notnull"`
	OrganizationState   OrganizationState      `bun:"organization_state,notnull"`
	MembershipRole      OrganizationMemberRole `bun:"membership_role,notnull"`
	MembershipCreatedAt time.Time              `bun:"membership_created_at,notnull"`
	MembershipUpdatedAt time.Time              `bun:"membership_updated_at,notnull"`
}

type CreateOrganizationRequest struct {
	Name string           `json:"name" validate:"required,min=3,max=255"`
	Kind OrganizationKind `json:"kind" validate:"required,oneof=standard management"`
}

type OrganizationResponse struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Slug      string            `json:"slug"`
	Kind      OrganizationKind  `json:"kind"`
	Logo      string            `json:"logo"`
	State     OrganizationState `json:"state"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

type OrganizationListResponse struct {
	Count int                    `json:"count"`
	Data  []OrganizationResponse `json:"data"`
}
