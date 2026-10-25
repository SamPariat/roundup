// Package main is the entry point for user-service.
// It wires together all layers (config, infrastructure, adapters, use-cases, handlers)
// and starts the HTTP server. All dependency injections happen here.
package main

import (
	"context"

	"github.com/SamPariat/roundup/services/user-service/internal/adapters/authz"
	"github.com/SamPariat/roundup/services/user-service/internal/adapters/persistence"
	"github.com/SamPariat/roundup/services/user-service/internal/application"
	"github.com/SamPariat/roundup/services/user-service/internal/config"
	transporthttp "github.com/SamPariat/roundup/services/user-service/internal/transport/http"
	"github.com/SamPariat/roundup/shared/api"
	"github.com/SamPariat/roundup/shared/logger"
	"github.com/SamPariat/roundup/shared/middleware"
	"github.com/SamPariat/roundup/shared/postgres"
	"github.com/SamPariat/roundup/shared/server"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()
	log := logger.New(api.UserService, cfg.LogLevel)

	pool, err := postgres.Connect(context.Background(), cfg.PostgresDSN)
	if err != nil {
		log.Fatal("failed to connect to the database", zap.Error(err))
	}

	// PostgresUserRepository implements both UserRepository and DeviceTokenRepository.
	userRepo := persistence.NewPostgresUserRepository(pool)
	squadRepo := persistence.NewPostgresSquadRepository(pool)

	authorizer, err := authz.NewCasbinAuthorizer()
	if err != nil {
		log.Fatal("failed to load the authorization policy", zap.Error(err))
	}

	userUC := application.NewUserUseCase(userRepo, userRepo)
	deviceTokenUC := application.NewDeviceTokenUseCase(userRepo)
	squadUC := application.NewSquadUseCase(squadRepo, userRepo, authorizer)

	app := fiber.New()
	app.Use(middleware.RequestID(), middleware.Logging(log), middleware.Recover(log))
	server.Health(app, log, map[string]server.Check{"postgres": pool.Ping})
	transporthttp.RegisterRoutes(app, &transporthttp.Handlers{
		User:        transporthttp.NewUserHandler(userUC),
		DeviceToken: transporthttp.NewDeviceTokenHandler(deviceTokenUC),
		Squad:       transporthttp.NewSquadHandler(squadUC),
	})

	log.Info("starting "+api.UserService, zap.String("env", cfg.Env), zap.String("port", cfg.Port))

	if err := server.Run(app, cfg.Port, log); err != nil {
		log.Error("server exited with error", zap.Error(err))
	}

	pool.Close()
}
