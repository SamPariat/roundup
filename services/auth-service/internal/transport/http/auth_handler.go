package http

import (
	"errors"

	"github.com/SamPariat/roundup/services/auth-service/internal/application"
	"github.com/SamPariat/roundup/services/auth-service/internal/domain"
	"github.com/SamPariat/roundup/services/auth-service/internal/transport/http/dtos"
	"github.com/SamPariat/roundup/shared/response"
	"github.com/gofiber/fiber/v3"
)

// AuthHandler handles HTTP requests for login, refresh, and logout. All of its
// routes are public: they are how a client gets a token in the first place.
type AuthHandler struct {
	otp    *application.OTPUseCase
	tokens *application.TokenUseCase
}

// NewAuthHandler constructs an AuthHandler with the given use cases.
func NewAuthHandler(otp *application.OTPUseCase, tokens *application.TokenUseCase) *AuthHandler {
	return &AuthHandler{otp: otp, tokens: tokens}
}

// RequestOTP handles POST /api/v1/auth/otp/request. Returns 202: the code is
// on its way, and the response says nothing more.
func (h *AuthHandler) RequestOTP(c fiber.Ctx) error {
	var req dtos.RequestOTPRequest
	if err := c.Bind().Body(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}

	if err := h.otp.Request(c.Context(), req.Phone); err != nil {
		return fail(c, authErrorStatus(err), err)
	}

	return c.SendStatus(fiber.StatusAccepted)
}

// VerifyOTP handles POST /api/v1/auth/otp/verify.
func (h *AuthHandler) VerifyOTP(c fiber.Ctx) error {
	var req dtos.VerifyOTPRequest
	if err := c.Bind().Body(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}

	result, err := h.otp.Verify(c.Context(), req.Phone, req.Code)
	if err != nil {
		return fail(c, authErrorStatus(err), err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(dtos.LoginToResponse(result)))
}

// Refresh handles POST /api/v1/auth/refresh.
func (h *AuthHandler) Refresh(c fiber.Ctx) error {
	var req dtos.RefreshTokenRequest
	if err := c.Bind().Body(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}

	pair, err := h.tokens.Refresh(c.Context(), req.RefreshToken)
	if err != nil {
		return fail(c, authErrorStatus(err), err)
	}

	return c.Status(fiber.StatusOK).JSON(response.OK(dtos.TokenPairToResponse(pair)))
}

// Logout handles POST /api/v1/auth/logout.
func (h *AuthHandler) Logout(c fiber.Ctx) error {
	var req dtos.RefreshTokenRequest
	if err := c.Bind().Body(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, err)
	}

	if err := h.tokens.Logout(c.Context(), req.RefreshToken); err != nil {
		return fail(c, authErrorStatus(err), err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func fail(c fiber.Ctx, status int, err error) error {
	return c.Status(status).JSON(response.Fail[any](err.Error()))
}

// authErrorStatus maps domain sentinel errors to HTTP status codes.
func authErrorStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrInvalidPhone):
		return fiber.StatusBadRequest
	case errors.Is(err, domain.ErrInvalidCode), errors.Is(err, domain.ErrInvalidRefreshToken):
		return fiber.StatusUnauthorized
	case errors.Is(err, domain.ErrTooManyRequests), errors.Is(err, domain.ErrTooManyAttempts):
		return fiber.StatusTooManyRequests
	case errors.Is(err, domain.ErrUserServiceUnavailable):
		return fiber.StatusServiceUnavailable
	default:
		return fiber.StatusInternalServerError
	}
}
