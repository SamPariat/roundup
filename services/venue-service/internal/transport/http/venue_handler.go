package http

import (
	"fmt"
	"time"

	"github.com/SamPariat/roundup/services/venue-service/internal/application"
	"github.com/SamPariat/roundup/services/venue-service/internal/domain"
	"github.com/SamPariat/roundup/services/venue-service/internal/transport/http/dtos"
	"github.com/SamPariat/roundup/shared/middleware"
	"github.com/SamPariat/roundup/shared/response"
	"github.com/gofiber/fiber/v3"
)

// VenueHandler handles HTTP requests for venue search and detail lookups.
// favorites answers which of the caller's squads saved a venue.
type VenueHandler struct {
	uc        *application.VenueUseCase
	favorites *application.FavoriteUseCase
}

// NewVenueHandler constructs a VenueHandler with the given use cases.
func NewVenueHandler(uc *application.VenueUseCase, favorites *application.FavoriteUseCase) *VenueHandler {
	return &VenueHandler{uc: uc, favorites: favorites}
}

// Search handles GET /api/v1/venues: one page of nearby venues matching the
// query params; ?cursor= continues from a previous page's nextCursor.
func (h *VenueHandler) Search(c fiber.Ctx) error {
	var req dtos.SearchRequest
	if err := c.Bind().Query(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	result, err := h.uc.Search(c.Context(), domain.SearchParams{
		Latitude:       req.Lat,
		Longitude:      req.Lng,
		RadiusInMeters: req.Radius,
		Query:          req.Query,
		Type:           req.Type,
		PageToken:      req.Cursor,
	})
	if err != nil {
		return response.Error(c, errorStatus(err), err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(dtos.SearchResultToResponse(result)))
}

// GetDetail handles GET /api/v1/venues/:placeID: full details for one venue,
// plus which of the caller's squads they saved it in.
func (h *VenueHandler) GetDetail(c fiber.Ctx) error {
	placeID := c.Params("placeID")

	detail, err := h.uc.GetDetail(c.Context(), placeID)
	if err != nil {
		return response.Error(c, errorStatus(err), err)
	}

	savedIn, err := h.favorites.SavedIn(c.Context(), middleware.UserIDFrom(c), placeID)
	if err != nil {
		return response.Error(c, errorStatus(err), err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(dtos.VenueDetailToResponse(*detail, savedIn)))
}

// photoMaxAge is how long clients and CDNs may cache a venue photo.
const photoMaxAge = 24 * time.Hour

// Photo handles GET /api/v1/venues/photos/:ref: the photo's bytes, so clients
// never need the Google API key. Public at the gateway, because image tags
// cannot send a bearer token.
func (h *VenueHandler) Photo(c fiber.Ctx) error {
	var req dtos.PhotoRequest
	if err := c.Bind().Query(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	photo, err := h.uc.Photo(c.Context(), c.Params("ref"), req.MaxWidth)
	if err != nil {
		return response.Error(c, errorStatus(err), err)
	}

	c.Set(fiber.HeaderContentType, photo.ContentType)
	c.Set(fiber.HeaderCacheControl, fmt.Sprintf("public, max-age=%d", int(photoMaxAge.Seconds())))

	return c.SendStream(photo.Body)
}
