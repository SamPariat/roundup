// Package main is the entry point for the gateway.
// It wires together config, logging, and routing, then starts the HTTP server.
// All dependency injections happen here.
package main

import (
	"github.com/SamPariatIL/roundup/services/gateway/internal/config"
	transporthttp "github.com/SamPariatIL/roundup/services/gateway/internal/transport/http"
	"github.com/SamPariatIL/roundup/shared/go/logger"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()
	log := logger.New("gateway", cfg.LogLevel)

	app := fiber.New()
	transporthttp.RegisterRoutes(app)

	log.Info("starting gateway", zap.String("port", cfg.Port))
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal("server exited with error", zap.Error(err))
	}
}
