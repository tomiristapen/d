package onboarding

import (
	"context"
	"database/sql"
)

type Repository interface {
	GetStatus(ctx context.Context, userID string) (Status, error)
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) GetStatus(ctx context.Context, userID string) (Status, error) {
	var completed bool
	if err := r.db.QueryRowContext(ctx, `SELECT profile_completed FROM users WHERE id = $1`, userID).Scan(&completed); err != nil {
		return Status{}, err
	}
	return Status{ProfileCompleted: completed}, nil
}

