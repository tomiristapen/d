package auth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type GoogleVerifier interface {
	VerifyIDToken(idToken string) (string, error)
}

type GoogleIDTokenVerifier struct {
	audiences []string

	jwksURL string
	client  *http.Client

	mu        sync.Mutex
	keysByKID map[string]*rsa.PublicKey
	expiresAt time.Time
}

func NewGoogleIDTokenVerifier(clientIDs string) *GoogleIDTokenVerifier {
	return NewGoogleIDTokenVerifierWithOptions(clientIDs, "", nil)
}

func NewGoogleIDTokenVerifierWithOptions(clientIDs string, jwksURL string, client *http.Client) *GoogleIDTokenVerifier {
	auds := splitCSV(clientIDs)
	if jwksURL == "" {
		jwksURL = "https://www.googleapis.com/oauth2/v3/certs"
	}
	if client == nil {
		client = http.DefaultClient
	}
	return &GoogleIDTokenVerifier{
		audiences: auds,
		jwksURL:   jwksURL,
		client:    client,
		keysByKID: map[string]*rsa.PublicKey{},
	}
}

func (v *GoogleIDTokenVerifier) VerifyIDToken(idToken string) (string, error) {
	if len(v.audiences) == 0 {
		return "", fmt.Errorf("google client id is not configured")
	}

	parser := jwt.NewParser(jwt.WithValidMethods([]string{"RS256"}))

	var claims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		jwt.RegisteredClaims
	}

	keyfunc := func(token *jwt.Token) (any, error) {
		kid, _ := token.Header["kid"].(string)
		if kid == "" {
			return nil, fmt.Errorf("missing kid")
		}
		key, err := v.getKey(kid)
		if err != nil {
			return nil, err
		}
		return key, nil
	}

	parsed, err := parser.ParseWithClaims(idToken, &claims, keyfunc)
	if err != nil {
		return "", fmt.Errorf("invalid google id token: %w", err)
	}
	if !parsed.Valid {
		return "", fmt.Errorf("invalid google id token: token is not valid")
	}

	iss := strings.TrimSpace(claims.Issuer)
	if iss != "accounts.google.com" && iss != "https://accounts.google.com" {
		return "", fmt.Errorf("invalid google id token: invalid issuer %q", iss)
	}
	if !audienceAllowed(claims.Audience, v.audiences) {
		return "", fmt.Errorf("invalid google id token: audience not allowed (aud=%v, allowed=%v)", []string(claims.Audience), v.audiences)
	}
	if claims.ExpiresAt == nil {
		return "", fmt.Errorf("invalid google id token: missing exp")
	}
	if time.Now().After(claims.ExpiresAt.Time) {
		return "", fmt.Errorf("invalid google id token: token is expired")
	}

	email := strings.ToLower(strings.TrimSpace(claims.Email))
	if email == "" || !strings.Contains(email, "@") {
		return "", fmt.Errorf("invalid google id token: missing email")
	}
	if !claims.EmailVerified {
		return "", fmt.Errorf("google email is not verified")
	}
	return email, nil
}

func (v *GoogleIDTokenVerifier) getKey(kid string) (*rsa.PublicKey, error) {
	v.mu.Lock()
	key := v.keysByKID[kid]
	expired := time.Now().After(v.expiresAt)
	v.mu.Unlock()

	if key != nil && !expired {
		return key, nil
	}

	if err := v.refresh(); err != nil {
		// If refresh fails but we still have a cached key, allow it (best effort).
		v.mu.Lock()
		defer v.mu.Unlock()
		if key := v.keysByKID[kid]; key != nil && time.Now().Before(v.expiresAt.Add(5*time.Minute)) {
			return key, nil
		}
		return nil, err
	}

	v.mu.Lock()
	defer v.mu.Unlock()
	key = v.keysByKID[kid]
	if key == nil {
		return nil, fmt.Errorf("unknown kid")
	}
	return key, nil
}

func (v *GoogleIDTokenVerifier) refresh() error {
	req, err := http.NewRequest(http.MethodGet, v.jwksURL, nil)
	if err != nil {
		return err
	}
	resp, err := v.client.Do(req)
	if err != nil {
		return fmt.Errorf("fetch google jwks: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("fetch google jwks: status %d", resp.StatusCode)
	}

	var jwks struct {
		Keys []struct {
			KID string `json:"kid"`
			KTY string `json:"kty"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("decode google jwks: %w", err)
	}

	keys := make(map[string]*rsa.PublicKey, len(jwks.Keys))
	for _, k := range jwks.Keys {
		if k.KTY != "RSA" || k.KID == "" || k.N == "" || k.E == "" {
			continue
		}
		pub, err := jwkToRSAPublicKey(k.N, k.E)
		if err != nil {
			continue
		}
		keys[k.KID] = pub
	}
	if len(keys) == 0 {
		return fmt.Errorf("no google jwks keys")
	}

	expiresAt := time.Now().Add(1 * time.Hour)
	if maxAge := parseMaxAge(resp.Header.Get("Cache-Control")); maxAge > 0 {
		expiresAt = time.Now().Add(maxAge)
	}

	v.mu.Lock()
	v.keysByKID = keys
	v.expiresAt = expiresAt
	v.mu.Unlock()
	return nil
}

func jwkToRSAPublicKey(nB64URL, eB64URL string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nB64URL)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(eB64URL)
	if err != nil {
		return nil, err
	}
	e := 0
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}
	if e == 0 {
		return nil, fmt.Errorf("invalid exponent")
	}
	n := new(big.Int).SetBytes(nBytes)
	return &rsa.PublicKey{N: n, E: e}, nil
}

func parseMaxAge(cacheControl string) time.Duration {
	parts := strings.Split(cacheControl, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if !strings.HasPrefix(p, "max-age=") {
			continue
		}
		secStr := strings.TrimPrefix(p, "max-age=")
		sec, err := time.ParseDuration(secStr + "s")
		if err != nil || sec <= 0 {
			return 0
		}
		return sec
	}
	return 0
}

func splitCSV(value string) []string {
	var out []string
	for _, part := range strings.Split(value, ",") {
		// Allow inline comments in .env values, e.g. "id1,id2 # android/web".
		if i := strings.Index(part, "#"); i >= 0 {
			part = part[:i]
		}
		part = strings.TrimSpace(part)
		// Be tolerant to accidental "KEY=value" copy-pastes from `env`/`printenv` output.
		part = strings.TrimPrefix(part, "GOOGLE_CLIENT_ID=")
		part = strings.TrimSpace(part)
		part = strings.Trim(part, `"'`)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func audienceAllowed(tokenAud jwt.ClaimStrings, allowed []string) bool {
	for _, a := range tokenAud {
		for _, ok := range allowed {
			if a == ok {
				return true
			}
		}
	}
	return false
}
