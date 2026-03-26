package auth

import (
	"context"
	"database/sql"
)

type TxRunner interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context, users UserRepository, codes VerificationCodeRepository) error) error
}

type RefreshTxRunner interface {
	WithinTxWithRefreshes(ctx context.Context, fn func(ctx context.Context, users UserRepository, codes VerificationCodeRepository, refreshes RefreshTokenRepository) error) error
}

type PostgresTxRunner struct {
	db *sql.DB
}

func NewPostgresTxRunner(db *sql.DB) *PostgresTxRunner {
	return &PostgresTxRunner{db: db}
}

func (r *PostgresTxRunner) WithinTx(ctx context.Context, fn func(ctx context.Context, users UserRepository, codes VerificationCodeRepository) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	users := newPostgresUserRepository(tx)
	codes := newPostgresVerificationRepository(tx)

	if err := fn(ctx, users, codes); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (r *PostgresTxRunner) WithinTxWithRefreshes(ctx context.Context, fn func(ctx context.Context, users UserRepository, codes VerificationCodeRepository, refreshes RefreshTokenRepository) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	users := newPostgresUserRepository(tx)
	codes := newPostgresVerificationRepository(tx)
	refreshes := newPostgresRefreshTokenRepository(tx)

	if err := fn(ctx, users, codes, refreshes); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
