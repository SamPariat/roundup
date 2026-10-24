package http

import (
	"github.com/SamPariat/roundup/shared/api"
	"github.com/SamPariat/roundup/shared/middleware"
	"github.com/gofiber/fiber/v3"
)

// Handlers group all HTTP handlers for dependency injection into RegisterRoutes.
// Add new handlers here as features are built — one field per handler.
type Handlers struct {
	User        *UserHandler
	DeviceToken *DeviceTokenHandler
	Squad       *SquadHandler
}

// RegisterRoutes mounts all route groups onto the Fiber app.
// All routes are versioned under /api/v1.
func RegisterRoutes(app *fiber.App, h *Handlers) {
	v1 := app.Group(api.V1)

	users := v1.Group("/users")
	// Called by auth-service after phone verification. Not a public route —
	// the gateway must not expose it to the mobile app.
	users.Post("/", h.User.Create)
	users.Post("/lookup", middleware.UserID(), h.User.Lookup)

	me := users.Group("/me", middleware.UserID())
	me.Get("/", h.User.GetMe)
	me.Patch("/", h.User.UpdateMe)
	me.Delete("/", h.User.DeleteMe)

	deviceTokens := me.Group("/device-tokens")
	deviceTokens.Post("/", h.DeviceToken.Register)
	deviceTokens.Delete("/:token", h.DeviceToken.Remove)

	// Every squad route acts as the caller, so it requires X-User-ID. A squad
	// the caller is not in answers 404, and DELETE /members/:userID with the
	// caller's own ID is how they leave.
	squads := v1.Group("/squads", middleware.UserID())
	squads.Post("/", h.Squad.Create)
	squads.Get("/", h.Squad.List)
	squads.Get("/:squadID", h.Squad.Get)

	members := squads.Group("/:squadID/members")
	members.Post("/", h.Squad.AddMember)
	members.Patch("/:userID", h.Squad.ChangeRole)
	members.Delete("/:userID", h.Squad.RemoveMember)
}
