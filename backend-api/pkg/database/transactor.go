package database

import (
	"context"
	"database/sql"

	"github.com/uptrace/bun"
)

type Transactor interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context, tx bun.Tx) error) error
}

func NewBunTransactor(db *bun.DB) Transactor {
	return &BunTransactor{db: db}
}

type BunTransactor struct {
	db *bun.DB
}

func (t *BunTransactor) WithinTransaction(ctx context.Context, fn func(ctx context.Context, tx bun.Tx) error) error {
	return t.db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		return fn(ctx, tx)
	})
}
