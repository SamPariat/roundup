// Package main is the entry point for venue-service.
// It wires together all layers (config, infrastructure, adapters, use-cases, handlers)
// and starts the HTTP server. All dependency injections happen here.
package main

import (
	"github.com/SamPariatIL/roundup/services/venue-service/internal/config"
	"github.com/SamPariatIL/roundup/services/venue-service/internal/logger"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.LogLevel)

	log.Info("Started venue-service")
}
