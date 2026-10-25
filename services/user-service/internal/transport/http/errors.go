package http

import (
	"errors"

	"github.com/SamPariat/roundup/services/user-service/internal/domain"
	"github.com/SamPariat/roundup/shared/pagination"
	"github.com/gofiber/fiber/v3"
)

// errorStatus maps domain sentinel errors to HTTP status codes for every
// user-service route. One table, so a new sentinel can't be mapped on one
// route and fall through to 500 on another.
func errorStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrInvalidPhone), errors.Is(err, domain.ErrInvalidName),
		errors.Is(err, domain.ErrTooManyPhones), errors.Is(err, domain.ErrInvalidToken),
		errors.Is(err, domain.ErrInvalidPlatform), errors.Is(err, domain.ErrInvalidSquadName),
		errors.Is(err, domain.ErrInvalidRole), errors.Is(err, pagination.ErrInvalid):
		return fiber.StatusBadRequest
	case errors.Is(err, domain.ErrForbidden):
		return fiber.StatusForbidden
	case errors.Is(err, domain.ErrUserNotFound), errors.Is(err, domain.ErrDeviceTokenNotFound),
		errors.Is(err, domain.ErrSquadNotFound), errors.Is(err, domain.ErrMemberNotFound):
		return fiber.StatusNotFound
	case errors.Is(err, domain.ErrPhoneTaken), errors.Is(err, domain.ErrAlreadyMember),
		errors.Is(err, domain.ErrSquadFull), errors.Is(err, domain.ErrLastAdmin):
		return fiber.StatusConflict
	default:
		return fiber.StatusInternalServerError
	}
}
