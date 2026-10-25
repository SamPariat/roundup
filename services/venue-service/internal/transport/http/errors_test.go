package http

import (
	"errors"
	"testing"

	"github.com/SamPariat/roundup/services/venue-service/internal/domain"
	"github.com/SamPariat/roundup/shared/domainerr"
	"github.com/SamPariat/roundup/shared/pagination"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
)

func TestErrorStatus(t *testing.T) {
	tests := []struct {
		err  error
		want int
	}{
		{domain.ErrInvalidSearch, fiber.StatusBadRequest},
		{pagination.ErrInvalid, fiber.StatusBadRequest},
		{domain.ErrSquadNotFound, fiber.StatusNotFound},
		{domain.ErrVenueNotFound, fiber.StatusNotFound},
		{domain.ErrPhotoNotFound, fiber.StatusNotFound},
		{domain.ErrNotSaved, fiber.StatusNotFound},
		{domain.ErrAlreadySaved, fiber.StatusConflict},
		{domain.ErrVisitAlreadyRecorded, fiber.StatusConflict},
		{domain.ErrProviderUnavailable, fiber.StatusServiceUnavailable},
		{domain.ErrUserServiceUnavailable, fiber.StatusServiceUnavailable},
		{domainerr.New(domain.ErrVenueNotFound, "abc"), fiber.StatusNotFound},
		{errors.New("boom"), fiber.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.err.Error(), func(t *testing.T) {
			assert.Equal(t, tt.want, errorStatus(tt.err))
		})
	}
}
