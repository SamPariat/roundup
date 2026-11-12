// Package main is the entry point for venue-service.
// It wires together all layers (config, infrastructure, adapters, use-cases, handlers)
// and starts the HTTP server. All dependency injections happen here.
package main

import (
	"context"

	"github.com/SamPariat/roundup/services/venue-service/internal/adapters/cache"
	"github.com/SamPariat/roundup/services/venue-service/internal/adapters/persistence"
	"github.com/SamPariat/roundup/services/venue-service/internal/adapters/places"
	"github.com/SamPariat/roundup/services/venue-service/internal/application"
	"github.com/SamPariat/roundup/services/venue-service/internal/config"
	transporthttp "github.com/SamPariat/roundup/services/venue-service/internal/transport/http"
	"github.com/SamPariat/roundup/shared/api"
	"github.com/SamPariat/roundup/shared/logger"
	"github.com/SamPariat/roundup/shared/middleware"
	"github.com/SamPariat/roundup/shared/postgres"
	"github.com/SamPariat/roundup/shared/server"
	"github.com/SamPariat/roundup/shared/userclient"
	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// main wires venue-service: Google Places behind the Redis cache, Postgres
// for favorites and history, and user-service for squad membership, then
// serves the routes until shutdown.
func main() {
	cfg := config.Load()
	log := logger.New(api.VenueService, cfg.LogLevel)

	redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})

	pool, err := postgres.Connect(context.Background(), cfg.PostgresDSN)
	if err != nil {
		log.Fatal("failed to connect to the database", zap.Error(err))
	}

	venueRepo := persistence.NewPostgresVenueRepository(pool)

	placeProvider := places.NewGooglePlacesAdapter(places.BaseURL, cfg.GoogleMapsAPIKey)
	venueCache := cache.NewRedisCacheAdapter(redisClient, cache.DefaultTTL, log)
	squadDirectory := userclient.NewUserServiceClient(cfg.UserServiceURL)

	venueUC := application.NewVenueUseCase(placeProvider, venueCache)
	favoriteUC := application.NewFavoriteUseCase(venueRepo, squadDirectory, venueUC)
	historyUC := application.NewHistoryUseCase(venueRepo, squadDirectory, venueUC)

	app := fiber.New()
	app.Use(middleware.RequestID(), middleware.Logging(log, true), middleware.Recover(log))
	server.Health(app, log, map[string]server.Check{"postgres": pool.Ping})
	transporthttp.RegisterRoutes(app, &transporthttp.Handlers{
		Venue:    transporthttp.NewVenueHandler(venueUC, favoriteUC),
		Favorite: transporthttp.NewFavoriteHandler(favoriteUC, venueUC),
		History:  transporthttp.NewHistoryHandler(historyUC, venueUC),
	})

	log.Info("starting "+api.VenueService, zap.String("env", cfg.Env), zap.String("port", cfg.Port))

	if err := server.Run(app, cfg.Port, log); err != nil {
		log.Error("server exited with error", zap.Error(err))
	}

	pool.Close()
	_ = redisClient.Close()
}
