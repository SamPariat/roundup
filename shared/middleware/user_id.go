package middleware

import (
	"github.com/SamPariat/roundup/shared/response"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// UserIDHeader carries the authenticated caller's user ID. The gateway sets
// it after validating the access token; upstream services trust it.
const UserIDHeader = "X-User-ID"

// userIDKey is the fiber.Locals key the caller's user ID is stored under;
// unexported so it is only set through UserID or SetUserID.
type userIDKey struct{}

// UserID requires a valid UUID in X-User-ID, rejecting the request with 401
// otherwise. On success the ID is stored for handlers to read via UserIDFrom.
func UserID() fiber.Handler {
	return func(c fiber.Ctx) error {
		id := c.Get(UserIDHeader)
		if _, err := uuid.Parse(id); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(
				response.FailWithCode[any]("UNAUTHORIZED", "missing or invalid "+UserIDHeader+" header"),
			)
		}

		SetUserID(c, id)

		return c.Next()
	}
}

// SetUserID records the caller's verified user ID for the rest of the
// request. Used by UserID and by the gateway once it has checked a token.
func SetUserID(c fiber.Ctx, id string) {
	c.Locals(userIDKey{}, id)
}

// UserIDFrom returns the caller's user ID stored by SetUserID, or an empty
// string if it never ran.
func UserIDFrom(c fiber.Ctx) string {
	id, _ := c.Locals(userIDKey{}).(string)

	return id
}
