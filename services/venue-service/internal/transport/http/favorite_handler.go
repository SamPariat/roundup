package http

import (
	"errors"

	"github.com/SamPariat/roundup/services/venue-service/internal/application"
	"github.com/SamPariat/roundup/services/venue-service/internal/domain"
	"github.com/SamPariat/roundup/services/venue-service/internal/transport/http/dtos"
	"github.com/SamPariat/roundup/shared/middleware"
	"github.com/SamPariat/roundup/shared/pagination"
	"github.com/SamPariat/roundup/shared/response"
	"github.com/gofiber/fiber/v3"
)

// FavoriteHandler handles HTTP requests for the favorite endpoints. The squad
// ID comes from the path and the caller's user ID from the X-User-ID header
// via middleware.UserIDFrom.
type FavoriteHandler struct {
	uc *application.FavoriteUseCase
}

// NewFavoriteHandler constructs a FavoriteHandler with the given use case.
func NewFavoriteHandler(uc *application.FavoriteUseCase) *FavoriteHandler {
	return &FavoriteHandler{uc: uc}
}

// Add handles POST /api/v1/squads/:squadID/favorites.
func (h *FavoriteHandler) Add(c fiber.Ctx) error {
	var req dtos.AddFavoriteRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	err := h.uc.Add(c.Context(), domain.AddFavoriteCommand{
		SquadID: c.Params("squadID"),
		UserID:  middleware.UserIDFrom(c),
		PlaceID: req.PlaceID,
	})
	if err != nil {
		return response.Error(c, favoriteErrorStatus(err), err)
	}

	return c.SendStatus(fiber.StatusCreated)
}

// Remove handles DELETE /api/v1/squads/:squadID/favorites/:placeID.
func (h *FavoriteHandler) Remove(c fiber.Ctx) error {
	err := h.uc.Remove(c.Context(), domain.RemoveFavoriteCommand{
		SquadID: c.Params("squadID"),
		UserID:  middleware.UserIDFrom(c),
		PlaceID: c.Params("placeID"),
	})
	if err != nil {
		return response.Error(c, favoriteErrorStatus(err), err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// List handles GET /api/v1/squads/:squadID/favorites?limit=&cursor=.
func (h *FavoriteHandler) List(c fiber.Ctx) error {
	page, err := pagination.FromQuery(c.Query("limit"), c.Query("cursor"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	favorites, next, err := h.uc.List(c.Context(), c.Params("squadID"), middleware.UserIDFrom(c), page)
	if err != nil {
		return response.Error(c, favoriteErrorStatus(err), err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(dtos.FavoritesToResponse(favorites, next)))
}

// favoriteErrorStatus maps domain sentinel errors to HTTP status codes for favorites.
func favoriteErrorStatus(err error) int {
	switch {
	case errors.Is(err, pagination.ErrInvalid):
		return fiber.StatusBadRequest
	case errors.Is(err, domain.ErrNotSaved), errors.Is(err, domain.ErrSquadNotFound), errors.Is(err, domain.ErrVenueNotFound):
		return fiber.StatusNotFound
	case errors.Is(err, domain.ErrAlreadySaved):
		return fiber.StatusConflict
	case errors.Is(err, domain.ErrUserServiceUnavailable), errors.Is(err, domain.ErrProviderUnavailable):
		return fiber.StatusServiceUnavailable
	default:
		return fiber.StatusInternalServerError
	}
}
