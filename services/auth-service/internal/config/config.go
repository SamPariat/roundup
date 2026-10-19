// Package config loads service configuration from environment variables.
// Call Load once in main.go; the resulting Config is passed into constructors
// and never read from the environment again outside this package.
package config

import (
	"crypto/ed25519"
	"fmt"
	"time"

	"github.com/SamPariat/roundup/shared/api"
	"github.com/SamPariat/roundup/shared/authtoken"
	"github.com/SamPariat/roundup/shared/env"
)

// Config holds all runtime configuration for the auth-service.
// Fields are populated once at startup and treated as read-only thereafter.
type Config struct {
	// Env is the deployment environment from APP_ENV: "beta" or "prod".
	Env string

	// Port is the HTTP listen port. Defaults to 3001.
	Port string

	// PostgresDSN is the full Postgres connection string. Required.
	PostgresDSN string

	// UserServiceURL is user-service's base URL, called directly (not via the
	// gateway) to find or create users at login. Required.
	UserServiceURL string

	// JWTPrivateKey signs access tokens. Parsed from the PEM in
	// JWT_PRIVATE_KEY (see `make jwt-keys`). Required.
	JWTPrivateKey ed25519.PrivateKey

	// AccessTokenTTL is the access token lifetime. Defaults to 15m.
	AccessTokenTTL time.Duration

	// RefreshTokenTTL is the refresh token lifetime. Defaults to 720h (30 days).
	RefreshTokenTTL time.Duration

	// OTPDevCode, when set, is accepted as a valid code for any phone.
	// Only allowed when Env is beta; Load panics if it is set in prod.
	OTPDevCode string

	// LogLevel controls zap verbosity: debug, info, warn, or error. Defaults to info.
	LogLevel string
}

// Load reads environment variables into a Config.
// It first loads .env.<APP_ENV> from the working directory (no-op if absent,
// e.g. in a deployed container).
// Panics immediately if a required variable is missing, empty, or invalid.
func Load() *Config {
	appEnv := env.Load()

	privateKey, err := authtoken.ParsePrivateKeyPEM(env.Require("JWT_PRIVATE_KEY"))
	if err != nil {
		panic(fmt.Sprintf("JWT_PRIVATE_KEY: %v", err))
	}

	cfg := &Config{
		Env:             appEnv,
		Port:            env.Get(env.KeyPort, api.AuthPort),
		PostgresDSN:     env.Require(env.KeyPostgresDSN),
		UserServiceURL:  env.Require("USER_SERVICE_URL"),
		JWTPrivateKey:   privateKey,
		AccessTokenTTL:  env.Duration("ACCESS_TOKEN_TTL", "15m"),
		RefreshTokenTTL: env.Duration("REFRESH_TOKEN_TTL", "720h"),
		OTPDevCode:      env.Get("OTP_DEV_CODE", ""),
		LogLevel:        env.Get(env.KeyLogLevel, env.DefaultLogLevel),
	}

	// A fixed code that logs anyone in must never reach production.
	if cfg.Env == env.Prod && cfg.OTPDevCode != "" {
		panic("OTP_DEV_CODE must not be set when APP_ENV=prod")
	}

	return cfg
}
