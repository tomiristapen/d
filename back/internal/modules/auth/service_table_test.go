package auth

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"back/internal/platform/jwtutil"

	"github.com/stretchr/testify/require"
)

type sequenceMailer struct {
	codes []string
}

func (m *sequenceMailer) SendVerificationCode(_ string, code string) error {
	m.codes = append(m.codes, code)
	return nil
}

func (m *sequenceMailer) SendPasswordResetCode(_ string, code string) error {
	m.codes = append(m.codes, code)
	return nil
}

func newTestService(users *memoryUserRepo, codes *memoryVerificationRepo, mailer any) *Service {
	sender, ok := mailer.(interface {
		SendVerificationCode(string, string) error
		SendPasswordResetCode(string, string) error
	})
	if !ok {
		panic("invalid mailer")
	}
	return NewService(
		users,
		codes,
		newMemoryRefreshRepo(),
		nil,
		jwtutil.NewManager("issuer", "access", "refresh", 15*time.Minute, 7*24*time.Hour),
		sender,
		stubGoogleVerifier{},
	)
}

func TestRegisterValidationTable(t *testing.T) {
	basePassword := "StrongPass1!"

	tests := []struct {
		name     string
		input    RegisterInput
		seedUser bool
		wantErr  string
		wantUser bool
	}{
		{
			name: "valid email and password",
			input: RegisterInput{
				Email:           "person@example.com",
				Password:        basePassword,
				ConfirmPassword: basePassword,
			},
			wantUser: true,
		},
		{
			name: "subdomain email",
			input: RegisterInput{
				Email:           "test@mail.co.uk",
				Password:        basePassword,
				ConfirmPassword: basePassword,
			},
			wantUser: true,
		},
		{
			name: "duplicate email",
			input: RegisterInput{
				Email:           "person@example.com",
				Password:        basePassword,
				ConfirmPassword: basePassword,
			},
			seedUser: true,
			wantErr:  "user already exists",
		},
		{
			name: "confirmation mismatch",
			input: RegisterInput{
				Email:           "person@example.com",
				Password:        basePassword,
				ConfirmPassword: "AnotherPass1!",
			},
			wantErr: "password confirmation does not match",
		},
		{
			name: "invalid email",
			input: RegisterInput{
				Email:           "abc",
				Password:        basePassword,
				ConfirmPassword: basePassword,
			},
			wantErr: "invalid email",
		},
		{
			name: "empty fields",
			input: RegisterInput{
				Email:           "",
				Password:        "",
				ConfirmPassword: "",
			},
			wantErr: "email is required",
		},
		{
			name: "email too long",
			input: RegisterInput{
				Email:           strings.Repeat("a", 245) + "@example.com",
				Password:        basePassword,
				ConfirmPassword: basePassword,
			},
			wantErr: "email is too long",
		},
		{
			name: "password min minus one fails",
			input: RegisterInput{
				Email:           "person@example.com",
				Password:        "Aa1!bcdef",
				ConfirmPassword: "Aa1!bcdef",
			},
			wantErr: fmt.Sprintf("at least %d", MinPasswordLength),
		},
		{
			name: "password min passes",
			input: RegisterInput{
				Email:           "person@example.com",
				Password:        "Aa1!bcdefg",
				ConfirmPassword: "Aa1!bcdefg",
			},
			wantUser: true,
		},
		{
			name: "password with special symbols passes",
			input: RegisterInput{
				Email:           "person@example.com",
				Password:        "Valid!@#$1A",
				ConfirmPassword: "Valid!@#$1A",
			},
			wantUser: true,
		},
		{
			name: "password max passes",
			input: RegisterInput{
				Email:           "person@example.com",
				Password:        makePassword(MaxPasswordLength),
				ConfirmPassword: makePassword(MaxPasswordLength),
			},
			wantUser: true,
		},
		{
			name: "password max plus one fails",
			input: RegisterInput{
				Email:           "person@example.com",
				Password:        makePassword(MaxPasswordLength + 1),
				ConfirmPassword: makePassword(MaxPasswordLength + 1),
			},
			wantErr: fmt.Sprintf("at most %d", MaxPasswordLength),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			users := newMemoryUserRepo()
			codes := newMemoryVerificationRepo()
			mailer := &captureMailer{}
			service := newTestService(users, codes, mailer)

			if tc.seedUser {
				hash, err := HashPassword(basePassword)
				require.NoError(t, err)
				require.NoError(t, users.Create(context.Background(), NewUser(NormalizeEmail(tc.input.Email), &hash, AuthProviderEmail, true)))
			}

			err := service.Register(context.Background(), tc.input)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
			_, err = users.FindByEmail(context.Background(), NormalizeEmail(tc.input.Email))
			require.NoError(t, err)
			require.NotEmpty(t, mailer.lastCode)
		})
	}
}

