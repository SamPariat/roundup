// Package env provides small helpers for reading service configuration from
// environment variables and an optional per-environment file (.env.beta or
// .env.prod). Shared across every Go service's internal/config package, so the
// loading convention stays identical without each service re-implementing it.
package env

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// AppEnvKey is the environment variable that selects the deployment
// environment, and with it the file Load reads: .env.<APP_ENV>.
const AppEnvKey = "APP_ENV"

// Supported values of APP_ENV.
const (
	Beta = "beta"
	Prod = "prod"
)

// Keys every service reads, so each config spells them the same way.
const (
	KeyPort        = "PORT"
	KeyLogLevel    = "LOG_LEVEL"
	KeyPostgresDSN = "POSTGRES_DSN"
)

// DefaultLogLevel is used when LOG_LEVEL is unset.
const DefaultLogLevel = "info"

// Load reads .env.<APP_ENV> from the working directory into the process
// environment and returns the environment name. A missing file is not an
// error (deployed containers get real environment variables instead), and
// variables already set in the process always take precedence over the file.
// Panics if APP_ENV is not a supported value.
func Load() string {
	name, err := appEnv()
	if err != nil {
		panic(err.Error())
	}

	_ = godotenv.Load(".env." + name)

	return name
}

// appEnv returns the validated value of APP_ENV, or an error explaining why
// it cannot be used.
func appEnv() (string, error) {
	value, ok := os.LookupEnv(AppEnvKey)

	if !ok || value == "" {
		return "", fmt.Errorf("%s is not set", AppEnvKey)
	}

	if value != Beta && value != Prod {
		return "", fmt.Errorf("%s=%s is not a supported value", AppEnvKey, value)
	}

	return value, nil
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

// Duration returns the value of a key parsed as a Go duration (e.g. "15m"),
// or defaultValue parsed the same way if the key is unset or empty. Panics if
// the result is not a positive duration.
func Duration(key, defaultValue string) time.Duration {
	d, err := time.ParseDuration(Get(key, defaultValue))
	if err != nil || d <= 0 {
		panic(key + " must be a positive duration like 10s, 15m or 720h")
	}

	return d
}
