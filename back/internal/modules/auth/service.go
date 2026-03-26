package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"back/internal/platform/jwtutil"
	"back/internal/platform/mailer"
)

type Service struct {
	users     UserRepository
	codes     VerificationCodeRepository
	refreshes RefreshTokenRepository
	txRunner  TxRunner
	tokens    *jwtutil.Manager
	mailer    mailer.Sender
	google    GoogleVerifier
}

func NewService(users UserRepository, codes VerificationCodeRepository, refreshes RefreshTokenRepository, txRunner TxRunner, tokens *jwtutil.Manager, mailer mailer.Sender, google GoogleVerifier) *Service {
	return &Service{
		users:     users,
		codes:     codes,
		refreshes: refreshes,
		txRunner:  txRunner,
		tokens:    tokens,
		mailer:    mailer,
		google:    google,
	}
}

type RegisterInput struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SetPasswordInput struct {
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

type ResetPasswordInput struct {
	Email           string `json:"email"`
	Code            string `json:"code"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

type EmailCodeLoginInput struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type GoogleLoginInput struct {
	IDToken string `json:"id_token"`
}

type AuthResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ProfileCompleted bool   `json:"profile_completed"`
	HasPassword      bool   `json:"has_password"`
}

func (s *Service) Register(ctx context.Context, input RegisterInput) error {
	email := NormalizeEmail(input.Email)
	if err := ValidateEmail(email); err != nil {
		return err
	}
	if input.Password != input.ConfirmPassword {
		return fmt.Errorf("password confirmation does not match")
	}
	if err := ValidatePassword(input.Password, email); err != nil {
		return err
	}
	if _, err := s.users.FindByEmail(ctx, email); err == nil {
		return fmt.Errorf("user already exists")
	} else if err != ErrNotFound {
		return err
	}

	hash, err := HashPassword(input.Password)
	if err != nil {
		return err
	}

	register := func(ctx context.Context, users UserRepository, codes VerificationCodeRepository) error {
		user := NewUser(email, &hash, AuthProviderEmail, false)
		if err := users.Create(ctx, user); err != nil {
			return err
		}
		return s.issueAndSendCode(ctx, codes, user, s.mailer.SendVerificationCode)
	}

	if s.txRunner != nil {
		return s.txRunner.WithinTx(ctx, register)
	}
	return register(ctx, s.users, s.codes)
}

type VerifyEmailInput struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

func (s *Service) SendVerificationCode(ctx context.Context, email string) error {
	email = NormalizeEmail(email)
	if err := ValidateEmail(email); err != nil {
		return err
	}
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("user not found")
	}
	if user.EmailVerified {
		return nil
	}

	send := func(ctx context.Context, _ UserRepository, codes VerificationCodeRepository) error {
		return s.issueAndSendCode(ctx, codes, user, s.mailer.SendVerificationCode)
	}

	if s.txRunner != nil {
		return s.txRunner.WithinTx(ctx, send)
	}
	return send(ctx, s.users, s.codes)
}

func (s *Service) SendLoginCode(ctx context.Context, email string) error {
	email = NormalizeEmail(email)
	if err := ValidateEmail(email); err != nil {
		return err
	}
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	send := func(ctx context.Context, _ UserRepository, codes VerificationCodeRepository) error {
		return s.issueAndSendCode(ctx, codes, user, s.mailer.SendVerificationCode)
	}

	if s.txRunner != nil {
		return s.txRunner.WithinTx(ctx, send)
	}
	return send(ctx, s.users, s.codes)
}

func (s *Service) SendPasswordResetCode(ctx context.Context, email string) error {
	email = NormalizeEmail(email)
	if err := ValidateEmail(email); err != nil {
		return err
	}
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	send := func(ctx context.Context, _ UserRepository, codes VerificationCodeRepository) error {
		return s.issueAndSendCode(ctx, codes, user, s.mailer.SendPasswordResetCode)
	}

	if s.txRunner != nil {
		return s.txRunner.WithinTx(ctx, send)
	}
	return send(ctx, s.users, s.codes)
}

func (s *Service) VerifyEmail(ctx context.Context, input VerifyEmailInput) error {
	email := NormalizeEmail(input.Email)
	code := strings.TrimSpace(input.Code)
	if email == "" || code == "" {
		return fmt.Errorf("email and code are required")
	}
	if err := ValidateEmail(email); err != nil {
		return fmt.Errorf("invalid email")
	}

	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("invalid email or code")
	}
	if user.EmailVerified {
		return nil
	}

	verify := func(ctx context.Context, users UserRepository, codes VerificationCodeRepository) error {
		if err := s.consumeCode(ctx, codes, user.ID, code); err != nil {
			return err
		}
		if err := users.MarkEmailVerified(ctx, user.ID); err != nil {
			return err
		}
		return nil
	}

	if s.txRunner != nil {
		return s.txRunner.WithinTx(ctx, verify)
	}
	return verify(ctx, s.users, s.codes)
}

func (s *Service) LoginWithCode(ctx context.Context, input EmailCodeLoginInput) (AuthResponse, error) {
	email := NormalizeEmail(input.Email)
	code := strings.TrimSpace(input.Code)
	if email == "" || code == "" {
		return AuthResponse{}, fmt.Errorf("email and code are required")
	}
	if err := ValidateEmail(email); err != nil {
		return AuthResponse{}, fmt.Errorf("invalid email or code")
	}

	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return AuthResponse{}, fmt.Errorf("invalid email or code")
	}

	login := func(ctx context.Context, users UserRepository, codes VerificationCodeRepository) error {
		if err := s.consumeCode(ctx, codes, user.ID, code); err != nil {
			return err
		}
		if !user.EmailVerified {
			if err := users.MarkEmailVerified(ctx, user.ID); err != nil {
				return err
			}
			user.EmailVerified = true
		}
		return nil
	}

	if s.txRunner != nil {
		if err := s.txRunner.WithinTx(ctx, login); err != nil {
			return AuthResponse{}, err
		}
	} else if err := login(ctx, s.users, s.codes); err != nil {
		return AuthResponse{}, err
	}

	return s.newAuthResponse(ctx, user)
}

func (s *Service) Login(ctx context.Context, input LoginInput) (AuthResponse, error) {
	email := NormalizeEmail(input.Email)
	if err := ValidateEmail(email); err != nil {
		return AuthResponse{}, fmt.Errorf("invalid credentials")
	}
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return AuthResponse{}, fmt.Errorf("invalid credentials")
	}
	if user.PasswordHash == nil || CheckPassword(*user.PasswordHash, input.Password) != nil {
		return AuthResponse{}, fmt.Errorf("invalid credentials")
	}
	if !user.EmailVerified {
		return AuthResponse{}, fmt.Errorf("email is not verified")
	}
	return s.newAuthResponse(ctx, user)
}

func (s *Service) SetPassword(ctx context.Context, userID string, input SetPasswordInput) error {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return fmt.Errorf("unauthorized")
	}
	if input.Password != input.ConfirmPassword {
		return fmt.Errorf("password confirmation does not match")
	}

	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if err := ValidatePassword(input.Password, user.Email); err != nil {
		return err
	}

	hash, err := HashPassword(input.Password)
	if err != nil {
		return err
	}
	return s.users.UpdatePassword(ctx, user.ID, hash)
}

func (s *Service) ResetPassword(ctx context.Context, input ResetPasswordInput) (AuthResponse, error) {
	email := NormalizeEmail(input.Email)
	code := strings.TrimSpace(input.Code)
	if email == "" || code == "" {
		return AuthResponse{}, fmt.Errorf("email and code are required")
	}
	if err := ValidateEmail(email); err != nil {
		return AuthResponse{}, fmt.Errorf("invalid email or code")
	}
	if input.Password != input.ConfirmPassword {
		return AuthResponse{}, fmt.Errorf("password confirmation does not match")
	}

	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return AuthResponse{}, fmt.Errorf("invalid email or code")
	}
	if err := ValidatePassword(input.Password, user.Email); err != nil {
		return AuthResponse{}, err
	}

	hash, err := HashPassword(input.Password)
	if err != nil {
		return AuthResponse{}, err
	}

	reset := func(ctx context.Context, users UserRepository, codes VerificationCodeRepository) error {
		if err := s.consumeCode(ctx, codes, user.ID, code); err != nil {
			return err
		}
		if !user.EmailVerified {
			if err := users.MarkEmailVerified(ctx, user.ID); err != nil {
				return err
			}
			user.EmailVerified = true
		}
		if err := users.UpdatePassword(ctx, user.ID, hash); err != nil {
			return err
		}
		user.PasswordHash = &hash
		return nil
	}

	if s.txRunner != nil {
		if err := s.txRunner.WithinTx(ctx, reset); err != nil {
			return AuthResponse{}, err
		}
	} else if err := reset(ctx, s.users, s.codes); err != nil {
		return AuthResponse{}, err
	}

	return s.newAuthResponse(ctx, user)
}

func generateVerificationCode(digits int) (string, error) {
	if digits <= 0 || digits > 12 {
		return "", fmt.Errorf("invalid verification code length")
	}
	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(digits)), nil)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	format := fmt.Sprintf("%%0%dd", digits)
	return fmt.Sprintf(format, n.Int64()), nil
}

func hashVerificationCode(code string) (hashHex string, saltHex string, err error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", "", err
	}
	sum := sha256.Sum256(append(salt, []byte(code)...))
	return hex.EncodeToString(sum[:]), hex.EncodeToString(salt), nil
}

func verifyVerificationCode(code, saltHex, hashHex string) (bool, error) {
	salt, err := hex.DecodeString(saltHex)
	if err != nil {
		return false, fmt.Errorf("invalid verification code data")
	}
	want, err := hex.DecodeString(hashHex)
	if err != nil {
		return false, fmt.Errorf("invalid verification code data")
	}
	got := sha256.Sum256(append(salt, []byte(code)...))
	return subtle.ConstantTimeCompare(got[:], want) == 1, nil
}

func (s *Service) GoogleLogin(ctx context.Context, input GoogleLoginInput) (AuthResponse, error) {
	email, err := s.google.VerifyIDToken(input.IDToken)
	if err != nil {
		return AuthResponse{}, err
	}
	user, err := s.users.FindByEmail(ctx, email)
	if err == ErrNotFound {
		user = NewUser(email, nil, AuthProviderGoogle, true)
		if err := s.users.Create(ctx, user); err != nil {
			return AuthResponse{}, err
		}
	} else if err != nil {
		return AuthResponse{}, err
	}
	return s.newAuthResponse(ctx, user)
}

func (s *Service) issueAndSendCode(ctx context.Context, codes VerificationCodeRepository, user User, send func(email, code string) error) error {
	code, err := generateVerificationCode(6)
	if err != nil {
		return err
	}
	codeHash, codeSalt, err := hashVerificationCode(code)
	if err != nil {
		return err
	}
	record := VerificationCode{
		UserID:    user.ID,
		CodeHash:  codeHash,
		CodeSalt:  codeSalt,
		ExpiresAt: time.Now().UTC().Add(10 * time.Minute),
		Attempts:  0,
	}
	if err := codes.Save(ctx, record); err != nil {
		return err
	}
	return send(user.Email, code)
}

func (s *Service) consumeCode(ctx context.Context, codes VerificationCodeRepository, userID string, code string) error {
	record, err := codes.FindByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("invalid email or code")
	}
	if record.ExpiresAt.Before(time.Now().UTC()) {
		_ = codes.DeleteByUserID(ctx, userID)
		return fmt.Errorf("verification code expired")
	}
	ok, err := verifyVerificationCode(code, record.CodeSalt, record.CodeHash)
	if err != nil {
		return err
	}
	if !ok {
		attempts, err := codes.IncrementAttempts(ctx, userID)
		if err != nil {
			return err
		}
		if attempts >= 5 {
			_ = codes.DeleteByUserID(ctx, userID)
		}
		return fmt.Errorf("invalid email or code")
	}
	return codes.DeleteByUserID(ctx, userID)
}

func (s *Service) newAuthResponse(ctx context.Context, user User) (AuthResponse, error) {
	return s.newAuthResponseWithRepository(ctx, user, s.refreshes)
}

func (s *Service) newAuthResponseWithRepository(ctx context.Context, user User, refreshes RefreshTokenRepository) (AuthResponse, error) {
	pair, err := s.tokens.GeneratePair(user.ID)
	if err != nil {
		return AuthResponse{}, err
	}
	refreshClaims, err := s.tokens.ParseRefreshToken(pair.RefreshToken)
	if err != nil {
		return AuthResponse{}, err
	}
	if refreshes != nil {
		if err := refreshes.Save(ctx, RefreshSession{
			TokenID:   refreshClaims.ID,
			UserID:    user.ID,
			ExpiresAt: refreshClaims.ExpiresAt.Time,
		}); err != nil {
			return AuthResponse{}, err
		}
	}
	return AuthResponse{
		AccessToken:      pair.AccessToken,
		RefreshToken:     pair.RefreshToken,
		ProfileCompleted: user.ProfileCompleted,
		HasPassword:      user.PasswordHash != nil,
	}, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (AuthResponse, error) {
	claims, err := s.tokens.ParseRefreshToken(refreshToken)
	if err != nil {
		return AuthResponse{}, fmt.Errorf("invalid refresh token")
	}
	tokenID := strings.TrimSpace(claims.ID)
	if tokenID == "" {
		return AuthResponse{}, fmt.Errorf("invalid refresh token")
	}

	now := time.Now().UTC()
	if txRunner, ok := s.txRunner.(RefreshTxRunner); ok {
		var resp AuthResponse
		err := txRunner.WithinTxWithRefreshes(ctx, func(ctx context.Context, users UserRepository, _ VerificationCodeRepository, refreshes RefreshTokenRepository) error {
			user, err := users.FindByID(ctx, claims.UserID)
			if err != nil {
				return fmt.Errorf("user not found")
			}
			if err := refreshes.Consume(ctx, tokenID, user.ID, now); err != nil {
				return fmt.Errorf("invalid refresh token")
			}
			resp, err = s.newAuthResponseWithRepository(ctx, user, refreshes)
			return err
		})
		return resp, err
	}

	user, err := s.users.FindByID(ctx, claims.UserID)
	if err != nil {
		return AuthResponse{}, fmt.Errorf("user not found")
	}
	if err := s.refreshes.Consume(ctx, tokenID, user.ID, now); err != nil {
		return AuthResponse{}, fmt.Errorf("invalid refresh token")
	}
	return s.newAuthResponseWithRepository(ctx, user, s.refreshes)
}
