package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"back/internal/platform/authctx"
	"back/internal/platform/jwtutil"

	"github.com/stretchr/testify/require"
)

func TestPostSetPasswordRequiresAuthentication(t *testing.T) {
	handler := NewHandler(new(Service))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/set-password", strings.NewReader(`{"password":"StrongPass1!","confirm_password":"StrongPass1!"}`))
	rr := httptest.NewRecorder()
	handler.PostSetPassword(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
	require.Contains(t, rr.Body.String(), `"code":"UNAUTHORIZED"`)
}

func TestPostSetPasswordSetsPasswordForGoogleUser(t *testing.T) {
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

	handler := NewHandler(service)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/set-password", strings.NewReader(`{"password":"StrongPass1!","confirm_password":"StrongPass1!"}`))
	req = req.WithContext(authctx.WithUserID(req.Context(), user.ID))
	rr := httptest.NewRecorder()
	handler.PostSetPassword(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Contains(t, rr.Body.String(), `"status":"password_set"`)

	_, err = service.Login(context.Background(), LoginInput{
		Email:    "google@example.com",
		Password: "StrongPass1!",
	})
	require.NoError(t, err)
}
