package http

import (
	"errors"

	"github.com/SamPariat/roundup/services/venue-service/internal/application"
	"github.com/SamPariat/roundup/services/venue-service/internal/domain"
	"github.com/SamPariat/roundup/services/venue-service/internal/transport/http/dtos"
	"github.com/SamPariat/roundup/shared/response"
	"github.com/gofiber/fiber/v3"
)

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
		return response.Error(c, fiber.StatusBadRequest, err)
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
		status := fiber.StatusInternalServerError
		if errors.Is(err, domain.ErrVisitAlreadyRecorded) {
			status = fiber.StatusConflict
		}

		return response.Error(c, status, err)
	}

	return c.SendStatus(fiber.StatusCreated)
}

// GetVisitHistory handles GET /api/v1/squads/:squadID/history.
func (h *HistoryHandler) GetVisitHistory(c fiber.Ctx) error {
	squadID := c.Params("squadID")

	summaries, err := h.uc.GetVisitHistory(c.Context(), squadID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err)
	}

	return c.Status(fiber.StatusOK).JSON(
		response.OK(dtos.VisitHistoryToResponse(summaries)),
	)
}
