package iam

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"backend-api/pkg/terrors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
)

type defaultIdentityRepo struct {
	db bun.IDB
}

func newIdentityRepository(db bun.IDB) identityRepository {
	return &defaultIdentityRepo{db: db}
}

func (r *defaultIdentityRepo) WithTx(tx bun.Tx) identityRepository {
	return &defaultIdentityRepo{db: tx}
}

func (r *defaultIdentityRepo) Create(ctx context.Context, identity *Identity) error {
	_, err := r.db.NewInsert().
		Model(identity).
		Exec(ctx)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return terrors.RecordAlreadyExists("identity with this email already exists")
		}

		log.Error().Err(err).Str("email", identity.Email).Msg("failed to insert identity into database")
		return terrors.OperationFailed("failed to create identity")
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, terrors.RecordNotFound("identity not found")
		}

		log.Error().Err(err).Str("email", email).Msg("failed to query identity by email")
		return nil, terrors.OperationFailed("failed to retrieve identity")
	}

	return &identity, nil
}

func (r *defaultIdentityRepo) UpdateEmailVerifiedAt(ctx context.Context, id string, verifiedAt time.Time) error {
	res, err := r.db.NewUpdate().
		Model((*Identity)(nil)).
		Set("email_verified_at = ?", verifiedAt).
		Set("updated_at = ?", time.Now().UTC()).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		log.Error().Err(err).Str("identity_id", id).Msg("failed to update email_verified_at")
		return terrors.OperationFailed("failed to update identity verification status")
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return terrors.RecordNotFound("identity not found")
	}

	return nil
}
