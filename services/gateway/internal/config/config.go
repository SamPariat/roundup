// Package config loads service configuration from environment variables.
// Call Load once in main.go; the resulting Config is passed into constructors
// and never read from the environment again outside this package.
package config

import "github.com/SamPariatIL/roundup/shared/go/env"

// Config holds all runtime configuration for the gateway.
// Fields are populated once at startup and treated as read-only thereafter.
type Config struct {
	// Port is the HTTP listen port. Defaults to 8080.
	Port string
	// RedisAddr is the Redis server address, used for token caching and
	// rate limiting. Defaults to localhost:6379.
	RedisAddr string
	// LogLevel controls zap verbosity: debug, info, warn, or error. Defaults to info.
	LogLevel string
}

// Load reads environment variables into a Config.
// It attempts to load a .env file first (no-op if absent, e.g., in production).
// Panics immediately if a required variable is missing or empty.
func Load() *Config {
	env.Load()

	return &Config{
		Port:      env.Get("PORT", "8080"),
		RedisAddr: env.Get("REDIS_ADDR", "localhost:6379"),
		LogLevel:  env.Get("LOG_LEVEL", "info"),
	}
}
