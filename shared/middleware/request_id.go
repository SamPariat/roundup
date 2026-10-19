package middleware

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// RequestIDHeader is the header used to propagate the request ID between the
// gateway and upstream services, and back to the caller.
const RequestIDHeader = "X-Request-Id"

// RequestIDLocalsKey is the fiber.Locals key the request ID is stored under
// for downstream middleware (e.g. Logging) and handlers to read.
const RequestIDLocalsKey = "request_id"

// requestIDKey is the context key for the request ID; unexported so only this
// package can set it.
type requestIDKey struct{}

// WithRequestID returns a copy of ctx carrying the request ID.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

// RequestIDFromContext returns the request ID stored by RequestID, or "" if
// there is none. Outgoing HTTP clients use it to pass the ID downstream.
func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

// RequestID reads X-Request-Id from the incoming request, generating a
// UUID v4 if absent. It stores the ID in fiber.Locals and in the request's
// context.Context (for code below the handler, such as HTTP clients), and
// echoes it on the response so callers and upstream services can correlate
// logs.
func RequestID() fiber.Handler {
	return func(c fiber.Ctx) error {
		id := c.Get(RequestIDHeader)
		if id == "" {
			id = uuid.NewString()
		}

		c.Locals(RequestIDLocalsKey, id)
		c.SetContext(WithRequestID(c.Context(), id))
		c.Set(RequestIDHeader, id)

		return c.Next()
	}
}
