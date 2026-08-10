// Package env provides small helpers for reading service configuration from
// environment variables and an optional .env file. Shared across every Go
// service's internal/config package, so the loading convention stays
// identical without each service re-implementing it.
package env

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Load reads a .env file into the process environment if one is present.
// It is a no-op when no .env file exists (e.g., in production), and existing
// environment variables always take precedence over values in the file.
func Load() {
	_ = godotenv.Load()
}

// Require returns the value of a key or panics if it is unset or empty.
func Require(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		panic(fmt.Sprintf("required environment variable %s is not set", key))
	}

	return value
}

// Get returns the value of a key, or defaultValue if the key is unset or empty.
func Get(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}

	return defaultValue
}
