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

// HistoryHandler handles HTTP requests for the visit history endpoints. The
// squad ID comes from the path and the caller's user ID from the X-User-ID
// header via middleware.UserIDFrom.
type HistoryHandler struct {
	uc *application.HistoryUseCase
}

// NewHistoryHandler constructs a HistoryHandler with the given use case.
func NewHistoryHandler(uc *application.HistoryUseCase) *HistoryHandler {
	return &HistoryHandler{uc: uc}
}

// RecordVisit handles POST /api/v1/squads/:squadID/history.
func (h *HistoryHandler) RecordVisit(c fiber.Ctx) error {
	squadID := c.Params("squadID")

	var req dtos.RecordVisitRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	err := h.uc.RecordVisit(c.Context(), domain.RecordVisitCommand{
		SquadID:         squadID,
		UserID:          middleware.UserIDFrom(c),
		EventID:         req.EventID,
		PlaceID:         req.PlaceID,
		Name:            req.Name,
		VisitedAt:       req.VisitedAt,
		AvgSpendInPaise: req.AvgSpendInPaise,
	})
	if err != nil {
		return response.Error(c, historyErrorStatus(err), err)
	}

	return c.SendStatus(fiber.StatusCreated)
}

// GetVisitHistory handles GET /api/v1/squads/:squadID/history.
func (h *HistoryHandler) GetVisitHistory(c fiber.Ctx) error {
	squadID := c.Params("squadID")

	summaries, err := h.uc.GetVisitHistory(c.Context(), squadID, middleware.UserIDFrom(c))
	if err != nil {
		return response.Error(c, historyErrorStatus(err), err)
	}

	return c.Status(fiber.StatusOK).JSON(
		response.OK(dtos.VisitHistoryToResponse(summaries)),
	)
}

// historyErrorStatus maps domain sentinel errors to HTTP status codes for visit history.
func historyErrorStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrVisitAlreadyRecorded):
		return fiber.StatusConflict
	case errors.Is(err, domain.ErrSquadNotFound):
		return fiber.StatusNotFound
	case errors.Is(err, domain.ErrUserServiceUnavailable):
		return fiber.StatusServiceUnavailable
	default:
		return fiber.StatusInternalServerError
	}
}
