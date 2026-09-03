package iam

import (
	"context"
	"time"
)

type identityRepository interface {
	FindActiveByEmail(ctx context.Context, email string) (*Identity, error)
	UpdateEmailVerifiedAt(ctx context.Context, id string, verifiedAt time.Time) error
	FindMembershipsByIdentityID(ctx context.Context, identityID string) ([]IdentityMembership, error)
	FindOrganizationByID(ctx context.Context, orgID string) (*ActiveOrganizationDTO, error)
	FindPendingInvitationsByEmail(ctx context.Context, email string) ([]PendingInvitationDTO, error)
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
