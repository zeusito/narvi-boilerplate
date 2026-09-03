package iam

import (
	"backend-api/pkg/terrors"
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/uptrace/bun"
)

type identityRepo struct {
	db *bun.DB
}

func newIdentityRepository(db *bun.DB) identityRepository {
	return &identityRepo{db: db}
}

func (r *identityRepo) FindActiveByEmail(ctx context.Context, email string) (*Identity, error) {
	identity := new(Identity)
	err := r.db.NewSelect().
		Model(identity).
		Where("lower(email) = ?", strings.ToLower(strings.TrimSpace(email))).
		Where("state = ?", "active").
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, terrors.OperationFailed(err.Error())
	}

	return identity, nil
}

func (r *identityRepo) UpdateEmailVerifiedAt(ctx context.Context, id string, verifiedAt time.Time) error {
	_, err := r.db.NewUpdate().
		Model((*Identity)(nil)).
		Set("email_verified_at = ?", verifiedAt).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return terrors.OperationFailed(err.Error())
	}

	return nil
}

func (r *identityRepo) FindMembershipsByIdentityID(ctx context.Context, identityID string) ([]IdentityMembership, error) {
	var memberships []IdentityMembership
	err := r.db.NewSelect().
		Table("organization_memberships").
		Column("organization_id", "role", "created_at").
		Where("identity_id = ?", identityID).
		Order("created_at ASC").
		Scan(ctx, &memberships)

	if err != nil {
		return nil, terrors.OperationFailed(err.Error())
	}

	return memberships, nil
}

func (r *identityRepo) FindOrganizationByID(ctx context.Context, orgID string) (*ActiveOrganizationDTO, error) {
	type orgRecord struct {
		ID   string `bun:"id"`
		Name string `bun:"name"`
		Slug string `bun:"slug"`
		Logo string `bun:"logo"`
	}

	record := new(orgRecord)
	err := r.db.NewSelect().
		Table("organizations").
		Column("id", "name", "slug", "logo").
		Where("id = ?", orgID).
		Scan(ctx, record)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, terrors.OperationFailed(err.Error())
	}

	return &ActiveOrganizationDTO{
		ID:   record.ID,
		Name: record.Name,
		Slug: record.Slug,
		Logo: record.Logo,
	}, nil
}

func (r *identityRepo) FindPendingInvitationsByEmail(ctx context.Context, email string) ([]PendingInvitationDTO, error) {
	var invitations []Invitation
	err := r.db.NewSelect().
		Model(&invitations).
		Where("lower(email) = ?", strings.ToLower(strings.TrimSpace(email))).
		Where("state = ?", "pending").
		Where("expires_at > ?", time.Now()).
		Order("created_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, terrors.OperationFailed(err.Error())
	}

	dtos := make([]PendingInvitationDTO, len(invitations))
	for i, inv := range invitations {
		dtos[i] = PendingInvitationDTO{
			ID:        inv.ID,
			Kind:      inv.Kind,
			TargetID:  inv.TargetID,
			Role:      inv.Role,
			State:     inv.State,
			ExpiresAt: inv.ExpiresAt,
		}
	}

	return dtos, nil
}
