// Package config loads service configuration from environment variables.
// Call Load once in main.go; the resulting Config is passed into constructors
// and never read from the environment again outside this package.
package config

import (
	"github.com/SamPariat/roundup/shared/api"
	"github.com/SamPariat/roundup/shared/env"
)

// Config holds all runtime configuration for the event-service.
// Fields are populated once at startup and treated as read-only thereafter.
type Config struct {
	// Env is the deployment environment from APP_ENV: "beta" or "prod".
	Env string

	// Port is the HTTP listen port. Defaults to 3003.
	Port string

	// PostgresDSN is the full Postgres connection string. Required.
	PostgresDSN string

	// UserServiceURL is user-service's base URL, e.g. "http://localhost:3002",
	// called directly to check squad membership. Required.
	UserServiceURL string

	// VenueServiceURL is venue-service's base URL, e.g.
	// "http://localhost:3006", called directly to snapshot suggested places
	// and record visits. Required.
	VenueServiceURL string

	// LogLevel controls zap verbosity: debug, info, warn, or error. Defaults to info.
	LogLevel string
}

// Load reads environment variables into a Config.
// It first loads .env.<APP_ENV> from the working directory (no-op if absent,
// e.g. in a deployed container).
// Panics immediately if a required variable is missing or empty.
func Load() *Config {
	appEnv := env.Load()

	return &Config{
		Env:             appEnv,
		Port:            env.Get(env.KeyPort, api.EventPort),
		PostgresDSN:     env.Require(env.KeyPostgresDSN),
		UserServiceURL:  env.Require("USER_SERVICE_URL"),
		VenueServiceURL: env.Require("VENUE_SERVICE_URL"),
		LogLevel:        env.Get(env.KeyLogLevel, env.DefaultLogLevel),
	}
}
