package http

import (
	"time"

	"github.com/SamPariat/roundup/shared/api"
	"github.com/gofiber/fiber/v3"
)

const (
	// refreshCookieName is the httpOnly cookie that carries a web client's
	// refresh token, so page scripts never see it.
	refreshCookieName = "roundup_refresh"

	// refreshCookiePath limits the cookie to the auth routes, the only ones
	// that need it.
	refreshCookiePath = api.V1 + "/auth"

	// clearCookieMaxAge tells the browser to drop the cookie now.
	clearCookieMaxAge = -1
)

// isWebClient reports whether the request came from the browser app. The
// custom header forces a CORS preflight for cross-site calls, which with
// SameSite=Strict and the gateway's origin allowlist covers CSRF.
func isWebClient(c fiber.Ctx) bool {
	return c.Get(api.ClientHeader) == api.ClientWeb
}

// refreshCookie builds the cookie that holds a refresh token until expiresAt.
// secure should be false only for plain-http local development.
func refreshCookie(token string, expiresAt time.Time, secure bool) *fiber.Cookie {
	return &fiber.Cookie{
		Name:     refreshCookieName,
		Value:    token,
		Path:     refreshCookiePath,
		Expires:  expiresAt,
		HTTPOnly: true,
		Secure:   secure,
		SameSite: fiber.CookieSameSiteStrictMode,
	}
}

// clearedRefreshCookie builds the cookie that removes a stored refresh token:
// the same cookie, empty and already expired.
func clearedRefreshCookie(secure bool) *fiber.Cookie {
	cookie := refreshCookie("", time.Unix(0, 0), secure)
	cookie.MaxAge = clearCookieMaxAge

	return cookie
}
