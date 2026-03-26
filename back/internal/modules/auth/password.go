package auth

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var (
	upperRegex   = regexp.MustCompile(`[A-Z]`)
	lowerRegex   = regexp.MustCompile(`[a-z]`)
	numberRegex  = regexp.MustCompile(`[0-9]`)
	specialRegex = regexp.MustCompile(`[^A-Za-z0-9]`)
)

var commonWeakPasswords = map[string]struct{}{
	"password123!": {},
	"1234567890!a": {},
	"qwerty123!a":  {},
	"admin123!a":   {},
}

const (
	MinPasswordLength = 10
	MaxPasswordLength = 72
)

func ValidatePassword(password, email string) error {
	switch {
	case len(password) < MinPasswordLength:
		return fmt.Errorf("password must be at least %d characters", MinPasswordLength)
	case len(password) > MaxPasswordLength:
		return fmt.Errorf("password must be at most %d characters", MaxPasswordLength)
	case !upperRegex.MatchString(password):
		return fmt.Errorf("password must include an uppercase letter")
	case !lowerRegex.MatchString(password):
		return fmt.Errorf("password must include a lowercase letter")
	case !numberRegex.MatchString(password):
		return fmt.Errorf("password must include a number")
	case !specialRegex.MatchString(password):
		return fmt.Errorf("password must include a special character")
	case containsEmail(password, email):
		return fmt.Errorf("password must not contain the email")
	}

	if _, exists := commonWeakPasswords[strings.ToLower(password)]; exists {
		return fmt.Errorf("password is too common")
	}
	return nil
}

func HashPassword(password string) (string, error) {
	value, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(value), nil
}

func CheckPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func containsEmail(password, email string) bool {
	email = strings.ToLower(strings.TrimSpace(email))
	local := strings.Split(email, "@")[0]
	password = strings.ToLower(password)
	return email != "" && (strings.Contains(password, email) || (local != "" && strings.Contains(password, local)))
}
