package iam

import (
	"backend-api/pkg/terrors"
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/uptrace/bun"
)

type sessionRepo struct {
	db *bun.DB
}

func newSessionRepository(db *bun.DB) sessionRepository {
	return &sessionRepo{db: db}
}

func (r *sessionRepo) Create(ctx context.Context, s *IdentitySession) error {
	_, err := r.db.NewInsert().
		Model(s).
		Exec(ctx)

	if err != nil {
		return terrors.OperationFailed(err.Error())
	}

	return nil
}

func (r *sessionRepo) FindActiveSessionIntrospection(ctx context.Context, sessionHash string) (*SessionIntrospectionView, error) {
	view := new(SessionIntrospectionView)
	err := r.db.NewSelect().
		Model(view).
		Where("session_id = ?", sessionHash).
		Where("session_expires_at > ?", time.Now()).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, terrors.OperationFailed(err.Error())
	}

	return view, nil
}
