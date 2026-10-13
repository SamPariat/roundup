package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestLoggingRecordsRequestFields(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	log := zap.New(core)

	app := fiber.New()
	app.Use(RequestID())
	app.Use(Logging(log))
	app.Get("/venues", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/venues", nil)
	_, err := app.Test(req)
	require.NoError(t, err)

	require.Equal(t, 1, logs.Len())
	entry := logs.All()[0]
	fields := entry.ContextMap()

	assert.Equal(t, "GET", fields["method"])
	assert.Equal(t, "/venues", fields["path"])
	assert.EqualValues(t, fiber.StatusOK, fields["status"])
	assert.NotEmpty(t, fields["request_id"])
}
