// Package main is the entry point for auth-service.
// It wires together all layers (config, infrastructure, adapters, use-cases, handlers)
// and starts the HTTP server. All dependency injections happen here.
package main

import (
	"context"

	"github.com/SamPariat/roundup/services/auth-service/internal/adapters/persistence"
	"github.com/SamPariat/roundup/services/auth-service/internal/adapters/sms"
	"github.com/SamPariat/roundup/services/auth-service/internal/adapters/users"
	"github.com/SamPariat/roundup/services/auth-service/internal/application"
	"github.com/SamPariat/roundup/services/auth-service/internal/config"
	transporthttp "github.com/SamPariat/roundup/services/auth-service/internal/transport/http"
	"github.com/SamPariat/roundup/shared/env"
	"github.com/SamPariat/roundup/shared/logger"
	"github.com/SamPariat/roundup/shared/middleware"
	"github.com/SamPariat/roundup/shared/postgres"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()
	log := logger.New("auth-service", cfg.LogLevel)

	// Only a console sender exists so far; prod needs a real SMS provider.
	if cfg.Env == env.Prod {
		log.Fatal("no SMS provider configured for prod; the console sender is beta-only")
	}

	pool, err := postgres.Connect(context.Background(), cfg.PostgresDSN)
	if err != nil {
		log.Fatal("failed to connect to the database", zap.Error(err))
	}

	otpRepo := persistence.NewPostgresOTPRepository(pool)
	sessionRepo := persistence.NewPostgresSessionRepository(pool)
	sender := sms.NewConsoleSender(log)
	userDirectory := users.NewUserServiceClient(cfg.UserServiceURL)

	issuer := application.NewTokenIssuer(cfg.JWTPrivateKey, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	otpUC := application.NewOTPUseCase(otpRepo, sender, userDirectory, sessionRepo, issuer, cfg.OTPDevCode)
	tokenUC := application.NewTokenUseCase(sessionRepo, issuer)

	app := fiber.New()
	app.Use(middleware.RequestID(), middleware.Logging(log))
	transporthttp.RegisterRoutes(app, &transporthttp.Handlers{
		Auth: transporthttp.NewAuthHandler(otpUC, tokenUC),
	})

	log.Info("starting auth-service", zap.String("env", cfg.Env), zap.String("port", cfg.Port))
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal("server exited with error", zap.Error(err))
	}
}
