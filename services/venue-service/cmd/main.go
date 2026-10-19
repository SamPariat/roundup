// Package main is the entry point for venue-service.
// It wires together all layers (config, infrastructure, adapters, use-cases, handlers)
// and starts the HTTP server. All dependency injections happen here.
package main

import (
	"context"
	"time"

	"github.com/SamPariat/roundup/services/venue-service/internal/adapters/cache"
	"github.com/SamPariat/roundup/services/venue-service/internal/adapters/persistence"
	"github.com/SamPariat/roundup/services/venue-service/internal/adapters/places"
	"github.com/SamPariat/roundup/services/venue-service/internal/application"
	"github.com/SamPariat/roundup/services/venue-service/internal/config"
	transporthttp "github.com/SamPariat/roundup/services/venue-service/internal/transport/http"
	"github.com/SamPariat/roundup/shared/logger"
	"github.com/SamPariat/roundup/shared/middleware"
	"github.com/SamPariat/roundup/shared/postgres"
	"github.com/SamPariat/roundup/shared/server"
	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()
	log := logger.New("venue-service", cfg.LogLevel)

	redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})

	pool, err := postgres.Connect(context.Background(), cfg.PostgresDSN)
	if err != nil {
		log.Fatal("failed to connect to the database", zap.Error(err))
	}

	venueRepo := persistence.NewPostgresVenueRepository(pool)

	mapsAdapter, err := places.NewGoogleMapsAdapter(cfg.GoogleMapsAPIKey)
	if err != nil {
		log.Fatal("failed to init Google Maps adapter", zap.Error(err))
	}

	cacheAdapter := cache.NewRedisCacheAdapter(redisClient, 10*time.Minute, log)

	venueUC := application.NewVenueUseCase(mapsAdapter, cacheAdapter)
	favoriteUC := application.NewFavoriteUseCase(venueRepo)
	historyUC := application.NewHistoryUseCase(venueRepo)

	app := fiber.New()
	app.Use(middleware.RequestID(), middleware.Logging(log), middleware.Recover(log))
	server.Health(app, log, map[string]server.Check{"postgres": pool.Ping})
	transporthttp.RegisterRoutes(app, &transporthttp.Handlers{
		Venue:    transporthttp.NewVenueHandler(venueUC),
		Favorite: transporthttp.NewFavoriteHandler(favoriteUC),
		History:  transporthttp.NewHistoryHandler(historyUC),
	})

	log.Info("starting venue-service", zap.String("env", cfg.Env), zap.String("port", cfg.Port))
	if err := server.Run(app, cfg.Port, log); err != nil {
		log.Error("server exited with error", zap.Error(err))
	}
	pool.Close()
	_ = redisClient.Close()
}
