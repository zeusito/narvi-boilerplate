package iam

import (
	"context"
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
		return err
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
		return nil, err
	}

	return view, nil
}
