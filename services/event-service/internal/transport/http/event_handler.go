package http

import (
	"context"

	"github.com/SamPariat/roundup/services/event-service/internal/application"
	"github.com/SamPariat/roundup/services/event-service/internal/domain"
	"github.com/SamPariat/roundup/services/event-service/internal/transport/http/dtos"
	"github.com/SamPariat/roundup/shared/middleware"
	"github.com/SamPariat/roundup/shared/pagination"
	"github.com/SamPariat/roundup/shared/response"
	"github.com/gofiber/fiber/v3"
)

// EventHandler handles HTTP requests for events. The caller's user ID comes
// from the X-User-ID header via middleware.UserIDFrom. Every route that
// changes an event answers with the updated event detail, so the app redraws
// from one response.
type EventHandler struct {
	uc *application.EventUseCase
}

// NewEventHandler constructs an EventHandler with the given use case.
func NewEventHandler(uc *application.EventUseCase) *EventHandler {
	return &EventHandler{uc: uc}
}

// Create handles POST /api/v1/squads/:squadID/events.
func (h *EventHandler) Create(c fiber.Ctx) error {
	var req dtos.CreateEventRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	event, err := h.uc.Create(c.Context(), application.CreateEventCommand{
		SquadID:        c.Params("squadID"),
		UserID:         middleware.UserIDFrom(c),
		Name:           req.Name,
		Type:           domain.Type(req.Type),
		StartsAt:       req.StartsAt,
		VotingDeadline: req.VotingDeadline,
	})
	if err != nil {
		return response.Error(c, errorStatus(err), err)
	}

	return c.Status(fiber.StatusCreated).JSON(response.OK(dtos.EventToResponse(*event)))
}

// List handles GET /api/v1/squads/:squadID/events?limit=&cursor=.
func (h *EventHandler) List(c fiber.Ctx) error {
	page, err := pagination.FromQuery(c.Query("limit"), c.Query("cursor"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	events, err := h.uc.List(c.Context(), c.Params("squadID"), middleware.UserIDFrom(c), page)
	if err != nil {
		return response.Error(c, errorStatus(err), err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(dtos.SquadEventsToResponse(events)))
}

// Get handles GET /api/v1/events/:eventID.
func (h *EventHandler) Get(c fiber.Ctx) error {
	return h.detail(c, func(ctx context.Context, eventID, userID string) (*domain.EventDetail, error) {
		return h.uc.Get(ctx, eventID, userID)
	})
}

// Update handles PATCH /api/v1/events/:eventID.
func (h *EventHandler) Update(c fiber.Ctx) error {
	var req dtos.UpdateEventRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	changes := domain.EventChanges{Name: req.Name, StartsAt: req.StartsAt, VotingDeadline: req.VotingDeadline}
	if req.Type != nil {
		t := domain.Type(*req.Type)
		changes.Type = &t
	}

	return h.detail(c, func(ctx context.Context, eventID, userID string) (*domain.EventDetail, error) {
		return h.uc.Update(ctx, eventID, userID, changes)
	})
}

// SetRSVP handles PUT /api/v1/events/:eventID/rsvp.
func (h *EventHandler) SetRSVP(c fiber.Ctx) error {
	var req dtos.RSVPRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	return h.detail(c, func(ctx context.Context, eventID, userID string) (*domain.EventDetail, error) {
		return h.uc.SetRSVP(ctx, eventID, userID, domain.RSVPStatus(req.Status))
	})
}

// Nominate handles POST /api/v1/events/:eventID/nominations.
func (h *EventHandler) Nominate(c fiber.Ctx) error {
	var req dtos.PlaceRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	return h.detail(c, func(ctx context.Context, eventID, userID string) (*domain.EventDetail, error) {
		return h.uc.Nominate(ctx, eventID, userID, req.PlaceID)
	})
}

// RemoveNomination handles DELETE /api/v1/events/:eventID/nominations/:placeID.
func (h *EventHandler) RemoveNomination(c fiber.Ctx) error {
	placeID := c.Params("placeID")

	return h.detail(c, func(ctx context.Context, eventID, userID string) (*domain.EventDetail, error) {
		return h.uc.RemoveNomination(ctx, eventID, userID, placeID)
	})
}

// Vote handles PUT /api/v1/events/:eventID/vote.
func (h *EventHandler) Vote(c fiber.Ctx) error {
	var req dtos.PlaceRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	return h.detail(c, func(ctx context.Context, eventID, userID string) (*domain.EventDetail, error) {
		return h.uc.Vote(ctx, eventID, userID, req.PlaceID)
	})
}

// StartVoting handles POST /api/v1/events/:eventID/voting/start.
func (h *EventHandler) StartVoting(c fiber.Ctx) error {
	return h.detail(c, h.uc.StartVoting)
}

// CloseVoting handles POST /api/v1/events/:eventID/voting/close.
func (h *EventHandler) CloseVoting(c fiber.Ctx) error {
	return h.detail(c, h.uc.CloseVoting)
}

// StartNight handles POST /api/v1/events/:eventID/start.
func (h *EventHandler) StartNight(c fiber.Ctx) error {
	return h.detail(c, h.uc.StartNight)
}

// Close handles POST /api/v1/events/:eventID/close.
func (h *EventHandler) Close(c fiber.Ctx) error {
	return h.detail(c, h.uc.Close)
}

// ConfirmDue handles POST /internal/v1/events/confirm-due.
func (h *EventHandler) ConfirmDue(c fiber.Ctx) error {
	return h.due(c, h.uc.ConfirmDue)
}

// CloseDue handles POST /internal/v1/events/close-due.
func (h *EventHandler) CloseDue(c fiber.Ctx) error {
	return h.due(c, h.uc.CloseDue)
}

// detail runs fn for the path's event and the caller, and answers with the
// event detail.
func (h *EventHandler) detail(c fiber.Ctx, fn func(ctx context.Context, eventID, userID string) (*domain.EventDetail, error)) error {
	d, err := fn(c.Context(), c.Params("eventID"), middleware.UserIDFrom(c))
	if err != nil {
		return response.Error(c, errorStatus(err), err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(dtos.EventDetailToResponse(d)))
}

// due runs a sweep and answers with how many events it moved on.
func (h *EventHandler) due(c fiber.Ctx, sweep func(ctx context.Context) (int, error)) error {
	n, err := sweep(c.Context())
	if err != nil {
		return response.Error(c, errorStatus(err), err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(dtos.DueResponse{Count: n}))
}