func TestVerifyEmailTable(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, users *memoryUserRepo, codes *memoryVerificationRepo) VerifyEmailInput
		wantErr   string
		wantValid bool
	}{
		{
			name: "correct code",
			setup: func(t *testing.T, users *memoryUserRepo, codes *memoryVerificationRepo) VerifyEmailInput {
				user := createEmailUser(t, users, "person@example.com", false)
				saveCode(t, codes, user.ID, "123456", time.Now().Add(10*time.Minute))
				return VerifyEmailInput{Email: user.Email, Code: "123456"}
			},
			wantValid: true,
		},
		{
			name: "wrong code",
			setup: func(t *testing.T, users *memoryUserRepo, codes *memoryVerificationRepo) VerifyEmailInput {
				user := createEmailUser(t, users, "person@example.com", false)
				saveCode(t, codes, user.ID, "123456", time.Now().Add(10*time.Minute))
				return VerifyEmailInput{Email: user.Email, Code: "000000"}
			},
			wantErr: "invalid email or code",
		},
		{
			name: "expired code",
			setup: func(t *testing.T, users *memoryUserRepo, codes *memoryVerificationRepo) VerifyEmailInput {
				user := createEmailUser(t, users, "person@example.com", false)
				saveCode(t, codes, user.ID, "123456", time.Now().Add(-time.Minute))
				return VerifyEmailInput{Email: user.Email, Code: "123456"}
			},
			wantErr: "expired",
		},
		{
			name: "unknown email",
			setup: func(t *testing.T, users *memoryUserRepo, codes *memoryVerificationRepo) VerifyEmailInput {
				return VerifyEmailInput{Email: "missing@example.com", Code: "123456"}
			},
			wantErr: "invalid email or code",
		},
		{
			name: "already verified email is idempotent",
			setup: func(t *testing.T, users *memoryUserRepo, codes *memoryVerificationRepo) VerifyEmailInput {
				user := createEmailUser(t, users, "person@example.com", true)
				saveCode(t, codes, user.ID, "123456", time.Now().Add(10*time.Minute))
				return VerifyEmailInput{Email: user.Email, Code: "123456"}
			},
			wantValid: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			users := newMemoryUserRepo()
			codes := newMemoryVerificationRepo()
			service := newTestService(users, codes, &noopMailer{})

			input := tc.setup(t, users, codes)
			err := service.VerifyEmail(context.Background(), input)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
			user, err := users.FindByEmail(context.Background(), NormalizeEmail(input.Email))
			require.NoError(t, err)
			require.True(t, user.EmailVerified)
		})
	}
}

