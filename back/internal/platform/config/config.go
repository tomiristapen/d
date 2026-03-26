package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	AppEnv           string
	HTTPPort         string
	DatabaseURL      string
	JWTAccessSecret  string
	JWTRefreshSecret string
	JWTIssuer        string
	AccessTokenTTL   time.Duration
	RefreshTokenTTL  time.Duration
	MailerFrom       string
	MailerBaseURL    string
	SMTPHost         string
	SMTPPort         string
	SMTPUser         string
	SMTPPass         string
	GoogleClientID   string
}

func Load() (Config, error) {
	accessTTL, err := time.ParseDuration(getEnv("ACCESS_TOKEN_TTL", "15m"))
	if err != nil {
		return Config{}, fmt.Errorf("parse access token ttl: %w", err)
	}
	refreshTTL, err := time.ParseDuration(getEnv("REFRESH_TOKEN_TTL", "168h"))
	if err != nil {
		return Config{}, fmt.Errorf("parse refresh token ttl: %w", err)
	}

	cfg := Config{
		AppEnv:           getEnv("APP_ENV", "development"),
		HTTPPort:         getEnv("HTTP_PORT", "8080"),
		DatabaseURL:      getEnv("DATABASE_URL", ""),
		JWTAccessSecret:  getEnv("JWT_ACCESS_SECRET", ""),
		JWTRefreshSecret: getEnv("JWT_REFRESH_SECRET", ""),
		JWTIssuer:        getEnv("JWT_ISSUER", "back-app"),
		AccessTokenTTL:   accessTTL,
		RefreshTokenTTL:  refreshTTL,
		MailerFrom:       getEnv("MAILER_FROM", "noreply@example.com"),
		MailerBaseURL:    getEnv("MAILER_BASE_URL", "http://localhost:8080"),
		SMTPHost:         getEnv("SMTP_HOST", ""),
		SMTPPort:         getEnv("SMTP_PORT", ""),
		SMTPUser:         getEnv("SMTP_USER", ""),
		SMTPPass:         getEnv("SMTP_PASS", ""),
		GoogleClientID:   getEnv("GOOGLE_CLIENT_ID", ""),
	}

	if cfg.DatabaseURL == "" || cfg.JWTAccessSecret == "" || cfg.JWTRefreshSecret == "" {
		return Config{}, fmt.Errorf("DATABASE_URL, JWT_ACCESS_SECRET, and JWT_REFRESH_SECRET are required")
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
