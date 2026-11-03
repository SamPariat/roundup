package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

// corsMaxAge is how long, in seconds, a browser may cache a preflight answer.
const corsMaxAge = 600

// CORS lets browsers on the given origins call the gateway with credentials
// (the web app's refresh cookie). Mount it before any rate limit or proxy so a
// preflight is answered here with 204 and never reaches an upstream.
//
// With no origins it does nothing: Fiber's cors treats an empty list as
// "allow everyone", which must never happen here.
func CORS(origins []string) fiber.Handler {
	if len(origins) == 0 {
		return func(c fiber.Ctx) error { return c.Next() }
	}

	return cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowCredentials: true,
		AllowHeaders:     []string{fiber.HeaderContentType, fiber.HeaderAuthorization, "X-Client"},
		ExposeHeaders:    []string{"X-Request-ID"},
		MaxAge:           corsMaxAge,
	})
}
