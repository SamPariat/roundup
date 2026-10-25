package http

import (
	"github.com/SamPariat/roundup/shared/api"
	"github.com/SamPariat/roundup/shared/middleware"
	"github.com/gofiber/fiber/v3"
)

// Handlers group all HTTP handlers for dependency injection into RegisterRoutes.
// Add new handlers here as features are built — one field per handler.
type Handlers struct {
	Event *EventHandler
}

// RegisterRoutes mounts all route groups onto the Fiber app. Public routes
// are versioned under /api/v1 and act on behalf of the caller, so they
// require X-User-ID (set by the gateway from the verified access token).
// The /internal routes are for scheduler-service; the gateway doesn't route
// them, so they are reachable only on the internal network.
func RegisterRoutes(app *fiber.App, h *Handlers) {
	v1 := app.Group(api.V1, middleware.UserID())

	squadEvents := v1.Group("/squads/:squadID/events")
	squadEvents.Get("/", h.Event.List)
	squadEvents.Post("/", h.Event.Create)

	event := v1.Group("/events/:eventID")
	event.Get("/", h.Event.Get)
	event.Patch("/", h.Event.Update)
	event.Put("/rsvp", h.Event.SetRSVP)
	event.Post("/nominations", h.Event.Nominate)
	event.Delete("/nominations/:placeID", h.Event.RemoveNomination)
	event.Put("/vote", h.Event.Vote)
	event.Post("/voting/start", h.Event.StartVoting)
	event.Post("/voting/close", h.Event.CloseVoting)
	event.Post("/start", h.Event.StartNight)
	event.Post("/close", h.Event.Close)

	internal := app.Group(api.InternalV1 + "/events")
	internal.Post("/confirm-due", h.Event.ConfirmDue)
	internal.Post("/close-due", h.Event.CloseDue)
}
