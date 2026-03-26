// Package logger constructs the service-wide zap logger.
// Call New once in main.go and inject the result into adapters and handlers.
package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New builds a production JSON logger at the given level.
// level must be a valid zapcore level string: debug, info, warn, or error.
// Panics at startup if the level is invalid or the logger cannot be built.
func New(level string) *zap.Logger {
	lvl, err := zapcore.ParseLevel(level)
	if err != nil {
		panic(fmt.Sprintf("invalid log level %q: %v", level, err))
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(lvl)

	log, err := cfg.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to build logger: %v", err))
	}

	return log
}
