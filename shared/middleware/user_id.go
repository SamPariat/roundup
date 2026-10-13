package middleware

import (
	"github.com/SamPariat/roundup/shared/response"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// UserIDHeader carries the authenticated caller's user ID. The gateway sets
// it after validating the access token; upstream services trust it.
const UserIDHeader = "X-User-ID"

// UserIDLocalsKey is the fiber.Locals key the caller's user ID is stored under.
const UserIDLocalsKey = "user_id"

// UserID requires a valid UUID in X-User-ID, rejecting the request with 401
// otherwise. On success the ID is stored in fiber.Locals for handlers to read
// via UserIDFrom.
func UserID() fiber.Handler {
	return func(c fiber.Ctx) error {
		id := c.Get(UserIDHeader)
		if _, err := uuid.Parse(id); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(
				response.Fail[any]("missing or invalid " + UserIDHeader + " header"),
			)
		}

		c.Locals(UserIDLocalsKey, id)

		return c.Next()
	}
}

// UserIDFrom returns the caller's user ID stored by the UserID middleware,
// or an empty string if the middleware did not run.
func UserIDFrom(c fiber.Ctx) string {
	id, _ := c.Locals(UserIDLocalsKey).(string)
	return id
}