func TestLoginTable(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, users *memoryUserRepo)
		input   LoginInput
		wantErr string
	}{
		{
			name: "valid credentials",
			setup: func(t *testing.T, users *memoryUserRepo) {
				createEmailUserWithPassword(t, users, "person@example.com", "StrongPass1!", true)
			},
			input: LoginInput{Email: "person@example.com", Password: "StrongPass1!"},
		},
		{
			name: "wrong password",
			setup: func(t *testing.T, users *memoryUserRepo) {
				createEmailUserWithPassword(t, users, "person@example.com", "StrongPass1!", true)
			},
			input:   LoginInput{Email: "person@example.com", Password: "WrongPass1!"},
			wantErr: "invalid credentials",
		},
		{
			name:    "unknown email",
			setup:   func(t *testing.T, users *memoryUserRepo) {},
			input:   LoginInput{Email: "missing@example.com", Password: "StrongPass1!"},
			wantErr: "invalid credentials",
		},
		{
			name: "unverified email",
			setup: func(t *testing.T, users *memoryUserRepo) {
				createEmailUserWithPassword(t, users, "person@example.com", "StrongPass1!", false)
			},
			input:   LoginInput{Email: "person@example.com", Password: "StrongPass1!"},
			wantErr: "not verified",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			users := newMemoryUserRepo()
			tc.setup(t, users)
			service := newTestService(users, newMemoryVerificationRepo(), &noopMailer{})

			_, err := service.Login(context.Background(), tc.input)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestLoginWithCodeEdges(t *testing.T) {
	t.Run("expired code", func(t *testing.T) {
		users := newMemoryUserRepo()
		codes := newMemoryVerificationRepo()
		service := newTestService(users, codes, &noopMailer{})
		user := createEmailUser(t, users, "person@example.com", true)
		saveCode(t, codes, user.ID, "123456", time.Now().Add(-time.Minute))

		_, err := service.LoginWithCode(context.Background(), EmailCodeLoginInput{
			Email: user.Email,
			Code:  "123456",
		})
		require.ErrorContains(t, err, "expired")
	})

	t.Run("repeated use fails", func(t *testing.T) {
		users := newMemoryUserRepo()
		codes := newMemoryVerificationRepo()
		service := newTestService(users, codes, &noopMailer{})
		user := createEmailUser(t, users, "person@example.com", true)
		saveCode(t, codes, user.ID, "123456", time.Now().Add(10*time.Minute))

		_, err := service.LoginWithCode(context.Background(), EmailCodeLoginInput{
			Email: user.Email,
			Code:  "123456",
		})
		require.NoError(t, err)

		_, err = service.LoginWithCode(context.Background(), EmailCodeLoginInput{
			Email: user.Email,
			Code:  "123456",
		})
		require.ErrorContains(t, err, "invalid email or code")
	})

	t.Run("latest code wins", func(t *testing.T) {
		users := newMemoryUserRepo()
		codes := newMemoryVerificationRepo()
		mailer := &sequenceMailer{}
		service := newTestService(users, codes, mailer)
		user := createEmailUser(t, users, "person@example.com", true)

		require.NoError(t, service.SendLoginCode(context.Background(), user.Email))
		require.NoError(t, service.SendLoginCode(context.Background(), user.Email))
		require.Len(t, mailer.codes, 2)

		_, err := service.LoginWithCode(context.Background(), EmailCodeLoginInput{
			Email: user.Email,
			Code:  mailer.codes[0],
		})
		require.ErrorContains(t, err, "invalid email or code")

		resp, err := service.LoginWithCode(context.Background(), EmailCodeLoginInput{
			Email: user.Email,
			Code:  mailer.codes[1],
		})
		require.NoError(t, err)
		require.NotEmpty(t, resp.AccessToken)
	})
}

func TestRefreshTable(t *testing.T) {
	t.Run("valid refresh token", func(t *testing.T) {
		users := newMemoryUserRepo()
		createEmailUserWithPassword(t, users, "person@example.com", "StrongPass1!", true)
		service := newTestService(users, newMemoryVerificationRepo(), &noopMailer{})

		pair, err := service.Login(context.Background(), LoginInput{
			Email:    "person@example.com",
			Password: "StrongPass1!",
		})
		require.NoError(t, err)

		resp, err := service.Refresh(context.Background(), pair.RefreshToken)
		require.NoError(t, err)
		require.NotEmpty(t, resp.AccessToken)
		require.NotEqual(t, pair.RefreshToken, resp.RefreshToken)
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		service := newTestService(newMemoryUserRepo(), newMemoryVerificationRepo(), &noopMailer{})
		_, err := service.Refresh(context.Background(), "not-a-token")
		require.ErrorContains(t, err, "invalid refresh token")
	})

	t.Run("expired refresh token", func(t *testing.T) {
		users := newMemoryUserRepo()
		user := createEmailUserWithPassword(t, users, "person@example.com", "StrongPass1!", true)
		service := NewService(
			users,
			newMemoryVerificationRepo(),
			newMemoryRefreshRepo(),
			nil,
			jwtutil.NewManager("issuer", "access", "refresh", 15*time.Minute, -1*time.Minute),
			&noopMailer{},
			stubGoogleVerifier{},
		)

		pair, err := service.tokens.GeneratePair(user.ID)
		require.NoError(t, err)

		_, err = service.Refresh(context.Background(), pair.RefreshToken)
		require.ErrorContains(t, err, "invalid refresh token")
	})

	t.Run("same refresh token cannot be reused", func(t *testing.T) {
		users := newMemoryUserRepo()
		createEmailUserWithPassword(t, users, "person@example.com", "StrongPass1!", true)
		service := newTestService(users, newMemoryVerificationRepo(), &noopMailer{})

		pair, err := service.Login(context.Background(), LoginInput{
			Email:    "person@example.com",
			Password: "StrongPass1!",
		})
		require.NoError(t, err)

		next, err := service.Refresh(context.Background(), pair.RefreshToken)
		require.NoError(t, err)
		_, err = service.Refresh(context.Background(), pair.RefreshToken)
		require.ErrorContains(t, err, "invalid refresh token")

		_, err = service.Refresh(context.Background(), next.RefreshToken)
		require.NoError(t, err)
	})
}

func TestGoogleLoginInvalidToken(t *testing.T) {
	service := NewService(
		newMemoryUserRepo(),
		newMemoryVerificationRepo(),
		newMemoryRefreshRepo(),
		nil,
		jwtutil.NewManager("issuer", "access", "refresh", 15*time.Minute, 7*24*time.Hour),
		&noopMailer{},
		stubGoogleVerifier{err: fmt.Errorf("invalid token")},
	)

	_, err := service.GoogleLogin(context.Background(), GoogleLoginInput{IDToken: "bad-token"})
	require.ErrorContains(t, err, "invalid token")
}

func TestSetPasswordOverwritesExistingPassword(t *testing.T) {
	users := newMemoryUserRepo()
	user := createEmailUserWithPassword(t, users, "person@example.com", "OldStrongPass1!", true)
	service := newTestService(users, newMemoryVerificationRepo(), &noopMailer{})

	err := service.SetPassword(context.Background(), user.ID, SetPasswordInput{
		Password:        "NewStrongPass1!",
		ConfirmPassword: "NewStrongPass1!",
	})
	require.NoError(t, err)

	_, err = service.Login(context.Background(), LoginInput{
		Email:    user.Email,
		Password: "NewStrongPass1!",
	})
	require.NoError(t, err)

	_, err = service.Login(context.Background(), LoginInput{
		Email:    user.Email,
		Password: "OldStrongPass1!",
	})
	require.ErrorContains(t, err, "invalid credentials")
}

func createEmailUser(t *testing.T, users *memoryUserRepo, email string, verified bool) User {
	t.Helper()
	return createEmailUserWithPassword(t, users, email, "StrongPass1!", verified)
}

func createEmailUserWithPassword(t *testing.T, users *memoryUserRepo, email, password string, verified bool) User {
	t.Helper()
	hash, err := HashPassword(password)
	require.NoError(t, err)
	user := NewUser(NormalizeEmail(email), &hash, AuthProviderEmail, verified)
	require.NoError(t, users.Create(context.Background(), user))
	return user
}

func saveCode(t *testing.T, codes *memoryVerificationRepo, userID, code string, expiresAt time.Time) {
	t.Helper()
	hash, salt, err := hashVerificationCode(code)
	require.NoError(t, err)
	require.NoError(t, codes.Save(context.Background(), VerificationCode{
		UserID:    userID,
		CodeHash:  hash,
		CodeSalt:  salt,
		ExpiresAt: expiresAt,
	}))
}

func makePassword(length int) string {
	if length <= 4 {
		return "Aa1!"
	}
	return "Aa1!" + strings.Repeat("x", length-4)
}
