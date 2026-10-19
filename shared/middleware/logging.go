package middleware

import (
	"time"

	"github.com/SamPariat/roundup/shared/response"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// Logging logs method, path, status, latency, and request ID for every
// request. Register it after RequestID so the request ID is available in
// fiber.Locals by the time a request completes. An error saved by
// response.Error is included, and 5xx responses are logged at error level.
func Logging(log *zap.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		requestID, _ := c.Locals(RequestIDLocalsKey).(string)
		status := c.Response().StatusCode()
		fields := []zap.Field{
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", status),
			zap.Duration("latency", time.Since(start)),
			zap.String("request_id", requestID),
		}
		if handlerErr, ok := c.Locals(response.ErrorLocalsKey).(error); ok {
			fields = append(fields, zap.Error(handlerErr))
		}

		if status >= fiber.StatusInternalServerError {
			log.Error("request", fields...)
		} else {
			log.Info("request", fields...)
		}

		return err
	}
}
