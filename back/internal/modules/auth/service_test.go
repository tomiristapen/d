package auth

import (
	"context"
	"fmt"
	"maps"
	"testing"
	"time"

	"back/internal/platform/jwtutil"

	"github.com/stretchr/testify/require"
)

type memoryUserRepo struct {
	byEmail map[string]User
	byID    map[string]User
}

func newMemoryUserRepo() *memoryUserRepo {
	return &memoryUserRepo{
		byEmail: map[string]User{},
		byID:    map[string]User{},
	}
}

func (r *memoryUserRepo) Create(_ context.Context, user User) error {
	r.byEmail[user.Email] = user
	r.byID[user.ID] = user
	return nil
}

func (r *memoryUserRepo) FindByEmail(_ context.Context, email string) (User, error) {
	user, ok := r.byEmail[email]
	if !ok {
		return User{}, ErrNotFound
	}
	return user, nil
}

func (r *memoryUserRepo) FindByID(_ context.Context, id string) (User, error) {
	user, ok := r.byID[id]
	if !ok {
		return User{}, ErrNotFound
	}
	return user, nil
}

func (r *memoryUserRepo) MarkEmailVerified(_ context.Context, userID string) error {
	user := r.byID[userID]
	user.EmailVerified = true
	r.byID[userID] = user
	r.byEmail[user.Email] = user
	return nil
}

func (r *memoryUserRepo) UpdatePassword(_ context.Context, userID string, passwordHash string) error {
	user, ok := r.byID[userID]
	if !ok {
		return ErrNotFound
	}
	user.PasswordHash = &passwordHash
	r.byID[userID] = user
	r.byEmail[user.Email] = user
	return nil
}

func (r *memoryUserRepo) UpdateProfileCompleted(_ context.Context, userID string, completed bool) error {
	user := r.byID[userID]
	user.ProfileCompleted = completed
	r.byID[userID] = user
	r.byEmail[user.Email] = user
	return nil
}

type memoryVerificationRepo struct {
	byUserID map[string]VerificationCode
}

func newMemoryVerificationRepo() *memoryVerificationRepo {
	return &memoryVerificationRepo{byUserID: map[string]VerificationCode{}}
}

func (r *memoryVerificationRepo) Save(_ context.Context, code VerificationCode) error {
	r.byUserID[code.UserID] = code
	return nil
}

func (r *memoryVerificationRepo) FindByUserID(_ context.Context, userID string) (VerificationCode, error) {
	value, ok := r.byUserID[userID]
	if !ok {
		return VerificationCode{}, ErrNotFound
	}
	return value, nil
}

func (r *memoryVerificationRepo) IncrementAttempts(_ context.Context, userID string) (int, error) {
	value, ok := r.byUserID[userID]
	if !ok {
		return 0, ErrNotFound
	}
	value.Attempts++
	r.byUserID[userID] = value
	return value.Attempts, nil
}

func (r *memoryVerificationRepo) DeleteByUserID(_ context.Context, userID string) error {
	delete(r.byUserID, userID)
	return nil
}

type memoryRefreshRepo struct {
	byTokenID map[string]RefreshSession
}

func newMemoryRefreshRepo() *memoryRefreshRepo {
	return &memoryRefreshRepo{byTokenID: map[string]RefreshSession{}}
}

func (r *memoryRefreshRepo) Save(_ context.Context, session RefreshSession) error {
	r.byTokenID[session.TokenID] = session
	return nil
}

func (r *memoryRefreshRepo) Consume(_ context.Context, tokenID string, userID string, now time.Time) error {
	session, ok := r.byTokenID[tokenID]
	if !ok || session.UserID != userID || session.ConsumedAt != nil || !session.ExpiresAt.After(now) {
		return ErrNotFound
	}
	session.ConsumedAt = &now
	r.byTokenID[tokenID] = session
	return nil
}

type noopMailer struct{}

func (noopMailer) SendVerificationCode(string, string) error { return nil }
func (noopMailer) SendPasswordResetCode(string, string) error {
	return nil
}

type errorMailer struct {
	err error
}

func (m errorMailer) SendVerificationCode(string, string) error { return m.err }
func (m errorMailer) SendPasswordResetCode(string, string) error {
	return m.err
}

type captureMailer struct {
	lastCode string
}

func (m *captureMailer) SendVerificationCode(_ string, code string) error {
	m.lastCode = code
	return nil
}

