package http

import (
	"errors"

	"github.com/SamPariat/roundup/services/venue-service/internal/domain"
	"github.com/SamPariat/roundup/shared/pagination"
	"github.com/gofiber/fiber/v3"
)

// errorStatus maps domain sentinel errors to HTTP status codes for every
// venue-service route. One table, so a new sentinel can't be mapped on one
// route and fall through to 500 on another.
func errorStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrInvalidSearch), errors.Is(err, pagination.ErrInvalid):
		return fiber.StatusBadRequest
	case errors.Is(err, domain.ErrSquadNotFound), errors.Is(err, domain.ErrVenueNotFound),
		errors.Is(err, domain.ErrPhotoNotFound), errors.Is(err, domain.ErrNotSaved):
		return fiber.StatusNotFound
	case errors.Is(err, domain.ErrAlreadySaved), errors.Is(err, domain.ErrVisitAlreadyRecorded):
		return fiber.StatusConflict
	case errors.Is(err, domain.ErrProviderUnavailable), errors.Is(err, domain.ErrUserServiceUnavailable):
		return fiber.StatusServiceUnavailable
	default:
		return fiber.StatusInternalServerError
	}
}
