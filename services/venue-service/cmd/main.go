// Package main is the entry point for venue-service.
// It wires together all layers (config, infrastructure, adapters, use-cases, handlers)
// and starts the HTTP server. All dependency injections happen here.
package main

import (
	"time"

	"github.com/SamPariatIL/roundup/services/venue-service/internal/adapters/cache"
	"github.com/SamPariatIL/roundup/services/venue-service/internal/adapters/places"
	"github.com/SamPariatIL/roundup/services/venue-service/internal/application"
	"github.com/SamPariatIL/roundup/services/venue-service/internal/config"
	"github.com/SamPariatIL/roundup/services/venue-service/internal/logger"
	transporthttp "github.com/SamPariatIL/roundup/services/venue-service/internal/transport/http"
	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.LogLevel)

	redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})

	mapsAdapter, err := places.NewGoogleMapsAdapter(cfg.GoogleMapsAPIKey)
	if err != nil {
		log.Fatal("failed to init Google Maps adapter", zap.Error(err))
	}

	cacheAdapter := cache.NewRedisCacheAdapter(redisClient, 10*time.Minute, log)

	venueUC := application.NewVenueUseCase(mapsAdapter, cacheAdapter)

	app := fiber.New()
	transporthttp.RegisterRoutes(app, &transporthttp.Handlers{
		Venue: transporthttp.NewVenueHandler(venueUC),
	})

	log.Info("starting venue-service", zap.String("port", cfg.Port))
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal("server exited with error", zap.Error(err))
	}
}