func (m *captureMailer) SendPasswordResetCode(_ string, code string) error {
	m.lastCode = code
	return nil
}

type stubGoogleVerifier struct {
	email string
	err   error
}

func (s stubGoogleVerifier) VerifyIDToken(string) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	return s.email, nil
}

type memoryTxRunner struct {
	users         *memoryUserRepo
	verifications *memoryVerificationRepo
	refreshes     *memoryRefreshRepo
}

func (r memoryTxRunner) WithinTx(ctx context.Context, fn func(ctx context.Context, users UserRepository, codes VerificationCodeRepository) error) error {
	usersCopy := &memoryUserRepo{
		byEmail: maps.Clone(r.users.byEmail),
		byID:    maps.Clone(r.users.byID),
	}
	verificationsCopy := &memoryVerificationRepo{
		byUserID: maps.Clone(r.verifications.byUserID),
	}

	if err := fn(ctx, usersCopy, verificationsCopy); err != nil {
		return err
	}

	r.users.byEmail = usersCopy.byEmail
	r.users.byID = usersCopy.byID
	r.verifications.byUserID = verificationsCopy.byUserID
	return nil
}

func (r memoryTxRunner) WithinTxWithRefreshes(ctx context.Context, fn func(ctx context.Context, users UserRepository, codes VerificationCodeRepository, refreshes RefreshTokenRepository) error) error {
	usersCopy := &memoryUserRepo{
		byEmail: maps.Clone(r.users.byEmail),
		byID:    maps.Clone(r.users.byID),
	}
	verificationsCopy := &memoryVerificationRepo{
		byUserID: maps.Clone(r.verifications.byUserID),
	}
	refreshCopy := &memoryRefreshRepo{
		byTokenID: maps.Clone(r.refreshes.byTokenID),
	}

	if err := fn(ctx, usersCopy, verificationsCopy, refreshCopy); err != nil {
		return err
	}

	r.users.byEmail = usersCopy.byEmail
	r.users.byID = usersCopy.byID
	r.verifications.byUserID = verificationsCopy.byUserID
	r.refreshes.byTokenID = refreshCopy.byTokenID
	return nil
}

func TestRegisterAndVerifyEmail(t *testing.T) {
	users := newMemoryUserRepo()
	verifications := newMemoryVerificationRepo()
	mailer := &captureMailer{}
	service := NewService(
		users,
		verifications,
		newMemoryRefreshRepo(),
		nil,
		jwtutil.NewManager("issuer", "access", "refresh", 15*time.Minute, 7*24*time.Hour),
		mailer,
		stubGoogleVerifier{},
	)

	err := service.Register(context.Background(), RegisterInput{
		Email:           "person@example.com",
		Password:        "StrongPass1!",
		ConfirmPassword: "StrongPass1!",
	})
	require.NoError(t, err)
	require.NotEmpty(t, mailer.lastCode)

	err = service.VerifyEmail(context.Background(), VerifyEmailInput{
		Email: "person@example.com",
		Code:  mailer.lastCode,
	})
	require.NoError(t, err)

	user, err := users.FindByEmail(context.Background(), "person@example.com")
	require.NoError(t, err)
	require.True(t, user.EmailVerified)
}

func TestLoginRequiresVerifiedEmail(t *testing.T) {
	users := newMemoryUserRepo()
	hash, err := HashPassword("StrongPass1!")
	require.NoError(t, err)
	user := NewUser("person@example.com", &hash, AuthProviderEmail, false)
	require.NoError(t, users.Create(context.Background(), user))

	service := NewService(
		users,
		newMemoryVerificationRepo(),
		newMemoryRefreshRepo(),
		nil,
		jwtutil.NewManager("issuer", "access", "refresh", 15*time.Minute, 7*24*time.Hour),
		noopMailer{},
		stubGoogleVerifier{},
	)

	_, err = service.Login(context.Background(), LoginInput{Email: user.Email, Password: "StrongPass1!"})
	require.ErrorContains(t, err, "not verified")
}

