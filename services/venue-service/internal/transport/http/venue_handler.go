package http

import (
	"fmt"

	"github.com/SamPariat/roundup/services/venue-service/internal/application"
	"github.com/SamPariat/roundup/services/venue-service/internal/domain"
	"github.com/SamPariat/roundup/services/venue-service/internal/transport/http/dtos"
	"github.com/SamPariat/roundup/shared/api"
	"github.com/SamPariat/roundup/shared/middleware"
	"github.com/SamPariat/roundup/shared/response"
	"github.com/gofiber/fiber/v3"
)

// Photo widths in pixels, one per kind of screen, and how many photos a
// venue's page shows. Google bills each photo looked up, so a response only
// carries the photos its screen draws.
const (
	// cardPhotoWidth fits a search card, 400 points at 2x.
	cardPhotoWidth = 800
	// detailPhotoWidth fits the largest tile of a venue's page, 570 points at 2x.
	detailPhotoWidth = 1200
	// thumbPhotoWidth fits a favorite or history row's thumbnail, 56 points at 3x.
	thumbPhotoWidth = 200
	// detailPhotoCount is the large photo and the four beside it.
	detailPhotoCount = 5
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
// query params, each with its first photo for the card; ?cursor= continues
// from a previous page's nextCursor.
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

	refs := make([]domain.PhotoRef, 0, len(result.Venues))
	for _, v := range result.Venues {
		refs = append(refs, firstPhotos(v.PhotoRefs, 1)...)
	}

	urls, err := h.uc.PhotoURLs(c.Context(), refs, cardPhotoWidth)
	if err != nil {
		return response.Error(c, errorStatus(err), err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(dtos.SearchResultToResponse(result, urls)))
}

// GetDetail handles GET /api/v1/venues/:placeID: full details for one venue,
// with its first detailPhotoCount photos, plus which of the caller's squads
// they saved it in.
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

	urls, err := h.uc.PhotoURLs(c.Context(), firstPhotos(detail.PhotoRefs, detailPhotoCount), detailPhotoWidth)
	if err != nil {
		return response.Error(c, errorStatus(err), err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(dtos.VenueDetailToResponse(*detail, savedIn, urls)))
}

// PhotoURLs handles POST api.VenuePhotoURLsPath: fresh direct URLs
// for photo references another service stored, since the URLs expire.
func (h *VenueHandler) PhotoURLs(c fiber.Ctx) error {
	var req api.PhotoURLsRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	if len(req.Refs) > api.MaxPhotoURLRefs {
		return response.Error(c, fiber.StatusBadRequest, fmt.Errorf("at most %d refs per request", api.MaxPhotoURLRefs))
	}

	if req.MaxWidth < 1 || req.MaxWidth > api.MaxPhotoURLWidth {
		return response.Error(c, fiber.StatusBadRequest, fmt.Errorf("maxWidth must be between 1 and %d", api.MaxPhotoURLWidth))
	}

	refs := make([]domain.PhotoRef, 0, len(req.Refs))
	for _, ref := range req.Refs {
		refs = append(refs, domain.PhotoRef{Reference: ref})
	}

	urls, err := h.uc.PhotoURLs(c.Context(), refs, req.MaxWidth)
	if err != nil {
		return response.Error(c, errorStatus(err), err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(api.PhotoURLsResponse{URLs: urls}))
}

// firstPhotos returns at most n of refs, from the start.
func firstPhotos(refs []domain.PhotoRef, n int) []domain.PhotoRef {
	return refs[:min(n, len(refs))]
}
