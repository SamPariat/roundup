package http

import (
	"fmt"

	"github.com/SamPariat/roundup/services/venue-service/internal/application"
	"github.com/SamPariat/roundup/services/venue-service/internal/domain"
	"github.com/SamPariat/roundup/services/venue-service/internal/transport/http/dtos"
	"github.com/SamPariat/roundup/shared/api"
	"github.com/SamPariat/roundup/shared/domainerr"
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
// query params, each with its first photo for the card; ?cursor= continues
// from a previous page's nextCursor.
func (h *VenueHandler) Search(c fiber.Ctx) error {
	var req dtos.SearchRequest
	if err := c.Bind().Query(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	photos, err := photoOptions(c)
	if err != nil {
		return response.Error(c, errorStatus(err), err)
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

	urls, err := h.uc.PhotoURLs(c.Context(), refs, photos.Width)
	if err != nil {
		return response.Error(c, errorStatus(err), err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(dtos.SearchResultToResponse(result, urls)))
}

// GetDetail handles GET /api/v1/venues/:placeID: full details for one venue,
// with as many of its photos as ?photoCount= asks for (5 when omitted), plus
// which of the caller's squads they saved it in.
func (h *VenueHandler) GetDetail(c fiber.Ctx) error {
	photos, err := photoOptions(c)
	if err != nil {
		return response.Error(c, errorStatus(err), err)
	}

	placeID := c.Params("placeID")

	detail, err := h.uc.GetDetail(c.Context(), placeID)
	if err != nil {
		return response.Error(c, errorStatus(err), err)
	}

	savedIn, err := h.favorites.SavedIn(c.Context(), middleware.UserIDFrom(c), placeID)
	if err != nil {
		return response.Error(c, errorStatus(err), err)
	}

	urls, err := h.uc.PhotoURLs(c.Context(), firstPhotos(detail.PhotoRefs, photos.Count), photos.Width)
	if err != nil {
		return response.Error(c, errorStatus(err), err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(dtos.VenueDetailToResponse(*detail, savedIn, urls)))
}

// PhotoURLs handles POST /internal/v1/venues/photo-urls: fresh direct URLs
// for photo references another service stored, since the URLs expire.
func (h *VenueHandler) PhotoURLs(c fiber.Ctx) error {
	var req api.PhotoURLsRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	if len(req.Refs) > api.MaxPhotoURLRefs {
		err := domainerr.New(domain.ErrInvalidPhotoOptions, fmt.Sprintf("at most %d refs per request", api.MaxPhotoURLRefs))

		return response.Error(c, errorStatus(err), err)
	}

	photos := domain.NewPhotoOptions(req.MaxWidth, 0)
	if err := photos.Validate(); err != nil {
		return response.Error(c, errorStatus(err), err)
	}

	refs := make([]domain.PhotoRef, 0, len(req.Refs))
	for _, ref := range req.Refs {
		refs = append(refs, domain.PhotoRef{Reference: ref})
	}

	urls, err := h.uc.PhotoURLs(c.Context(), refs, photos.Width)
	if err != nil {
		return response.Error(c, errorStatus(err), err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(api.PhotoURLsResponse{URLs: urls}))
}

// photoOptions reads ?photoWidth= and ?photoCount=, filling what is left out
// with the defaults. A value the client may not ask for is
// domain.ErrInvalidPhotoOptions, and one that isn't a number a bind error;
// both are 400 through errorStatus.
func photoOptions(c fiber.Ctx) (domain.PhotoOptions, error) {
	var req dtos.PhotoRequest
	if err := c.Bind().Query(&req); err != nil {
		return domain.PhotoOptions{}, domainerr.New(domain.ErrInvalidPhotoOptions, err.Error())
	}

	photos := domain.NewPhotoOptions(req.PhotoWidth, req.PhotoCount)

	return photos, photos.Validate()
}

// firstPhotos returns at most n of refs, from the start.
func firstPhotos(refs []domain.PhotoRef, n int) []domain.PhotoRef {
	return refs[:min(n, len(refs))]
}
