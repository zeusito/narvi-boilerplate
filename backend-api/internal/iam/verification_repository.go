package iam

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

type verificationRepo struct {
	db *bun.DB
}

func newVerificationRepository(db *bun.DB) verificationRepository {
	return &verificationRepo{db: db}
}

func (r *verificationRepo) Create(ctx context.Context, v *Verification) error {
	_, err := r.db.NewInsert().
		Model(v).
		Exec(ctx)

	if err != nil {
		return err
	}

	return nil
}

func (r *verificationRepo) FindLatestActive(ctx context.Context, identityID, kind string) (*Verification, error) {
	v := new(Verification)
	err := r.db.NewSelect().
		Model(v).
		Where("identity_id = ?", identityID).
		Where("kind = ?", kind).
		Where("expires_at > ?", time.Now().UTC()).
		Order("created_at DESC").
		Limit(1).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return v, nil
}

func (r *verificationRepo) IncrementAttempts(ctx context.Context, id string) error {
	_, err := r.db.NewUpdate().
		Model((*Verification)(nil)).
		Set("attempts = attempts + 1").
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return err
	}

	return nil
}

func (r *verificationRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().
		Model((*Verification)(nil)).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return err
	}

	return nil
}

func (r *verificationRepo) DeleteAllForIdentityAndKind(ctx context.Context, identityID, kind string) error {
	_, err := r.db.NewDelete().
		Model((*Verification)(nil)).
		Where("identity_id = ?", identityID).
		Where("kind = ?", kind).
		Exec(ctx)

	if err != nil {
		return err
	}

	return nil
}
