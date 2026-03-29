package http

import "github.com/gofiber/fiber/v3"

// router.go registers all HTTP routes for the venue-service.
// Add new handlers to the Handlers struct and register their routes here.

// Handlers group all HTTP handlers for dependency injection into RegisterRoutes.
// Add new handlers here as features are built — one field per handler.
type Handlers struct {
	Venue *VenueHandler
}

// RegisterRoutes mounts all route groups onto the Fiber app.
// All routes are versioned under /api/v1.
func RegisterRoutes(app *fiber.App, h *Handlers) {
	v1 := app.Group("/api/v1")

	venues := v1.Group("/venues")
	venues.Get("/", h.Venue.Search)
	venues.Get("/:placeID", h.Venue.GetDetail)
}
