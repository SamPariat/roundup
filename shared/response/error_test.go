package response

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/SamPariat/roundup/shared/domainerr"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestError(t *testing.T) {
	internal := errors.New(`dial tcp 10.0.0.7:5432: connect: connection refused`)

	lastAdmin := domainerr.Sentinel("LAST_ADMIN", "a squad needs at least one admin")

	tests := []struct {
		name     string
		status   int
		err      error
		wantMsg  string
		wantCode string
	}{
		{"ClientErrorShowsMessage", fiber.StatusBadRequest, errors.New("phone must be in E.164 format"), "phone must be in E.164 format", "VALIDATION"},
		{"ServerErrorShowsMessage", fiber.StatusInternalServerError, internal, internal.Error(), "INTERNAL"},
		{"CodedSentinelSetsCode", fiber.StatusConflict, lastAdmin, lastAdmin.Error(), "LAST_ADMIN"},
		{"CodeSurvivesWrapping", fiber.StatusConflict, fmt.Errorf("leave: %w", domainerr.New(lastAdmin, "squad 7")), "leave: a squad needs at least one admin: squad 7", "LAST_ADMIN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logged error

			app := fiber.New()
			app.Get("/", func(c fiber.Ctx) error {
				err := Error(c, tt.status, tt.err)
				logged = ErrorFrom(c)

				return err
			})

			resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
			require.NoError(t, err)

			var body Response[any]
			require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
			assert.Equal(t, tt.status, resp.StatusCode)
			assert.False(t, body.Success)
			assert.Equal(t, tt.wantMsg, body.Error)
			assert.Equal(t, tt.wantCode, body.Code)
			assert.Equal(t, tt.err, logged, "the full error is kept for the log")
		})
	}
}
