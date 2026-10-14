// Package main is the entry point for user-service.
// It wires together all layers (config, infrastructure, adapters, use-cases, handlers)
// and starts the HTTP server. All dependency injections happen here.
package main

import (
	"context"

	"github.com/SamPariat/roundup/services/user-service/internal/adapters/persistence"
	"github.com/SamPariat/roundup/services/user-service/internal/application"
	"github.com/SamPariat/roundup/services/user-service/internal/config"
	transporthttp "github.com/SamPariat/roundup/services/user-service/internal/transport/http"
	"github.com/SamPariat/roundup/shared/logger"
	"github.com/SamPariat/roundup/shared/middleware"
	"github.com/SamPariat/roundup/shared/postgres"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()
	log := logger.New("user-service", cfg.LogLevel)

	pool, err := postgres.Connect(context.Background(), cfg.PostgresDSN)
	if err != nil {
		log.Fatal("failed to connect to the database", zap.Error(err))
	}

	repo := persistence.NewPostgresUserRepository(pool)

	userUC := application.NewUserUseCase(repo, repo)
	deviceTokenUC := application.NewDeviceTokenUseCase(repo)

	app := fiber.New()
	app.Use(middleware.RequestID(), middleware.Logging(log))
	transporthttp.RegisterRoutes(app, &transporthttp.Handlers{
		User:        transporthttp.NewUserHandler(userUC),
		DeviceToken: transporthttp.NewDeviceTokenHandler(deviceTokenUC),
	})

	log.Info("starting user-service", zap.String("env", cfg.Env), zap.String("port", cfg.Port))
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal("server exited with error", zap.Error(err))
	}
}
