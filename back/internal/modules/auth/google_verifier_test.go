package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

func TestGoogleIDTokenVerifier_VerifiesTokenAndReturnsEmail(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	kid := "testkid"
	jwks := map[string]any{
		"keys": []map[string]any{
			{
				"kty": "RSA",
				"kid": kid,
				"n":   base64.RawURLEncoding.EncodeToString(privateKey.PublicKey.N.Bytes()),
				"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(privateKey.PublicKey.E)).Bytes()),
			},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "max-age=3600")
		_ = json.NewEncoder(w).Encode(jwks)
	}))
	t.Cleanup(srv.Close)

	clientID := "client-123"
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss":            "https://accounts.google.com",
		"aud":            clientID,
		"exp":            time.Now().Add(5 * time.Minute).Unix(),
		"iat":            time.Now().Add(-1 * time.Minute).Unix(),
		"email":          "Person@Example.com",
		"email_verified": true,
	})
	token.Header["kid"] = kid
	signed, err := token.SignedString(privateKey)
	require.NoError(t, err)

	verifier := NewGoogleIDTokenVerifierWithOptions(clientID, srv.URL, srv.Client())
	email, err := verifier.VerifyIDToken(signed)
	require.NoError(t, err)
	require.Equal(t, "person@example.com", email)
}

func TestGoogleIDTokenVerifier_RejectsWrongAudience(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	kid := "testkid"
	jwks := map[string]any{
		"keys": []map[string]any{
			{
				"kty": "RSA",
				"kid": kid,
				"n":   base64.RawURLEncoding.EncodeToString(privateKey.PublicKey.N.Bytes()),
				"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(privateKey.PublicKey.E)).Bytes()),
			},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(jwks)
	}))
	t.Cleanup(srv.Close)

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss":            "accounts.google.com",
		"aud":            "not-allowed",
		"exp":            time.Now().Add(5 * time.Minute).Unix(),
		"iat":            time.Now().Add(-1 * time.Minute).Unix(),
		"email":          "person@example.com",
		"email_verified": true,
	})
	token.Header["kid"] = kid
	signed, err := token.SignedString(privateKey)
	require.NoError(t, err)

	verifier := NewGoogleIDTokenVerifierWithOptions("allowed-1,allowed-2", srv.URL, srv.Client())
	_, err = verifier.VerifyIDToken(signed)
	require.Error(t, err)
}

func TestGoogleIDTokenVerifier_RejectsUnverifiedEmail(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	kid := "testkid"
	jwks := map[string]any{
		"keys": []map[string]any{
			{
				"kty": "RSA",
				"kid": kid,
				"n":   base64.RawURLEncoding.EncodeToString(privateKey.PublicKey.N.Bytes()),
				"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(privateKey.PublicKey.E)).Bytes()),
			},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(jwks)
	}))
	t.Cleanup(srv.Close)

	clientID := "client-123"
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss":            "accounts.google.com",
		"aud":            clientID,
		"exp":            time.Now().Add(5 * time.Minute).Unix(),
		"iat":            time.Now().Add(-1 * time.Minute).Unix(),
		"email":          "person@example.com",
		"email_verified": false,
	})
	token.Header["kid"] = kid
	signed, err := token.SignedString(privateKey)
	require.NoError(t, err)

	verifier := NewGoogleIDTokenVerifierWithOptions(clientID, srv.URL, srv.Client())
	_, err = verifier.VerifyIDToken(signed)
	require.Error(t, err)
}

func TestSplitCSV_StripsGoogleClientIDPrefix(t *testing.T) {
	items := splitCSV(`GOOGLE_CLIENT_ID=client-1, client-2`)
	require.Equal(t, []string{"client-1", "client-2"}, items)
}