func TestGoogleLoginCreatesUser(t *testing.T) {
	users := newMemoryUserRepo()
	service := NewService(
		users,
		newMemoryVerificationRepo(),
		newMemoryRefreshRepo(),
		nil,
		jwtutil.NewManager("issuer", "access", "refresh", 15*time.Minute, 7*24*time.Hour),
		noopMailer{},
		stubGoogleVerifier{email: "google@example.com"},
	)

	resp, err := service.GoogleLogin(context.Background(), GoogleLoginInput{IDToken: "token"})
	require.NoError(t, err)
	require.NotEmpty(t, resp.AccessToken)

	user, err := users.FindByEmail(context.Background(), "google@example.com")
	require.NoError(t, err)
	require.Equal(t, AuthProviderGoogle, user.AuthProvider)
	require.True(t, user.EmailVerified)
}

func TestSetPasswordAllowsGoogleUserToLoginWithEmailPassword(t *testing.T) {
	users := newMemoryUserRepo()
	service := NewService(
		users,
		newMemoryVerificationRepo(),
		newMemoryRefreshRepo(),
		nil,
		jwtutil.NewManager("issuer", "access", "refresh", 15*time.Minute, 7*24*time.Hour),
		noopMailer{},
		stubGoogleVerifier{email: "google@example.com"},
	)

	_, err := service.GoogleLogin(context.Background(), GoogleLoginInput{IDToken: "token"})
	require.NoError(t, err)

	user, err := users.FindByEmail(context.Background(), "google@example.com")
	require.NoError(t, err)
	require.Nil(t, user.PasswordHash)

	err = service.SetPassword(context.Background(), user.ID, SetPasswordInput{
		Password:        "StrongPass1!",
		ConfirmPassword: "StrongPass1!",
	})
	require.NoError(t, err)

	updated, err := users.FindByEmail(context.Background(), "google@example.com")
	require.NoError(t, err)
	require.NotNil(t, updated.PasswordHash)
	require.Equal(t, AuthProviderGoogle, updated.AuthProvider)

	_, err = service.Login(context.Background(), LoginInput{
		Email:    "google@example.com",
		Password: "StrongPass1!",
	})
	require.NoError(t, err)
}

func TestSendLoginCodeAndLoginWithCodeAllowGoogleFallback(t *testing.T) {
	users := newMemoryUserRepo()
	verifications := newMemoryVerificationRepo()
	mailer := &captureMailer{}
	service := NewService(
		users,
		verifications,
		newMemoryRefreshRepo(),
		nil,
		jwtutil.NewManager("issuer", "access", "refresh", 15*time.Minute, 7*24*time.Hour),
		mailer,
		stubGoogleVerifier{email: "google@example.com"},
	)

	resp, err := service.GoogleLogin(context.Background(), GoogleLoginInput{IDToken: "token"})
	require.NoError(t, err)
	require.False(t, resp.HasPassword)

	err = service.SendLoginCode(context.Background(), "google@example.com")
	require.NoError(t, err)
	require.NotEmpty(t, mailer.lastCode)

	resp, err = service.LoginWithCode(context.Background(), EmailCodeLoginInput{
		Email: "google@example.com",
		Code:  mailer.lastCode,
	})
	require.NoError(t, err)
	require.NotEmpty(t, resp.AccessToken)
	require.False(t, resp.HasPassword)
}

func TestLoginWithCodeVerifiesUnverifiedEmailUser(t *testing.T) {
	users := newMemoryUserRepo()
	verifications := newMemoryVerificationRepo()
	mailer := &captureMailer{}
	service := NewService(
		users,
		verifications,
		newMemoryRefreshRepo(),
		nil,
		jwtutil.NewManager("issuer", "access", "refresh", 15*time.Minute, 7*24*time.Hour),
		mailer,
		stubGoogleVerifier{},
	)

	err := service.Register(context.Background(), RegisterInput{
		Email:           "person@example.com",
		Password:        "StrongPass1!",
		ConfirmPassword: "StrongPass1!",
	})
	require.NoError(t, err)
	require.NotEmpty(t, mailer.lastCode)

	resp, err := service.LoginWithCode(context.Background(), EmailCodeLoginInput{
		Email: "person@example.com",
		Code:  mailer.lastCode,
	})
	require.NoError(t, err)
	require.NotEmpty(t, resp.AccessToken)
	require.True(t, resp.HasPassword)

	user, err := users.FindByEmail(context.Background(), "person@example.com")
	require.NoError(t, err)
	require.True(t, user.EmailVerified)
}

