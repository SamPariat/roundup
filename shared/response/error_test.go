package response

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestError(t *testing.T) {
	internal := errors.New(`dial tcp 10.0.0.7:5432: connect: connection refused`)

	tests := []struct {
		name    string
		status  int
		err     error
		wantMsg string
	}{
		{"ClientErrorShowsMessage", fiber.StatusBadRequest, errors.New("phone must be in E.164 format"), "phone must be in E.164 format"},
		{"ServerErrorShowsMessage", fiber.StatusInternalServerError, internal, internal.Error()},
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
			assert.Equal(t, tt.err, logged, "the full error is kept for the log")
		})
	}
}
