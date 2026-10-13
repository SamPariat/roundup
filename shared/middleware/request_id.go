package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// RequestIDHeader is the header used to propagate the request ID between the
// gateway and upstream services, and back to the caller.
const RequestIDHeader = "X-Request-Id"

// RequestIDLocalsKey is the fiber.Locals key the request ID is stored under
// for downstream middleware (e.g. Logging) and handlers to read.
const RequestIDLocalsKey = "request_id"

// RequestID reads X-Request-Id from the incoming request, generating a
// UUID v4 if absent. It stores the ID in fiber.Locals and echoes it on the
// response so callers and upstream services can correlate logs.
func RequestID() fiber.Handler {
	return func(c fiber.Ctx) error {
		id := c.Get(RequestIDHeader)
		if id == "" {
			id = uuid.NewString()
		}

		c.Locals(RequestIDLocalsKey, id)
		c.Set(RequestIDHeader, id)

		return c.Next()
	}
}
