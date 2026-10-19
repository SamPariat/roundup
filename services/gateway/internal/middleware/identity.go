// Package middleware holds the gateway's identity middleware: the only place
// where a caller's identity is established for the rest of the system.
package middleware

import (
	"crypto/ed25519"
	"strings"

	"github.com/SamPariat/roundup/shared/authtoken"
	sharedmw "github.com/SamPariat/roundup/shared/middleware"
	"github.com/SamPariat/roundup/shared/response"
	"github.com/gofiber/fiber/v3"
)

// StripInternalHeaders removes headers that only the gateway may set, so a
// client can't supply them:
//   - X-User-ID: upstream services trust it completely; only RequireAuth sets it.
//   - X-Request-Id: the gateway generates every request ID, so IDs in the logs
//     are always gateway-made UUIDs, never client-chosen strings.
//
// Mount it first, before RequestID, which would otherwise reuse a client value.
func StripInternalHeaders() fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Request().Header.Del(sharedmw.UserIDHeader)
		c.Request().Header.Del(sharedmw.RequestIDHeader)
		return c.Next()
	}
}

// RequireAuth verifies the "Authorization: Bearer <access token>" header with
// pub and forwards the token's user as X-User-ID. Requests without a valid
// token get 401 and never reach an upstream service.
func RequireAuth(pub ed25519.PublicKey) fiber.Handler {
	return func(c fiber.Ctx) error {
		token, ok := strings.CutPrefix(c.Get(fiber.HeaderAuthorization), "Bearer ")
		if !ok || token == "" {
			return unauthorized(c, "missing bearer token")
		}

		claims, err := authtoken.Verify(pub, token)
		if err != nil {
			return unauthorized(c, "invalid or expired access token")
		}

		c.Request().Header.Set(sharedmw.UserIDHeader, claims.UserID)
		c.Locals(sharedmw.UserIDLocalsKey, claims.UserID)

		return c.Next()
	}
}

func unauthorized(c fiber.Ctx, msg string) error {
	c.Set(fiber.HeaderWWWAuthenticate, "Bearer")
	return c.Status(fiber.StatusUnauthorized).JSON(response.Fail[any](msg))
}
