package http

import (
	"github.com/SamPariat/roundup/shared/middleware"
	"github.com/gofiber/fiber/v3"
)

// router.go registers all HTTP routes for the user-service.
// Add new handlers to the Handlers struct and register their routes here.

// Handlers group all HTTP handlers for dependency injection into RegisterRoutes.
// Add new handlers here as features are built — one field per handler.
type Handlers struct {
	User        *UserHandler
	DeviceToken *DeviceTokenHandler
}

// RegisterRoutes mounts all route groups onto the Fiber app.
// All routes are versioned under /api/v1.
func RegisterRoutes(app *fiber.App, h *Handlers) {
	v1 := app.Group("/api/v1")

	users := v1.Group("/users")
	// Called by auth-service after phone verification. Not a public route —
	// the gateway must not expose it to the mobile app.
	users.Post("/", h.User.Create)

	me := users.Group("/me", middleware.UserID())
	me.Get("/", h.User.GetMe)
	me.Patch("/", h.User.UpdateMe)
	me.Delete("/", h.User.DeleteMe)

	deviceTokens := me.Group("/device-tokens")
	deviceTokens.Post("/", h.DeviceToken.Register)
	deviceTokens.Delete("/:token", h.DeviceToken.Remove)
}
