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
	"github.com/SamPariat/roundup/shared/server"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()
	log := logger.New("gateway", cfg.LogLevel)

	app := fiber.New()
	// StripInternalHeaders runs first so RequestID always generates a new ID
	// instead of reusing one the client sent.
	app.Use(gatewaymw.StripInternalHeaders(), middleware.RequestID(), middleware.Logging(log), middleware.Recover(log))

	// The gateway has no dependency of its own to check; each upstream reports
	// its own readiness.
	server.Health(app, log, nil)

	transporthttp.RegisterRoutes(app, cfg.Upstreams, cfg.JWTPublicKey, cfg.UpstreamTimeout)

	log.Info("starting gateway", zap.String("env", cfg.Env), zap.String("port", cfg.Port))
	if err := server.Run(app, cfg.Port, log); err != nil {
		log.Error("server exited with error", zap.Error(err))
	}
}
