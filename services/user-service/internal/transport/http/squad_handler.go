package http

import (
	"errors"

	"github.com/SamPariat/roundup/services/user-service/internal/application"
	"github.com/SamPariat/roundup/services/user-service/internal/domain"
	"github.com/SamPariat/roundup/services/user-service/internal/transport/http/dtos"
	"github.com/SamPariat/roundup/shared/middleware"
	"github.com/SamPariat/roundup/shared/pagination"
	"github.com/SamPariat/roundup/shared/response"
	"github.com/gofiber/fiber/v3"
)

// SquadHandler handles HTTP requests for the squad endpoints. The caller's
// user ID comes from the X-User-ID header via middleware.UserIDFrom.
type SquadHandler struct {
	uc *application.SquadUseCase
}

// NewSquadHandler constructs a SquadHandler with the given use case.
func NewSquadHandler(uc *application.SquadUseCase) *SquadHandler {
	return &SquadHandler{uc: uc}
}

// Create handles POST /api/v1/squads. The caller becomes the squad's admin.
func (h *SquadHandler) Create(c fiber.Ctx) error {
	var req dtos.CreateSquadRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	squad, err := h.uc.Create(c.Context(), domain.CreateSquadCommand{
		CreatorID: middleware.UserIDFrom(c),
		Name:      req.Name,
		AvatarURL: req.AvatarURL,
	})
	if err != nil {
		return response.Error(c, squadErrorStatus(err), err)
	}

	return c.Status(fiber.StatusCreated).JSON(response.OK(dtos.SquadToResponse(squad)))
}

// List handles GET /api/v1/squads?limit=&cursor=: one page of the caller's squads.
func (h *SquadHandler) List(c fiber.Ctx) error {
	page, err := pagination.FromQuery(c.Query("limit"), c.Query("cursor"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	squads, next, err := h.uc.List(c.Context(), middleware.UserIDFrom(c), page)
	if err != nil {
		return response.Error(c, squadErrorStatus(err), err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(dtos.SquadsToResponse(squads, next)))
}

// Get handles GET /api/v1/squads/:squadID. Non-members get 404. venue-service
// also calls this to check membership.
func (h *SquadHandler) Get(c fiber.Ctx) error {
	squad, err := h.uc.Get(c.Context(), c.Params("squadID"), middleware.UserIDFrom(c))
	if err != nil {
		return response.Error(c, squadErrorStatus(err), err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(dtos.SquadDetailToResponse(squad)))
}

// Update handles PATCH /api/v1/squads/:squadID: admins rename the squad or
// change its picture.
func (h *SquadHandler) Update(c fiber.Ctx) error {
	var req dtos.UpdateSquadRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	squad, err := h.uc.Update(c.Context(), domain.UpdateSquadCommand{
		SquadID:   c.Params("squadID"),
		ActorID:   middleware.UserIDFrom(c),
		Name:      req.Name,
		AvatarURL: req.AvatarURL,
	})
	if err != nil {
		return response.Error(c, squadErrorStatus(err), err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(dtos.SquadToResponse(squad)))
}

// AddMember handles POST /api/v1/squads/:squadID/members.
func (h *SquadHandler) AddMember(c fiber.Ctx) error {
	var req dtos.AddMemberRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	member, err := h.uc.AddMember(c.Context(), domain.AddMemberCommand{
		SquadID: c.Params("squadID"),
		ActorID: middleware.UserIDFrom(c),
		Phone:   req.Phone,
	})
	if err != nil {
		return response.Error(c, squadErrorStatus(err), err)
	}

	return c.Status(fiber.StatusCreated).JSON(response.OK(dtos.SquadMemberToResponse(member)))
}

// ChangeRole handles PATCH /api/v1/squads/:squadID/members/:userID.
func (h *SquadHandler) ChangeRole(c fiber.Ctx) error {
	var req dtos.ChangeRoleRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	err := h.uc.ChangeRole(c.Context(), domain.ChangeRoleCommand{
		SquadID: c.Params("squadID"),
		ActorID: middleware.UserIDFrom(c),
		UserID:  c.Params("userID"),
		Role:    domain.Role(req.Role),
	})
	if err != nil {
		return response.Error(c, squadErrorStatus(err), err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// RemoveMember handles DELETE /api/v1/squads/:squadID/members/:userID.
// Deleting your own user ID leaves the squad.
func (h *SquadHandler) RemoveMember(c fiber.Ctx) error {
	_, err := h.uc.RemoveMember(c.Context(), domain.RemoveMemberCommand{
		SquadID: c.Params("squadID"),
		ActorID: middleware.UserIDFrom(c),
		UserID:  c.Params("userID"),
	})
	if err != nil {
		return response.Error(c, squadErrorStatus(err), err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// squadErrorStatus maps domain sentinel errors to HTTP status codes for squads.
func squadErrorStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrInvalidSquadName), errors.Is(err, domain.ErrInvalidPhone),
		errors.Is(err, domain.ErrInvalidRole), errors.Is(err, pagination.ErrInvalid):
		return fiber.StatusBadRequest
	case errors.Is(err, domain.ErrForbidden):
		return fiber.StatusForbidden
	case errors.Is(err, domain.ErrSquadNotFound), errors.Is(err, domain.ErrMemberNotFound),
		errors.Is(err, domain.ErrUserNotFound):
		return fiber.StatusNotFound
	case errors.Is(err, domain.ErrAlreadyMember), errors.Is(err, domain.ErrSquadFull),
		errors.Is(err, domain.ErrLastAdmin):
		return fiber.StatusConflict
	default:
		return fiber.StatusInternalServerError
	}
}
