package middleware

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// RequestIDHeader is the header used to propagate the request ID between the
// gateway and upstream services, and back to the caller.
const RequestIDHeader = "X-Request-ID"

// requestIDKey is both the fiber.Locals key and the context key for the
// request ID; unexported so only this package can set it.
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

// RequestIDFrom returns the request ID stored by RequestID or
// GenerateRequestID, or "" if neither ran.
func RequestIDFrom(c fiber.Ctx) string {
	id, _ := c.Locals(requestIDKey{}).(string)

	return id
}

// RequestID is for internal services: it reuses the X-Request-ID the gateway
// forwarded, generating a UUID v4 only if absent (e.g. a direct call in
// local development).
func RequestID() fiber.Handler {
	return func(c fiber.Ctx) error {
		id := c.Get(RequestIDHeader)
		if id == "" {
			id = uuid.NewString()
		}

		return withID(c, id)
	}
}

// GenerateRequestID is for the gateway, where requests come from the
// internet: it always generates a UUID v4 and ignores any X-Request-ID the
// client sent, so every ID in the logs is gateway-made. The proxy overwrites
// the header with this ID before forwarding.
func GenerateRequestID() fiber.Handler {
	return func(c fiber.Ctx) error {
		return withID(c, uuid.NewString())
	}
}

// withID stores id in fiber.Locals and in the request's context.Context (for
// code below the handler, such as HTTP clients), and echoes it on the
// response so callers and upstream services can correlate logs.
func withID(c fiber.Ctx, id string) error {
	c.Locals(requestIDKey{}, id)
	c.SetContext(WithRequestID(c.Context(), id))
	c.Set(RequestIDHeader, id)

	return c.Next()
}
