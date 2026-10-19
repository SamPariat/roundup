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

// StripIdentity removes any X-User-ID the client sent. Upstream services
// trust that header completely, so only RequireAuth may set it. Mount it
// globally, before any route.
func StripIdentity() fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Request().Header.Del(sharedmw.UserIDHeader)

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
		sharedmw.SetUserID(c, claims.UserID)

		return c.Next()
	}
}

func unauthorized(c fiber.Ctx, msg string) error {
	c.Set(fiber.HeaderWWWAuthenticate, "Bearer")

	return c.Status(fiber.StatusUnauthorized).JSON(response.Fail[any](msg))
}
