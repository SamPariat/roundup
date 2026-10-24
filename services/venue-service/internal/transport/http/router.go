package http

import (
	"github.com/SamPariat/roundup/shared/api"
	"github.com/SamPariat/roundup/shared/middleware"
	"github.com/gofiber/fiber/v3"
)

// Handlers group all HTTP handlers for dependency injection into RegisterRoutes.
// Add new handlers here as features are built — one field per handler.
type Handlers struct {
	Venue    *VenueHandler
	Favorite *FavoriteHandler
	History  *HistoryHandler
}

// RegisterRoutes mounts all route groups onto the Fiber app.
// All routes are versioned under /api/v1.
func RegisterRoutes(app *fiber.App, h *Handlers) {
	v1 := app.Group(api.V1)

	venues := v1.Group("/venues")
	venues.Get("/", h.Venue.Search)
	venues.Get("/:placeID", h.Venue.GetDetail)

	// Squad routes act on behalf of the caller, so they require X-User-ID
	// (set by the gateway from the verified access token).
	squads := v1.Group("/squads/:squadID", middleware.UserID())

	favorites := squads.Group("/favorites")
	favorites.Get("/", h.Favorite.List)
	favorites.Get("/:placeID/check", h.Favorite.IsSaved)
	favorites.Post("/", h.Favorite.Add)
	favorites.Delete("/:placeID", h.Favorite.Remove)

	history := squads.Group("/history")
	history.Get("/", h.History.List)
	history.Post("/", h.History.Record)
}
