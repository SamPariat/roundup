package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthRateLimitBlocksAfterMax(t *testing.T) {
	cfg := authRateLimitConfig()
	require.Positive(t, cfg.Max, "set Max explicitly rather than relying on the limiter default")
	require.Positive(t, cfg.Expiration, "set Expiration explicitly rather than relying on the limiter default")

	app := fiber.New()
	app.Post("/api/v1/auth/otp/request", AuthRateLimit(), func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusAccepted)
	})

	call := func() int {
		resp, err := app.Test(httptest.NewRequest(http.MethodPost, "/api/v1/auth/otp/request", nil))
		require.NoError(t, err)
		return resp.StatusCode
	}

	for i := range cfg.Max {
		require.Equal(t, fiber.StatusAccepted, call(), "request %d is within the limit", i+1)
	}
	assert.Equal(t, fiber.StatusTooManyRequests, call())
}
