package http

import (
	"errors"
	"testing"

	"github.com/SamPariat/roundup/services/user-service/internal/domain"
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
		{domain.ErrInvalidPhone, fiber.StatusBadRequest},
		{domain.ErrInvalidName, fiber.StatusBadRequest},
		{domain.ErrTooManyPhones, fiber.StatusBadRequest},
		{domain.ErrInvalidToken, fiber.StatusBadRequest},
		{domain.ErrInvalidPlatform, fiber.StatusBadRequest},
		{domain.ErrInvalidSquadName, fiber.StatusBadRequest},
		{domain.ErrInvalidRole, fiber.StatusBadRequest},
		{pagination.ErrInvalid, fiber.StatusBadRequest},
		{domain.ErrForbidden, fiber.StatusForbidden},
		{domain.ErrUserNotFound, fiber.StatusNotFound},
		{domain.ErrDeviceTokenNotFound, fiber.StatusNotFound},
		{domain.ErrSquadNotFound, fiber.StatusNotFound},
		{domain.ErrMemberNotFound, fiber.StatusNotFound},
		{domain.ErrPhoneTaken, fiber.StatusConflict},
		{domain.ErrAlreadyMember, fiber.StatusConflict},
		{domain.ErrSquadFull, fiber.StatusConflict},
		{domain.ErrLastAdmin, fiber.StatusConflict},
		{domainerr.New(domain.ErrSquadNotFound, "s1"), fiber.StatusNotFound},
		{errors.New("boom"), fiber.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.err.Error(), func(t *testing.T) {
			assert.Equal(t, tt.want, errorStatus(tt.err))
		})
	}
}
