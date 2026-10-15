// Package main is the entry point for the gateway.
// It wires together config, logging, and routing, then starts the HTTP server.
// All dependency injections happen here.
package main

import (
	"github.com/SamPariat/roundup/services/gateway/internal/config"
	gatewaymw "github.com/SamPariat/roundup/services/gateway/internal/middleware"
	transporthttp "github.com/SamPariat/roundup/services/gateway/internal/transport/http"
	"github.com/SamPariat/roundup/shared/logger"
	"github.com/SamPariat/roundup/shared/middleware"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()
	log := logger.New("gateway", cfg.LogLevel)

	app := fiber.New()
	app.Use(middleware.RequestID())
	app.Use(middleware.Logging(log))
	app.Use(recover.New())
	app.Use(gatewaymw.StripIdentity())

	transporthttp.RegisterRoutes(app, cfg.Upstreams, cfg.JWTPublicKey, cfg.UpstreamTimeout)

	log.Info("starting gateway", zap.String("env", cfg.Env), zap.String("port", cfg.Port))
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal("server exited with error", zap.Error(err))
	}
}
