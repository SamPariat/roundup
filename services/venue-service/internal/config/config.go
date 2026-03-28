// Package config loads service configuration from environment variables.
// Call Load once in main.go; the resulting Config is passed into constructors
// and never read from the environment again outside this package.
package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all runtime configuration for the venue-service.
// Fields are populated once at startup and treated as read-only thereafter.
type Config struct {
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
// It attempts to load a .env file first (no-op if absent, e.g., in production).
// Panics immediately if a required variable is missing or empty.
func Load() *Config {
	_ = godotenv.Load() // no-op when .env is absent; env vars from the platform take precedence

	return &Config{
		Port:             getEnv("PORT", "3006"),
		PostgresDSN:      requireEnv("POSTGRES_DSN"),
		RedisAddr:        getEnv("REDIS_ADDR", "localhost:6379"),
		GoogleMapsAPIKey: requireEnv("GOOGLE_MAPS_API_KEY"),
		LogLevel:         getEnv("LOG_LEVEL", "info"),
	}
}

// requireEnv returns the value of the key or panics if it is unset or empty.
func requireEnv(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		panic(fmt.Sprintf("required environment variable %s is not set", key))
	}

	return value
}

// getEnv returns the value of the key, or defaultValue if the key is unset or empty.
func getEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}

	return defaultValue
}
