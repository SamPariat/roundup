package middleware

import (
	"fmt"
	"runtime/debug"

	"github.com/SamPariat/roundup/shared/response"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// Recover turns a panic in a handler into a logged 500 instead of a crashed
// process. Register it after Logging so the request log records the 500.
// The client gets the panic value in the envelope; the stack goes only to
// the log.
func Recover(log *zap.Logger) fiber.Handler {
	return func(c fiber.Ctx) (err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic recovered",
					zap.Any("panic", r),
					zap.ByteString("stack", debug.Stack()),
					zap.String("request_id", RequestIDFrom(c)),
				)
				err = response.Error(c, fiber.StatusInternalServerError, fmt.Errorf("panic: %v", r))
			}
		}()

		return c.Next()
	}
}
