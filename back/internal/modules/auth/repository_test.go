package auth

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestPostgresUserRepositoryFindByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresUserRepository(db)
	now := time.Now()
	hash := "hash"
	rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "auth_provider", "email_verified", "profile_completed", "created_at", "updated_at"}).
		AddRow("user-1", "person@example.com", hash, AuthProviderEmail, true, false, now, now)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, email, password_hash, auth_provider, email_verified, profile_completed, created_at, updated_at
		FROM users WHERE email = $1
	`)).WithArgs("person@example.com").WillReturnRows(rows)

	user, err := repo.FindByEmail(context.Background(), "person@example.com")
	require.NoError(t, err)
	require.Equal(t, "user-1", user.ID)
	require.NotNil(t, user.PasswordHash)
	require.Equal(t, hash, *user.PasswordHash)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresUserRepositoryUpdatePassword(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresUserRepository(db)

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE users
		SET password_hash = $2, updated_at = NOW()
		WHERE id = $1
	`)).WithArgs("user-1", "hash").WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.UpdatePassword(context.Background(), "user-1", "hash")
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
