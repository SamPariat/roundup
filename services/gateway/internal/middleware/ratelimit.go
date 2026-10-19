package middleware

import (
	"time"

	"github.com/SamPariat/roundup/shared/response"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

// Auth route limits: each client IP gets authRateLimitMax requests per
// authRateLimitWindow, counting every request.
const (
	authRateLimitMax    = 20
	authRateLimitWindow = 15 * time.Minute
)

// AuthRateLimit limits how often one client may call the unauthenticated
// /auth/* routes. It complements auth-service's per-phone limits (3 codes per
// 15 minutes, 5 guesses per code), which can't stop one client cycling
// through many phone numbers. Counters are kept in memory, so each gateway
// instance limits independently; that is acceptable while beta runs few
// instances.
func AuthRateLimit() fiber.Handler {
	cfg := authRateLimitConfig()
	cfg.LimitReached = func(c fiber.Ctx) error {
		return c.Status(fiber.StatusTooManyRequests).
			JSON(response.Fail[any]("too many requests, try again later"))
	}

	return limiter.New(cfg)
}

// authRateLimitConfig decides who is counted, how many requests they get and
// over what window.
func authRateLimitConfig() limiter.Config {
	return limiter.Config{
		Max:        authRateLimitMax,
		Expiration: authRateLimitWindow,
		KeyGenerator: func(c fiber.Ctx) string {
			return c.IP()
		},
	}
}
