package middleware

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// Logging logs method, path, status, latency, and request ID for every
// request. Register it after RequestID so the request ID is available in
// fiber.Locals by the time a request completes.
func Logging(log *zap.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		requestID, _ := c.Locals(RequestIDLocalsKey).(string)
		log.Info("request",
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", c.Response().StatusCode()),
			zap.Duration("latency", time.Since(start)),
			zap.String("request_id", requestID),
		)

		return err
	}
}
