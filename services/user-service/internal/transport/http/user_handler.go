package http

import (
	"github.com/SamPariat/roundup/services/user-service/internal/application"
	"github.com/SamPariat/roundup/services/user-service/internal/domain"
	"github.com/SamPariat/roundup/services/user-service/internal/transport/http/dtos"
	"github.com/SamPariat/roundup/shared/middleware"
	"github.com/SamPariat/roundup/shared/response"
	"github.com/gofiber/fiber/v3"
)

// UserHandler handles HTTP requests for the user endpoints. On /me routes the
// caller's user ID comes from the X-User-ID header via middleware.UserIDFrom.
type UserHandler struct {
	uc *application.UserUseCase
}

// NewUserHandler constructs a UserHandler with the given use case.
func NewUserHandler(uc *application.UserUseCase) *UserHandler {
	return &UserHandler{uc: uc}
}

// Create handles POST /api/v1/users. Called by auth-service after a phone is
// verified. Returns 201 for a new user and 200 if the user already existed.
func (h *UserHandler) Create(c fiber.Ctx) error {
	var req dtos.CreateUserRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	user, created, err := h.uc.CreateOrGet(c.Context(), req.Phone)
	if err != nil {
		return response.Error(c, errorStatus(err), err)
	}

	status := fiber.StatusOK
	if created {
		status = fiber.StatusCreated
	}

	return c.Status(status).JSON(response.OK(dtos.UserToResponse(user)))
}

// GetMe handles GET /api/v1/users/me.
func (h *UserHandler) GetMe(c fiber.Ctx) error {
	user, err := h.uc.Get(c.Context(), middleware.UserIDFrom(c))
	if err != nil {
		return response.Error(c, errorStatus(err), err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(dtos.UserToResponse(user)))
}

// UpdateMe handles PATCH /api/v1/users/me.
func (h *UserHandler) UpdateMe(c fiber.Ctx) error {
	var req dtos.UpdateProfileRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	user, err := h.uc.UpdateProfile(c.Context(), domain.UpdateProfileCommand{
		UserID:    middleware.UserIDFrom(c),
		Name:      req.Name,
		AvatarURL: req.AvatarURL,
	})
	if err != nil {
		return response.Error(c, errorStatus(err), err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(dtos.UserToResponse(user)))
}

// DeleteMe handles DELETE /api/v1/users/me.
func (h *UserHandler) DeleteMe(c fiber.Ctx) error {
	if err := h.uc.Delete(c.Context(), middleware.UserIDFrom(c)); err != nil {
		return response.Error(c, errorStatus(err), err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// Lookup handles POST /api/v1/users/lookup: which of the caller's contacts
// are on Roundup. Phones travel in the body so they stay out of URLs and logs.
func (h *UserHandler) Lookup(c fiber.Ctx) error {
	var req dtos.LookupUsersRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err)
	}

	users, err := h.uc.Lookup(c.Context(), middleware.UserIDFrom(c), req.Phones)
	if err != nil {
		return response.Error(c, errorStatus(err), err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(dtos.UsersToLookupResponse(users)))
}
