package middleware

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestRecoverTurnsPanicIntoLogged500(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	log := zap.New(core)

	app := fiber.New()
	app.Use(RequestID(), Logging(log), Recover(log))
	app.Get("/boom", func(fiber.Ctx) error { panic("secret internal state") })

	resp, err := app.Test(httptest.NewRequest("GET", "/boom", nil))
	require.NoError(t, err)
	body, _ := io.ReadAll(resp.Body)

	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	assert.Contains(t, string(body), "panic: secret internal state")

	panics := logs.FilterMessage("panic recovered").All()
	require.Len(t, panics, 1)
	assert.NotEmpty(t, panics[0].ContextMap()["stack"])

	requests := logs.FilterMessage("request").All()
	require.Len(t, requests, 1)
	assert.EqualValues(t, fiber.StatusInternalServerError, requests[0].ContextMap()["status"])
}
