package auth

import (
	"fmt"
	"net/mail"
	"strings"
)

const MaxEmailLength = 254

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func ValidateEmail(email string) error {
	email = NormalizeEmail(email)
	if email == "" {
		return fmt.Errorf("email is required")
	}
	if len(email) > MaxEmailLength {
		return fmt.Errorf("email is too long")
	}

	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return fmt.Errorf("invalid email")
	}
	return nil
}
