package auth

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidatePassword(t *testing.T) {
	t.Run("accepts strong password", func(t *testing.T) {
		err := ValidatePassword("StrongPass1!", "person@example.com")
		require.NoError(t, err)
	})

	t.Run("rejects password containing email", func(t *testing.T) {
		err := ValidatePassword("Person123!@#", "person@example.com")
		require.ErrorContains(t, err, "must not contain the email")
	})

	t.Run("rejects common weak password", func(t *testing.T) {
		err := ValidatePassword("Password123!", "person@example.com")
		require.ErrorContains(t, err, "too common")
	})
}
