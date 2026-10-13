package http

import (
	"errors"
	"net/url"

	"github.com/SamPariat/roundup/services/user-service/internal/application"
	"github.com/SamPariat/roundup/services/user-service/internal/domain"
	"github.com/SamPariat/roundup/services/user-service/internal/transport/http/dtos"
	"github.com/SamPariat/roundup/shared/middleware"
	"github.com/SamPariat/roundup/shared/response"
	"github.com/gofiber/fiber/v3"
)

// device_token_handler.go contains DeviceTokenHandler, which handles HTTP
// requests for registering and removing FCM push tokens.

// DeviceTokenHandler handles HTTP requests for the device token endpoints.
type DeviceTokenHandler struct {
	uc *application.DeviceTokenUseCase
}

// NewDeviceTokenHandler constructs a DeviceTokenHandler with the given use case.
func NewDeviceTokenHandler(uc *application.DeviceTokenUseCase) *DeviceTokenHandler {
	return &DeviceTokenHandler{uc: uc}
}

// Register handles POST /api/v1/users/me/device-tokens.
func (h *DeviceTokenHandler) Register(c fiber.Ctx) error {
	var req dtos.RegisterDeviceTokenRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.Fail[any](err.Error()),
		)
	}

	token, err := h.uc.Register(c.Context(), domain.RegisterDeviceTokenCommand{
		UserID:   middleware.UserIDFrom(c),
		Token:    req.Token,
		Platform: domain.Platform(req.Platform),
	})
	if err != nil {
		return c.Status(deviceTokenErrorStatus(err)).JSON(
			response.Fail[any](err.Error()),
		)
	}

	return c.Status(fiber.StatusCreated).JSON(response.OK(dtos.DeviceTokenToResponse(token)))
}

// Remove handles DELETE /api/v1/users/me/device-tokens/:token.
// FCM tokens can contain characters like ':', so clients should URL-encode them.
func (h *DeviceTokenHandler) Remove(c fiber.Ctx) error {
	token, err := url.PathUnescape(c.Params("token"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.Fail[any]("token path param is not valid URL encoding"),
		)
	}

	err = h.uc.Remove(c.Context(), domain.RemoveDeviceTokenCommand{
		UserID: middleware.UserIDFrom(c),
		Token:  token,
	})
	if err != nil {
		return c.Status(deviceTokenErrorStatus(err)).JSON(
			response.Fail[any](err.Error()),
		)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// deviceTokenErrorStatus maps domain sentinel errors to HTTP status codes for device tokens.
func deviceTokenErrorStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrInvalidToken), errors.Is(err, domain.ErrInvalidPlatform):
		return fiber.StatusBadRequest
	case errors.Is(err, domain.ErrDeviceTokenNotFound), errors.Is(err, domain.ErrUserNotFound):
		return fiber.StatusNotFound
	default:
		return fiber.StatusInternalServerError
	}
}
