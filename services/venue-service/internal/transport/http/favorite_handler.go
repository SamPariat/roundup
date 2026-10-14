package http

import (
	"errors"

	"github.com/SamPariat/roundup/services/venue-service/internal/application"
	"github.com/SamPariat/roundup/services/venue-service/internal/domain"
	"github.com/SamPariat/roundup/services/venue-service/internal/transport/http/dtos"
	"github.com/SamPariat/roundup/shared/middleware"
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

// AddFavorite handles POST /api/v1/squads/:squadID/favorites.
func (h *FavoriteHandler) AddFavorite(c fiber.Ctx) error {
	squadID := c.Params("squadID")

	var req dtos.AddFavoriteRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.Fail[any](err.Error()),
		)
	}

	err := h.uc.AddFavorite(c.Context(), domain.AddFavoriteCommand{
		SquadID: squadID,
		UserID:  middleware.UserIDFrom(c),
		PlaceID: req.PlaceID,
		Name:    req.Name,
	})
	if err != nil {
		return c.Status(favoriteErrorStatus(err)).JSON(
			response.Fail[any](err.Error()),
		)
	}

	return c.SendStatus(fiber.StatusCreated)
}

// RemoveFavorite handles DELETE /api/v1/squads/:squadID/favorites/:placeID.
func (h *FavoriteHandler) RemoveFavorite(c fiber.Ctx) error {
	squadID := c.Params("squadID")
	placeID := c.Params("placeID")
	userID := middleware.UserIDFrom(c)

	err := h.uc.RemoveFavorite(c.Context(), domain.RemoveFavoriteCommand{
		SquadID: squadID,
		UserID:  userID,
		PlaceID: placeID,
	})
	if err != nil {
		return c.Status(favoriteErrorStatus(err)).JSON(
			response.Fail[any](err.Error()),
		)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ListFavorites handles GET /api/v1/squads/:squadID/favorites.
func (h *FavoriteHandler) ListFavorites(c fiber.Ctx) error {
	squadID := c.Params("squadID")

	favorites, err := h.uc.ListFavorites(c.Context(), squadID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.Fail[dtos.FavoritesResponse](err.Error()),
		)
	}

	return c.Status(fiber.StatusOK).JSON(
		response.OK(dtos.FavoritesToResponse(favorites)),
	)
}

// IsFavorite handles GET /api/v1/squads/:squadID/favorites/:placeID/check.
func (h *FavoriteHandler) IsFavorite(c fiber.Ctx) error {
	squadID := c.Params("squadID")
	placeID := c.Params("placeID")
	userID := middleware.UserIDFrom(c)

	isFav, err := h.uc.IsFavorite(c.Context(), squadID, userID, placeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.Fail[dtos.IsFavoriteResponse](err.Error()),
		)
	}

	return c.Status(fiber.StatusOK).JSON(
		response.OK(dtos.IsFavoriteResponse{IsFavorite: isFav}),
	)
}

// favoriteErrorStatus maps domain sentinel errors to HTTP status codes for favorites.
func favoriteErrorStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrAlreadySaved):
		return fiber.StatusConflict
	case errors.Is(err, domain.ErrNotSaved):
		return fiber.StatusNotFound
	default:
		return fiber.StatusInternalServerError
	}
}
