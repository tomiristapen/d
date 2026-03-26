package jwtutil

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

type Claims struct {
	UserID string    `json:"user_id"`
	Type   TokenType `json:"type"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Manager struct {
	issuer        string
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

func NewManager(issuer, accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration) *Manager {
	return &Manager{
		issuer:        issuer,
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
	}
}

func (m *Manager) GeneratePair(userID string) (TokenPair, error) {
	accessToken, err := m.sign(userID, TokenTypeAccess, m.accessSecret, m.accessTTL)
	if err != nil {
		return TokenPair{}, err
	}
	refreshToken, err := m.sign(userID, TokenTypeRefresh, m.refreshSecret, m.refreshTTL)
	if err != nil {
		return TokenPair{}, err
	}
	return TokenPair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (m *Manager) ParseAccessToken(token string) (Claims, error) {
	return m.parse(token, m.accessSecret, TokenTypeAccess)
}

func (m *Manager) ParseRefreshToken(token string) (Claims, error) {
	return m.parse(token, m.refreshSecret, TokenTypeRefresh)
}

func (m *Manager) sign(userID string, tokenType TokenType, secret []byte, ttl time.Duration) (string, error) {
	now := time.Now()
	tokenID := ""
	if tokenType == TokenTypeRefresh {
		tokenID = uuid.NewString()
	}
	claims := Claims{
		UserID: userID,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   userID,
			ID:        tokenID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func (m *Manager) parse(tokenString string, secret []byte, expectedType TokenType) (Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		return Claims{}, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return Claims{}, fmt.Errorf("invalid token")
	}
	if claims.Type != expectedType {
		return Claims{}, fmt.Errorf("unexpected token type")
	}
	return *claims, nil
}
