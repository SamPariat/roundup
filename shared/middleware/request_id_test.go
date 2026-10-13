package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestApp() *fiber.App {
	app := fiber.New()
	app.Use(RequestID())
	app.Get("/", func(c fiber.Ctx) error {
		id, _ := c.Locals(RequestIDLocalsKey).(string)
		return c.SendString(id)
	})

	return app
}

func TestRequestIDGeneratesWhenAbsent(t *testing.T) {
	app := newTestApp()

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	echoed := resp.Header.Get(RequestIDHeader)
	assert.NotEmpty(t, echoed)
}

func TestRequestIDPropagatesExisting(t *testing.T) {
	app := newTestApp()

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set(RequestIDHeader, "fixed-id-123")
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, "fixed-id-123", resp.Header.Get(RequestIDHeader))
}
