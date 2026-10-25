package http

import (
	"errors"

	"github.com/SamPariat/roundup/services/event-service/internal/domain"
	"github.com/SamPariat/roundup/shared/pagination"
	"github.com/gofiber/fiber/v3"
)

// errorStatus maps domain sentinel errors to HTTP status codes for every
// event-service route. One table, so a new sentinel can't be mapped on one
// route and fall through to 500 on another.
func errorStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrInvalidEvent), errors.Is(err, domain.ErrInvalidRSVP),
		errors.Is(err, domain.ErrNotNominated), errors.Is(err, pagination.ErrInvalid):
		return fiber.StatusBadRequest
	case errors.Is(err, domain.ErrForbidden):
		return fiber.StatusForbidden
	case errors.Is(err, domain.ErrEventNotFound), errors.Is(err, domain.ErrSquadNotFound),
		errors.Is(err, domain.ErrNominationNotFound), errors.Is(err, domain.ErrVenueNotFound):
		return fiber.StatusNotFound
	case errors.Is(err, domain.ErrInvalidTransition), errors.Is(err, domain.ErrNoNominations),
		errors.Is(err, domain.ErrAlreadyNominated):
		return fiber.StatusConflict
	case errors.Is(err, domain.ErrUserServiceUnavailable), errors.Is(err, domain.ErrVenueServiceUnavailable):
		return fiber.StatusServiceUnavailable
	default:
		return fiber.StatusInternalServerError
	}
}
