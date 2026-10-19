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

// defaultUpstreamTimeout is used when UPSTREAM_TIMEOUT is unset.
const defaultUpstreamTimeout = "10s"

// Config holds all runtime configuration for the gateway.
// Fields are populated once at startup and treated as read-only thereafter.
type Config struct {
	// Env is the deployment environment from APP_ENV: "beta" or "prod".
	Env string

	// Port is the HTTP listen port. Defaults to 8080.
	Port string

	// LogLevel controls zap verbosity: debug, info, warn, or error. Defaults to info.
	LogLevel string

	// Upstreams are the base URLs requests are forwarded to. Required.
	Upstreams Upstreams

	// JWTPublicKey verifies access tokens minted by auth-service. Parsed from
	// the PEM in JWT_PUBLIC_KEY (see `make jwt-keys`). Required.
	JWTPublicKey ed25519.PublicKey

	// UpstreamTimeout bounds each forwarded request. Defaults to defaultUpstreamTimeout.
	UpstreamTimeout time.Duration
}

// Upstreams holds each service's base URL, e.g. "http://localhost:3002".
type Upstreams struct {
	Auth  string
	User  string
	Venue string
}

// Load reads environment variables into a Config.
// It first loads .env.<APP_ENV> from the working directory (no-op if absent,
// e.g. in a deployed container).
// Panics immediately if a required variable is missing, empty, or invalid.
func Load() *Config {
	appEnv := env.Load()

	publicKey, err := authtoken.ParsePublicKeyPEM(env.Require("JWT_PUBLIC_KEY"))
	if err != nil {
		panic(fmt.Sprintf("JWT_PUBLIC_KEY: %v", err))
	}

	return &Config{
		Env:      appEnv,
		Port:     env.Get(env.KeyPort, api.GatewayPort),
		LogLevel: env.Get(env.KeyLogLevel, env.DefaultLogLevel),
		Upstreams: Upstreams{
			Auth:  env.Require("AUTH_SERVICE_URL"),
			User:  env.Require("USER_SERVICE_URL"),
			Venue: env.Require("VENUE_SERVICE_URL"),
		},
		JWTPublicKey:    publicKey,
		UpstreamTimeout: env.Duration("UPSTREAM_TIMEOUT", defaultUpstreamTimeout),
	}
}
