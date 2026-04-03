package http

import (
	"github.com/SamPariatIL/roundup/services/venue-service/internal/application"
	"github.com/SamPariatIL/roundup/services/venue-service/internal/domain"
	"github.com/SamPariatIL/roundup/services/venue-service/internal/transport/http/dtos"
	"github.com/gofiber/fiber/v3"
)

// history_handler.go contains HistoryHandler, which handles HTTP requests for
// recording venue visits and fetching visit history for a squad.

// HistoryHandler handles HTTP requests for the visit history endpoints.
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
		return c.Status(fiber.StatusBadRequest).JSON(
			dtos.Fail[any](err.Error()),
		)
	}

	err := h.uc.RecordVisit(c.Context(), domain.RecordVisitCommand{
		SquadID:         squadID,
		EventID:         req.EventID,
		PlaceID:         req.PlaceID,
		Name:            req.Name,
		VisitedAt:       req.VisitedAt,
		AvgSpendInPaise: req.AvgSpendInPaise,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			dtos.Fail[any](err.Error()),
		)
	}

	return c.SendStatus(fiber.StatusCreated)
}

// GetVisitHistory handles GET /api/v1/squads/:squadID/history.
func (h *HistoryHandler) GetVisitHistory(c fiber.Ctx) error {
	squadID := c.Params("squadID")

	summaries, err := h.uc.GetVisitHistory(c.Context(), squadID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			dtos.Fail[dtos.VisitHistoryResponse](err.Error()),
		)
	}

	return c.Status(fiber.StatusOK).JSON(
		dtos.OK(dtos.VisitHistoryToResponse(summaries)),
	)
}
