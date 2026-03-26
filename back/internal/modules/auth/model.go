package auth

import "time"

type AuthProvider string

const (
	AuthProviderEmail  AuthProvider = "email"
	AuthProviderGoogle AuthProvider = "google"
)

type User struct {
	ID               string       `json:"id"`
	Email            string       `json:"email"`
	PasswordHash     *string      `json:"-"`
	AuthProvider     AuthProvider `json:"auth_provider"`
	EmailVerified    bool         `json:"email_verified"`
	ProfileCompleted bool         `json:"profile_completed"`
	CreatedAt        time.Time    `json:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at"`
}

type VerificationCode struct {
	UserID    string
	CodeHash  string
	CodeSalt  string
	ExpiresAt time.Time
	Attempts  int
}

type RefreshSession struct {
	TokenID    string
	UserID     string
	ExpiresAt  time.Time
	ConsumedAt *time.Time
	CreatedAt  time.Time
}
