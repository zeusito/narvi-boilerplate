package iam

import (
	"context"
	"strings"
	"time"

	"github.com/uptrace/bun"
)

type defaultIdentityRepo struct {
	db *bun.DB
}

func newIdentityRepository(db *bun.DB) identityRepository {
	return &defaultIdentityRepo{db: db}
}

func (r *defaultIdentityRepo) Create(ctx context.Context, identity *Identity) error {
	_, err := r.db.NewInsert().
		Model(identity).
		Exec(ctx)

	if err != nil {
		return err
	}

	return nil
}

func (r *defaultIdentityRepo) FindActiveByEmail(ctx context.Context, email string) (*Identity, error) {
	var identity Identity
	err := r.db.NewSelect().
		Model(&identity).
		Where("lower(email) = ?", strings.ToLower(strings.TrimSpace(email))).
		Where("state = ?", IdentityStateActive).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return &identity, nil
}

func (r *defaultIdentityRepo) UpdateEmailVerifiedAt(ctx context.Context, id string, verifiedAt time.Time) error {
	_, err := r.db.NewUpdate().
		Model((*Identity)(nil)).
		Set("email_verified_at = ?", verifiedAt).
		Set("updated_at = ?", time.Now().UTC()).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return err
	}

	return nil
}
