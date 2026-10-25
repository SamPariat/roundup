// Package main is the entry point for event-service.
// It wires together all layers (config, infrastructure, adapters, use-cases, handlers)
// and starts the HTTP server. All dependency injections happen here.
package main

import (
	"context"

	"github.com/SamPariat/roundup/services/event-service/internal/adapters/persistence"
	"github.com/SamPariat/roundup/services/event-service/internal/adapters/venues"
	"github.com/SamPariat/roundup/services/event-service/internal/application"
	"github.com/SamPariat/roundup/services/event-service/internal/config"
	"github.com/SamPariat/roundup/services/event-service/internal/domain"
	"github.com/SamPariat/roundup/services/event-service/internal/policy"
	transporthttp "github.com/SamPariat/roundup/services/event-service/internal/transport/http"
	"github.com/SamPariat/roundup/shared/api"
	"github.com/SamPariat/roundup/shared/authz"
	"github.com/SamPariat/roundup/shared/logger"
	"github.com/SamPariat/roundup/shared/middleware"
	"github.com/SamPariat/roundup/shared/postgres"
	"github.com/SamPariat/roundup/shared/server"
	"github.com/SamPariat/roundup/shared/userclient"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()
	log := logger.New(api.EventService, cfg.LogLevel)

	pool, err := postgres.Connect(context.Background(), cfg.PostgresDSN)
	if err != nil {
		log.Fatal("failed to connect to the database", zap.Error(err))
	}

	authorizer, err := authz.New[domain.Role, domain.Action](policy.CSV)
	if err != nil {
		log.Fatal("failed to load the authorization policy", zap.Error(err))
	}

	eventRepo := persistence.NewPostgresEventRepository(pool)
	squadDirectory := userclient.NewUserServiceClient(cfg.UserServiceURL)
	venueDirectory := venues.NewVenueServiceClient(cfg.VenueServiceURL)

	eventUC := application.NewEventUseCase(eventRepo, squadDirectory, venueDirectory, authorizer, log)

	app := fiber.New()
	app.Use(middleware.RequestID(), middleware.Logging(log), middleware.Recover(log))
	server.Health(app, log, map[string]server.Check{"postgres": pool.Ping})
	transporthttp.RegisterRoutes(app, &transporthttp.Handlers{
		Event: transporthttp.NewEventHandler(eventUC),
	})

	log.Info("starting "+api.EventService, zap.String("env", cfg.Env), zap.String("port", cfg.Port))

	if err := server.Run(app, cfg.Port, log); err != nil {
		log.Error("server exited with error", zap.Error(err))
	}

	pool.Close()
}
