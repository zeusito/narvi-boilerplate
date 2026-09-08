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

type sessionRepo struct {
	db bun.IDB
}

func newSessionRepository(db bun.IDB) sessionRepository {
	return &sessionRepo{db: db}
}

func (r *sessionRepo) WithTx(tx bun.Tx) sessionRepository {
	return &sessionRepo{db: tx}
}

func (r *sessionRepo) Create(ctx context.Context, s *IdentitySession) error {
	_, err := r.db.NewInsert().
		Model(s).
		Exec(ctx)

	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to insert session into database")
		return terrors.OperationFailed("failed to create session")
	}

	return nil
}

func (r *sessionRepo) FindActiveSessionIntrospection(ctx context.Context, sessionHash string) (*SessionIntrospectionView, error) {
	view := new(SessionIntrospectionView)
	err := r.db.NewSelect().
		Model(view).
		Where("session_id = ?", sessionHash).
		Where("session_expires_at > ?", time.Now().UTC()).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, terrors.RecordNotFound("active session not found")
		}

		return nil, terrors.OperationFailed("failed to retrieve session")
	}

	return view, nil
}

func (r *sessionRepo) RemoveBySessionID(ctx context.Context, sessionID string) error {
	_, err := r.db.NewDelete().
		Model((*IdentitySession)(nil)).
		Where("id = ?", sessionID).
		Exec(ctx)

	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to remove session")
		return terrors.OperationFailed("failed to remove session")
	}

	return nil
}
