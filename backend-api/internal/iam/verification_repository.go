package iam

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"backend-api/pkg/terrors"

	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
)

type verificationRepo struct {
	db bun.IDB
}

func newVerificationRepository(db bun.IDB) verificationRepository {
	return &verificationRepo{db: db}
}

func (r *verificationRepo) WithTx(tx bun.Tx) verificationRepository {
	return &verificationRepo{db: tx}
}

func (r *verificationRepo) Create(ctx context.Context, v *Verification) error {
	_, err := r.db.NewInsert().
		Model(v).
		Exec(ctx)

	if err != nil {
		log.Error().Err(err).Str("identity_id", v.IdentityID).Msg("failed to insert verification record")
		return terrors.OperationFailed("failed to create verification")
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, terrors.RecordNotFound("active verification not found")
		}

		log.Error().Err(err).Str("identity_id", identityID).Str("kind", kind).Msg("failed to query active verification")
		return nil, terrors.OperationFailed("failed to retrieve verification")
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
		log.Error().Err(err).Str("verification_id", id).Msg("failed to increment verification attempts")
		return terrors.OperationFailed("failed to increment verification attempts")
	}

	return nil
}

func (r *verificationRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().
		Model((*Verification)(nil)).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		log.Error().Err(err).Str("verification_id", id).Msg("failed to delete verification record")
		return terrors.OperationFailed("failed to delete verification")
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
		log.Error().Err(err).Str("identity_id", identityID).Str("kind", kind).Msg("failed to delete prior verifications")
		return terrors.OperationFailed("failed to clean prior verifications")
	}

	return nil
}
