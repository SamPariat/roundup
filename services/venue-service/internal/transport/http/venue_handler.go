// venue_handler.go contains VenueHandler, which handles HTTP requests for venue
// search and detail endpoints. It parses requests, delegates to VenueUseCase,
// and maps results into the standard Response envelope.
package http

import (
	"errors"

	"github.com/SamPariatIL/roundup/services/venue-service/internal/application"
	"github.com/SamPariatIL/roundup/services/venue-service/internal/domain"
	"github.com/SamPariatIL/roundup/services/venue-service/internal/transport/http/dtos"
	"github.com/gofiber/fiber/v3"
)

// VenueHandler handles HTTP requests for venue search and detail lookups.
type VenueHandler struct {
	uc *application.VenueUseCase
}

// NewVenueHandler constructs a VenueHandler with the given use case.
func NewVenueHandler(uc *application.VenueUseCase) *VenueHandler {
	return &VenueHandler{uc: uc}
}

// Search handles GET /api/v1/venues — returns nearby venues matching the query params.
func (h *VenueHandler) Search(c fiber.Ctx) error {
	var req dtos.SearchRequest
	if err := c.Bind().Query(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			dtos.Fail[dtos.SearchResponse](err.Error()),
		)
	}

	venues, err := h.uc.Search(c.Context(), domain.SearchParams{
		Latitude:       req.Lat,
		Longitude:      req.Lng,
		RadiusInMeters: req.Radius,
		Query:          req.Query,
		Type:           req.Type,
	})
	if err != nil {
		return c.Status(venueErrorStatus(err)).JSON(
			dtos.Fail[dtos.SearchResponse](err.Error()),
		)
	}

	return c.Status(fiber.StatusOK).JSON(
		dtos.OK(dtos.VenuesToSearchResponse(venues)),
	)
}

// GetDetail handles GET /api/v1/venues/:placeID — returns full details for a venue.
func (h *VenueHandler) GetDetail(c fiber.Ctx) error {
	placeID := c.Params("placeID")

	if placeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			dtos.Fail[dtos.VenueDetailResponse]("placeID is required"),
		)
	}

	detail, err := h.uc.GetDetail(c.Context(), placeID)
	if err != nil {
		return c.Status(venueErrorStatus(err)).JSON(
			dtos.Fail[dtos.VenueDetailResponse](err.Error()),
		)
	}

	return c.Status(fiber.StatusOK).JSON(
		dtos.OK(dtos.VenueDetailToResponse(*detail)),
	)
}

// venueErrorStatus maps domain sentinel errors to HTTP status codes.
func venueErrorStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrVenueNotFound):
		return fiber.StatusNotFound
	case errors.Is(err, domain.ErrProviderUnavailable):
		return fiber.StatusServiceUnavailable
	default:
		return fiber.StatusInternalServerError
	}
}
