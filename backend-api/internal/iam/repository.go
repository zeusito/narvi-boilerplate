package iam

import (
	"context"
	"time"
)

type organizationRepository interface {
	Create(ctx context.Context, org *Organization) error
	FindOneByID(ctx context.Context, id string) (*Organization, error)
	FindOneBySlug(ctx context.Context, slug string) (*Organization, error)
	FindMembershipsByOrganizationID(ctx context.Context, orgID string) ([]OrganizationMembershipView, error)
	FindMembershipsByIdentityID(ctx context.Context, identityID string) ([]OrganizationMembershipView, error)
}

type identityRepository interface {
	Create(ctx context.Context, identity *Identity) error
	FindActiveByEmail(ctx context.Context, email string) (*Identity, error)
	UpdateEmailVerifiedAt(ctx context.Context, id string, verifiedAt time.Time) error
}

type verificationRepository interface {
	Create(ctx context.Context, v *Verification) error
	FindLatestActive(ctx context.Context, identityID, kind string) (*Verification, error)
	IncrementAttempts(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
	DeleteAllForIdentityAndKind(ctx context.Context, identityID, kind string) error
}

type sessionRepository interface {
	Create(ctx context.Context, s *IdentitySession) error
	FindActiveSessionIntrospection(ctx context.Context, sessionHash string) (*SessionIntrospectionView, error)
}
