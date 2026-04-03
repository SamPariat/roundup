package http

import (
	"errors"

	"github.com/SamPariatIL/roundup/services/venue-service/internal/application"
	"github.com/SamPariatIL/roundup/services/venue-service/internal/domain"
	"github.com/SamPariatIL/roundup/services/venue-service/internal/transport/http/dtos"
	"github.com/gofiber/fiber/v3"
)

// favorite_handler.go contains FavoriteHandler, which handles HTTP requests for
// the favorites feature. Squad ID is always taken from the path parameter.

// FavoriteHandler handles HTTP requests for the favorite endpoints.
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
			dtos.Fail[any](err.Error()),
		)
	}

	err := h.uc.AddFavorite(c.Context(), domain.AddFavoriteCommand{
		SquadID: squadID,
		UserID:  req.UserID,
		PlaceID: req.PlaceID,
		Name:    req.Name,
	})
	if err != nil {
		return c.Status(favoriteErrorStatus(err)).JSON(
			dtos.Fail[any](err.Error()),
		)
	}

	return c.SendStatus(fiber.StatusCreated)
}

// RemoveFavorite handles DELETE /api/v1/squads/:squadID/favorites/:placeID.
func (h *FavoriteHandler) RemoveFavorite(c fiber.Ctx) error {
	squadID := c.Params("squadID")
	placeID := c.Params("placeID")
	userID := c.Query("userID")

	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			dtos.Fail[any]("userID query param is required"),
		)
	}

	err := h.uc.RemoveFavorite(c.Context(), domain.RemoveFavoriteCommand{
		SquadID: squadID,
		UserID:  userID,
		PlaceID: placeID,
	})
	if err != nil {
		return c.Status(favoriteErrorStatus(err)).JSON(
			dtos.Fail[any](err.Error()),
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
			dtos.Fail[dtos.FavoritesResponse](err.Error()),
		)
	}

	return c.Status(fiber.StatusOK).JSON(
		dtos.OK(dtos.FavoritesToResponse(favorites)),
	)
}

// IsFavorite handles GET /api/v1/squads/:squadID/favorites/:placeID/check.
func (h *FavoriteHandler) IsFavorite(c fiber.Ctx) error {
	squadID := c.Params("squadID")
	placeID := c.Params("placeID")
	userID := c.Query("userID")

	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			dtos.Fail[dtos.IsFavoriteResponse]("userID query param is required"),
		)
	}

	isFav, err := h.uc.IsFavorite(c.Context(), squadID, userID, placeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			dtos.Fail[dtos.IsFavoriteResponse](err.Error()),
		)
	}

	return c.Status(fiber.StatusOK).JSON(
		dtos.OK(dtos.IsFavoriteResponse{IsFavorite: isFav}),
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
