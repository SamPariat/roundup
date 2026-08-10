// Package http registers the gateway's HTTP routes.
// Route groups mirror the upstream services they will eventually proxy to;
// until proxying is wired up (see ROUNDUP-TASKS.md → API Gateway → Routing),
// each group falls through to a placeholder handler.
package http

import (
	"github.com/gofiber/fiber/v3"
)

// RegisterRoutes mounts /health and the versioned route groups onto the Fiber app.
func RegisterRoutes(app *fiber.App) {
	app.Get("/health", healthHandler)

	v1 := app.Group("/api/v1")

	routeGroups := []string{"auth", "users", "events", "billing", "venues", "media", "realtime"}
	for _, name := range routeGroups {
		v1.Group("/"+name).All("/*", notImplementedHandler)
	}
}

// healthHandler reports liveness for orchestration and uptime checks.
func healthHandler(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}

// notImplementedHandler responds for route groups whose upstream proxy
// hasn't been wired up yet.
//
// The route groups above proxy to
// services that mostly don't exist yet (auth-service, user-service, etc.),
// so every request under /api/v1/* currently lands here. Consider: status
// code (501 "not built yet" vs. 503 "unavailable" vs. something else), and
// whether the JSON body should follow the envelope convention used by
// venue-service (internal/transport/http/dtos.Response in that service) so
// gateway responses stay consistent with the rest of the API once proxying
// lands.
func notImplementedHandler(c fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not implemented"})
}