func TestSendPasswordResetCodeAndResetPasswordAllowRecovery(t *testing.T) {
	users := newMemoryUserRepo()
	verifications := newMemoryVerificationRepo()
	mailer := &captureMailer{}
	service := NewService(
		users,
		verifications,
		newMemoryRefreshRepo(),
		nil,
		jwtutil.NewManager("issuer", "access", "refresh", 15*time.Minute, 7*24*time.Hour),
		mailer,
		stubGoogleVerifier{},
	)

	hash, err := HashPassword("OldStrongPass1!")
	require.NoError(t, err)
	user := NewUser("person@example.com", &hash, AuthProviderEmail, true)
	require.NoError(t, users.Create(context.Background(), user))

	err = service.SendPasswordResetCode(context.Background(), "person@example.com")
	require.NoError(t, err)
	require.NotEmpty(t, mailer.lastCode)

	resp, err := service.ResetPassword(context.Background(), ResetPasswordInput{
		Email:           "person@example.com",
		Code:            mailer.lastCode,
		Password:        "NewStrongPass1!",
		ConfirmPassword: "NewStrongPass1!",
	})
	require.NoError(t, err)
	require.NotEmpty(t, resp.AccessToken)
	require.True(t, resp.HasPassword)

	_, err = service.Login(context.Background(), LoginInput{
		Email:    "person@example.com",
		Password: "NewStrongPass1!",
	})
	require.NoError(t, err)
}

func TestResetPasswordCanAddPasswordForGoogleUser(t *testing.T) {
	users := newMemoryUserRepo()
	verifications := newMemoryVerificationRepo()
	mailer := &captureMailer{}
	service := NewService(
		users,
		verifications,
		newMemoryRefreshRepo(),
		nil,
		jwtutil.NewManager("issuer", "access", "refresh", 15*time.Minute, 7*24*time.Hour),
		mailer,
		stubGoogleVerifier{email: "google@example.com"},
	)

	_, err := service.GoogleLogin(context.Background(), GoogleLoginInput{IDToken: "token"})
	require.NoError(t, err)

	err = service.SendPasswordResetCode(context.Background(), "google@example.com")
	require.NoError(t, err)

	resp, err := service.ResetPassword(context.Background(), ResetPasswordInput{
		Email:           "google@example.com",
		Code:            mailer.lastCode,
		Password:        "StrongPass1!",
		ConfirmPassword: "StrongPass1!",
	})
	require.NoError(t, err)
	require.True(t, resp.HasPassword)

	_, err = service.Login(context.Background(), LoginInput{
		Email:    "google@example.com",
		Password: "StrongPass1!",
	})
	require.NoError(t, err)
}

func TestSetPasswordRejectsConfirmationMismatch(t *testing.T) {
	users := newMemoryUserRepo()
	service := NewService(
		users,
		newMemoryVerificationRepo(),
		newMemoryRefreshRepo(),
		nil,
		jwtutil.NewManager("issuer", "access", "refresh", 15*time.Minute, 7*24*time.Hour),
		noopMailer{},
		stubGoogleVerifier{email: "google@example.com"},
	)

	_, err := service.GoogleLogin(context.Background(), GoogleLoginInput{IDToken: "token"})
	require.NoError(t, err)

	user, err := users.FindByEmail(context.Background(), "google@example.com")
	require.NoError(t, err)

	err = service.SetPassword(context.Background(), user.ID, SetPasswordInput{
		Password:        "StrongPass1!",
		ConfirmPassword: "DifferentPass1!",
	})
	require.ErrorContains(t, err, "confirmation")
}

func TestRegisterDoesNotPersistUserWhenEmailSendFails(t *testing.T) {
	users := newMemoryUserRepo()
	verifications := newMemoryVerificationRepo()
	refreshes := newMemoryRefreshRepo()
	service := NewService(
		users,
		verifications,
		refreshes,
		memoryTxRunner{users: users, verifications: verifications, refreshes: refreshes},
		jwtutil.NewManager("issuer", "access", "refresh", 15*time.Minute, 7*24*time.Hour),
		errorMailer{err: fmt.Errorf("smtp down")},
		stubGoogleVerifier{},
	)

	err := service.Register(context.Background(), RegisterInput{
		Email:           "person@example.com",
		Password:        "StrongPass1!",
		ConfirmPassword: "StrongPass1!",
	})
	require.ErrorContains(t, err, "smtp down")

	_, err = users.FindByEmail(context.Background(), "person@example.com")
	require.ErrorIs(t, err, ErrNotFound)
	require.Empty(t, verifications.byUserID)
}
