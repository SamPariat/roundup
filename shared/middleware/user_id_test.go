package middleware

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newUserIDTestApp() *fiber.App {
	app := fiber.New()
	app.Use(UserID())
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString(UserIDFrom(c))
	})

	return app
}

func TestUserIDRejectsMissingHeader(t *testing.T) {
	app := newUserIDTestApp()

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestUserIDRejectsNonUUID(t *testing.T) {
	app := newUserIDTestApp()

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set(UserIDHeader, "not-a-uuid")
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestUserIDStoresValidID(t *testing.T) {
	app := newUserIDTestApp()
	id := "3f1c2a9e-5b7d-4e8f-9a0b-1c2d3e4f5a6b"

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set(UserIDHeader, id)
	resp, err := app.Test(req)
	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, id, string(body))
}
