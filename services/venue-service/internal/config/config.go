// Package config loads service configuration from environment variables.
// Call Load once in main.go; the resulting Config is passed into constructors
// and never read from the environment again outside this package.
package config

import "github.com/SamPariat/roundup/shared/env"

// Config holds all runtime configuration for the venue-service.
// Fields are populated once at startup and treated as read-only thereafter.
type Config struct {
	// Env is the deployment environment from APP_ENV: "beta" or "prod".
	Env string
	// Port is the HTTP listen port. Defaults to 3006.
	Port string
	// PostgresDSN is the full Postgres connection string. Required.
	PostgresDSN string
	// RedisAddr is the Redis server address. Defaults to localhost:6379.
	RedisAddr string
	// GoogleMapsAPIKey is the Google Maps Places API key. Required.
	GoogleMapsAPIKey string
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
		Env:              appEnv,
		Port:             env.Get("PORT", "3006"),
		PostgresDSN:      env.Require("POSTGRES_DSN"),
		RedisAddr:        env.Get("REDIS_ADDR", "localhost:6379"),
		GoogleMapsAPIKey: env.Require("GOOGLE_MAPS_API_KEY"),
		LogLevel:         env.Get("LOG_LEVEL", "info"),
	}
}
