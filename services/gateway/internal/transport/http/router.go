// Package http registers the gateway's HTTP routes: which paths are public,
// which need an access token, and which upstream service each one goes to.
package http

import (
	"crypto/ed25519"
	"time"

	"github.com/SamPariat/roundup/services/gateway/internal/config"
	"github.com/SamPariat/roundup/services/gateway/internal/middleware"
	"github.com/SamPariat/roundup/services/gateway/internal/proxy"
	"github.com/SamPariat/roundup/shared/api"
	"github.com/SamPariat/roundup/shared/response"
	"github.com/gofiber/fiber/v3"
)

// RegisterRoutes mounts the /api/v1 routes. Client-sent identity headers must
// already be stripped (middleware.StripIdentity, mounted globally).
func RegisterRoutes(app *fiber.App, upstreams config.Upstreams, pub ed25519.PublicKey, timeout time.Duration) {
	auth := middleware.RequireAuth(pub)
	v1 := app.Group(api.V1)

	// Public: login and token refresh are how a client gets a token. Being
	// unauthenticated, they are rate limited per client IP.
	v1.All("/auth/*", middleware.AuthRateLimit(), proxy.To(upstreams.Auth, timeout))

	// Creating users is internal: auth-service calls user-service directly at
	// login. Registered before the /users group so it wins.
	app.Post(api.UsersPath, notFoundHandler)

	v1.Group("/users", auth).All("/*", proxy.To(upstreams.User, timeout))
	v1.Group("/venues", auth).All("/*", proxy.To(upstreams.Venue, timeout))

	// A squad's favorites and history live in venue-service; the squad itself
	// and its members live in user-service. The specific groups are registered
	// first so they win over the /squads catch-all.
	v1.Group("/squads/:squadID/favorites", auth).All("/*", proxy.To(upstreams.Venue, timeout))
	v1.Group("/squads/:squadID/history", auth).All("/*", proxy.To(upstreams.Venue, timeout))
	v1.Group("/squads", auth).All("/*", proxy.To(upstreams.User, timeout))

	// Services that don't exist yet.
	for _, name := range []string{"events", "billing", "media", "realtime"} {
		v1.All("/"+name+"/*", notImplementedHandler)
	}
}

// notFoundHandler answers 404 for routes the gateway deliberately hides, such
// as the internal POST /api/v1/users.
func notFoundHandler(c fiber.Ctx) error {
	return c.Status(fiber.StatusNotFound).JSON(response.Fail[any]("not found"))
}

// notImplementedHandler answers for route groups whose service isn't built yet.
func notImplementedHandler(c fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(response.Fail[any]("not implemented yet"))
}
