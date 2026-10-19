// Package server holds the HTTP lifecycle every service shares: health and
// readiness endpoints for the platform, and serving until a shutdown signal
// with in-flight requests allowed to finish.
package server

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/SamPariat/roundup/shared/response"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// ShutdownTimeout bounds how long in-flight requests may run after a
// shutdown signal. Cloud Run and ECS both allow at least 10s before SIGKILL.
const ShutdownTimeout = 8 * time.Second

// readyTimeout bounds every readiness check so a hung dependency fails the
// probe instead of hanging it.
const readyTimeout = 2 * time.Second

// Check reports whether a dependency the service cannot work without is
// reachable, e.g. a Postgres ping.
type Check func(ctx context.Context) error

// Health registers two unauthenticated probes:
//   - GET /health (liveness): the process is up and serving. It never touches
//     dependencies, so a database outage doesn't make the platform restart
//     healthy containers in a loop.
//   - GET /ready (readiness): every check passes, so traffic may be sent.
//     Failing names the check and returns its error.
func Health(app *fiber.App, log *zap.Logger, checks map[string]Check) {
	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(response.OK(fiber.Map{"status": "ok"}))
	})

	app.Get("/ready", func(c fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), readyTimeout)
		defer cancel()

		for name, check := range checks {
			if err := check(ctx); err != nil {
				log.Warn("readiness check failed", zap.String("check", name), zap.Error(err))
				return c.Status(fiber.StatusServiceUnavailable).
					JSON(response.Fail[any](name + " not ready: " + err.Error()))
			}
		}
		return c.JSON(response.OK(fiber.Map{"status": "ready"}))
	})
}

// Run serves app on port until SIGINT or SIGTERM, then stops accepting
// connections and waits up to ShutdownTimeout for in-flight requests. It
// returns after shutdown, so callers close pools and clients afterwards.
func Run(app *fiber.App, port string, log *zap.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	app.Hooks().OnPreShutdown(func() error {
		log.Info("shutting down, draining in-flight requests")
		return nil
	})

	err := app.Listen(":"+port, fiber.ListenConfig{
		GracefulContext:       ctx,
		ShutdownTimeout:       ShutdownTimeout,
		DisableStartupMessage: true,
	})
	if err != nil {
		return err
	}

	log.Info("server stopped")
	return nil
}
