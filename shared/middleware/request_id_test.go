package middleware

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestApp() *fiber.App {
	app := fiber.New()
	app.Use(RequestID())
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString(RequestIDFrom(c))
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

func TestRequestIDIsInContext(t *testing.T) {
	app := fiber.New()
	app.Use(RequestID())
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString(RequestIDFromContext(c.Context()))
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set(RequestIDHeader, "ctx-id-456")
	resp, err := app.Test(req)
	require.NoError(t, err)

	body, _ := io.ReadAll(resp.Body)

	assert.Equal(t, "ctx-id-456", string(body), "code below the handler sees the ID via context")
}

func TestRequestIDFromContextEmptyWhenUnset(t *testing.T) {
	assert.Empty(t, RequestIDFromContext(t.Context()))
}

func TestGenerateRequestIDIgnoresClientHeader(t *testing.T) {
	app := fiber.New()
	app.Use(GenerateRequestID())
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString(RequestIDFromContext(c.Context()))
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set(RequestIDHeader, "client-chosen-id")
	resp, err := app.Test(req)
	require.NoError(t, err)

	body, _ := io.ReadAll(resp.Body)

	require.NoError(t, uuid.Validate(string(body)), "a fresh UUID is generated")
	assert.Equal(t, string(body), resp.Header.Get(RequestIDHeader))
}
