package http

import (
	"github.com/SamPariat/roundup/services/venue-service/internal/application"
	"github.com/SamPariat/roundup/services/venue-service/internal/domain"
	"github.com/SamPariat/roundup/services/venue-service/internal/transport/http/dtos"
	"github.com/SamPariat/roundup/shared/middleware"
	"github.com/SamPariat/roundup/shared/pagination"
	"github.com/SamPariat/roundup/shared/response"
	"github.com/gofiber/fiber/v3"
)

// HistoryHandler handles HTTP requests for the visit history endpoints. The
// squad ID comes from the path and the caller's user ID from the X-User-ID
// header via middleware.UserIDFrom.
type HistoryHandler struct {
	uc     *application.HistoryUseCase
	venues *application.VenueUseCase
}

// NewHistoryHandler constructs a HistoryHandler; venues looks up the photos' URLs.
func NewHistoryHandler(uc *application.HistoryUseCase, venues *application.VenueUseCase) *HistoryHandler {
	return &HistoryHandler{uc: uc, venues: venues}
}

// Record handles POST /api/v1/squads/:squadID/history.
func (h *HistoryHandler) Record(c fiber.Ctx) error {
	var req dtos.RecordVisitRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	err := h.uc.Record(c.Context(), domain.RecordVisitCommand{
		SquadID:         c.Params("squadID"),
		UserID:          middleware.UserIDFrom(c),
		EventID:         req.EventID,
		PlaceID:         req.PlaceID,
		VisitedAt:       req.VisitedAt,
		AvgSpendInPaise: req.AvgSpendInPaise,
	})
	if err != nil {
		return response.Error(c, errorStatus(err), err)
	}

	return c.SendStatus(fiber.StatusCreated)
}

// List handles GET /api/v1/squads/:squadID/history?limit=&cursor=.
func (h *HistoryHandler) List(c fiber.Ctx) error {
	page, err := pagination.FromQuery(c.Query("limit"), c.Query("cursor"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	photos, err := photoOptions(c)
	if err != nil {
		return response.Error(c, errorStatus(err), err)
	}

	summaries, next, err := h.uc.List(c.Context(), c.Params("squadID"), middleware.UserIDFrom(c), page)
	if err != nil {
		return response.Error(c, errorStatus(err), err)
	}

	refs := make([]domain.PhotoRef, 0, len(summaries))
	for _, item := range summaries {
		refs = append(refs, item.Photo)
	}

	urls, err := h.venues.PhotoURLs(c.Context(), refs, photos.Width)
	if err != nil {
		return response.Error(c, errorStatus(err), err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(dtos.VisitHistoryToResponse(summaries, next, urls)))
}
