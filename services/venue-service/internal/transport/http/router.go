package http

import (
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
	v1 := app.Group("/api/v1")

	venues := v1.Group("/venues")
	venues.Get("/", h.Venue.Search)
	venues.Get("/:placeID", h.Venue.GetDetail)

	// Squad routes act on behalf of the caller, so they require X-User-ID
	// (set by the gateway from the verified access token).
	squads := v1.Group("/squads/:squadID", middleware.UserID())

	favorites := squads.Group("/favorites")
	favorites.Get("/", h.Favorite.ListFavorites)
	favorites.Get("/:placeID/check", h.Favorite.IsFavorite)
	favorites.Post("/", h.Favorite.AddFavorite)
	favorites.Delete("/:placeID", h.Favorite.RemoveFavorite)

	history := squads.Group("/history")
	history.Get("/", h.History.GetVisitHistory)
	history.Post("/", h.History.RecordVisit)
}
