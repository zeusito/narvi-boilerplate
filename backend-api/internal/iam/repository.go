package iam

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

type organizationRepository interface {
	WithTx(tx bun.Tx) organizationRepository
	Create(ctx context.Context, org *Organization) error
	FindOneByID(ctx context.Context, id string) (*Organization, error)
	FindOneBySlug(ctx context.Context, slug string) (*Organization, error)
	FindAllMembershipsByOrganizationID(ctx context.Context, orgID string) ([]OrganizationMembershipView, error)
	FindAllMembershipsByIdentityID(ctx context.Context, identityID string) ([]OrganizationMembershipView, error)
	FindOldestMembershipsByIdentityID(ctx context.Context, identityID string) (*OrganizationMembershipView, error)
}

type identityRepository interface {
	WithTx(tx bun.Tx) identityRepository
	Create(ctx context.Context, identity *Identity) error
	FindActiveByEmail(ctx context.Context, email string) (*Identity, error)
	UpdateEmailVerifiedAt(ctx context.Context, id string, verifiedAt time.Time) error
}

type verificationRepository interface {
	WithTx(tx bun.Tx) verificationRepository
	Create(ctx context.Context, v *Verification) error
	FindLatestActive(ctx context.Context, identityID, kind string) (*Verification, error)
	IncrementAttempts(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
	DeleteAllForIdentityAndKind(ctx context.Context, identityID, kind string) error
}

type sessionRepository interface {
	WithTx(tx bun.Tx) sessionRepository
	Create(ctx context.Context, s *IdentitySession) error
	FindActiveSessionIntrospection(ctx context.Context, sessionHash string) (*SessionIntrospectionView, error)
	RemoveBySessionID(ctx context.Context, sessionID string) error
}
