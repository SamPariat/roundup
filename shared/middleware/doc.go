// Package middleware holds the Fiber middleware every Go service mounts:
// request IDs, request logging, panic recovery, and the X-User-ID check.
// Values a middleware stores for later handlers are read through accessors
// (RequestIDFrom, UserIDFrom), never through raw fiber.Locals keys.
package middleware
