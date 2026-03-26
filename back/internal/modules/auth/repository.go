package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("not found")

type UserRepository interface {
	Create(ctx context.Context, user User) error
	FindByEmail(ctx context.Context, email string) (User, error)
	FindByID(ctx context.Context, id string) (User, error)
	MarkEmailVerified(ctx context.Context, userID string) error
	UpdatePassword(ctx context.Context, userID string, passwordHash string) error
	UpdateProfileCompleted(ctx context.Context, userID string, completed bool) error
}

type VerificationCodeRepository interface {
	Save(ctx context.Context, code VerificationCode) error
	FindByUserID(ctx context.Context, userID string) (VerificationCode, error)
	IncrementAttempts(ctx context.Context, userID string) (int, error)
	DeleteByUserID(ctx context.Context, userID string) error
}

type RefreshTokenRepository interface {
	Save(ctx context.Context, session RefreshSession) error
	Consume(ctx context.Context, tokenID string, userID string, now time.Time) error
}

type sqlQuerier interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type PostgresUserRepository struct {
	q sqlQuerier
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{q: db}
}

func newPostgresUserRepository(q sqlQuerier) *PostgresUserRepository {
	return &PostgresUserRepository{q: q}
}

func (r *PostgresUserRepository) Create(ctx context.Context, user User) error {
	_, err := r.q.ExecContext(ctx, `
		INSERT INTO users(id, email, password_hash, auth_provider, email_verified, profile_completed, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`, user.ID, user.Email, user.PasswordHash, user.AuthProvider, user.EmailVerified, user.ProfileCompleted, user.CreatedAt, user.UpdatedAt)
	return err
}

func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (User, error) {
	row := r.q.QueryRowContext(ctx, `
		SELECT id, email, password_hash, auth_provider, email_verified, profile_completed, created_at, updated_at
		FROM users WHERE email = $1
	`, email)
	return scanUser(row)
}

func (r *PostgresUserRepository) FindByID(ctx context.Context, id string) (User, error) {
	row := r.q.QueryRowContext(ctx, `
		SELECT id, email, password_hash, auth_provider, email_verified, profile_completed, created_at, updated_at
		FROM users WHERE id = $1
	`, id)
	return scanUser(row)
}

func (r *PostgresUserRepository) MarkEmailVerified(ctx context.Context, userID string) error {
	_, err := r.q.ExecContext(ctx, `UPDATE users SET email_verified = TRUE, updated_at = NOW() WHERE id = $1`, userID)
	return err
}

func (r *PostgresUserRepository) UpdatePassword(ctx context.Context, userID string, passwordHash string) error {
	result, err := r.q.ExecContext(ctx, `
		UPDATE users
		SET password_hash = $2, updated_at = NOW()
		WHERE id = $1
	`, userID, passwordHash)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresUserRepository) UpdateProfileCompleted(ctx context.Context, userID string, completed bool) error {
	_, err := r.q.ExecContext(ctx, `UPDATE users SET profile_completed = $2, updated_at = NOW() WHERE id = $1`, userID, completed)
	return err
}

type scanner interface {
	Scan(dest ...interface{}) error
}

func scanUser(row scanner) (User, error) {
	var user User
	var passwordHash sql.NullString
	err := row.Scan(&user.ID, &user.Email, &passwordHash, &user.AuthProvider, &user.EmailVerified, &user.ProfileCompleted, &user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, err
	}
	if passwordHash.Valid {
		user.PasswordHash = &passwordHash.String
	}
	return user, nil
}

type PostgresVerificationRepository struct {
	q sqlQuerier
}

func NewPostgresVerificationRepository(db *sql.DB) *PostgresVerificationRepository {
	return &PostgresVerificationRepository{q: db}
}

func newPostgresVerificationRepository(q sqlQuerier) *PostgresVerificationRepository {
	return &PostgresVerificationRepository{q: q}
}

func (r *PostgresVerificationRepository) Save(ctx context.Context, code VerificationCode) error {
	_, err := r.q.ExecContext(ctx, `
		INSERT INTO email_verification_codes(user_id, code_hash, code_salt, expires_at, attempts, created_at)
		VALUES ($1, $2, $3, $4, 0, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			code_hash = EXCLUDED.code_hash,
			code_salt = EXCLUDED.code_salt,
			expires_at = EXCLUDED.expires_at,
			attempts = 0,
			created_at = NOW()
	`, code.UserID, code.CodeHash, code.CodeSalt, code.ExpiresAt)
	return err
}

func (r *PostgresVerificationRepository) FindByUserID(ctx context.Context, userID string) (VerificationCode, error) {
	var result VerificationCode
	err := r.q.QueryRowContext(ctx, `
		SELECT user_id, code_hash, code_salt, expires_at, attempts
		FROM email_verification_codes WHERE user_id = $1
	`, userID).Scan(&result.UserID, &result.CodeHash, &result.CodeSalt, &result.ExpiresAt, &result.Attempts)
	if errors.Is(err, sql.ErrNoRows) {
		return VerificationCode{}, ErrNotFound
	}
	return result, err
}

func (r *PostgresVerificationRepository) IncrementAttempts(ctx context.Context, userID string) (int, error) {
	var attempts int
	err := r.q.QueryRowContext(ctx, `
		UPDATE email_verification_codes
		SET attempts = attempts + 1
		WHERE user_id = $1
		RETURNING attempts
	`, userID).Scan(&attempts)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	return attempts, err
}

func (r *PostgresVerificationRepository) DeleteByUserID(ctx context.Context, userID string) error {
	_, err := r.q.ExecContext(ctx, `DELETE FROM email_verification_codes WHERE user_id = $1`, userID)
	return err
}

type PostgresRefreshTokenRepository struct {
	q sqlQuerier
}

func NewPostgresRefreshTokenRepository(db *sql.DB) *PostgresRefreshTokenRepository {
	return &PostgresRefreshTokenRepository{q: db}
}

func newPostgresRefreshTokenRepository(q sqlQuerier) *PostgresRefreshTokenRepository {
	return &PostgresRefreshTokenRepository{q: q}
}

func (r *PostgresRefreshTokenRepository) Save(ctx context.Context, session RefreshSession) error {
	_, err := r.q.ExecContext(ctx, `
		INSERT INTO refresh_sessions(token_id, user_id, expires_at, consumed_at, created_at)
		VALUES ($1, $2, $3, $4, NOW())
	`, session.TokenID, session.UserID, session.ExpiresAt, session.ConsumedAt)
	return err
}

func (r *PostgresRefreshTokenRepository) Consume(ctx context.Context, tokenID string, userID string, now time.Time) error {
	result, err := r.q.ExecContext(ctx, `
		UPDATE refresh_sessions
		SET consumed_at = $3
		WHERE token_id = $1
		  AND user_id = $2
		  AND consumed_at IS NULL
		  AND expires_at > $3
	`, tokenID, userID, now)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func NewUser(email string, passwordHash *string, provider AuthProvider, emailVerified bool) User {
	now := time.Now().UTC()
	return User{
		ID:               uuid.NewString(),
		Email:            email,
		PasswordHash:     passwordHash,
		AuthProvider:     provider,
		EmailVerified:    emailVerified,
		ProfileCompleted: false,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}
