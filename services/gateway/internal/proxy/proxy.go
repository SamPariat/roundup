// Package proxy forwards gateway requests to upstream services.
package proxy

import (
	"sync"
	"time"

	sharedmw "github.com/SamPariat/roundup/shared/middleware"
	"github.com/SamPariat/roundup/shared/response"
	"github.com/gofiber/fiber/v3"
	fiberproxy "github.com/gofiber/fiber/v3/middleware/proxy"
)

var allowInternal sync.Once

// allowInternalUpstreams relaxes Fiber's proxy policy, which by default
// refuses private and loopback addresses to stop SSRF. Upstream services are
// internal by design, and the policy has no per-host allowlist, so private
// addresses are allowed. This is safe only because To never lets a client
// choose the host: the URL is always a configured base URL plus the request
// path, which starts with "/". Every other default (hop-by-hop header
// stripping, http/https only, no HTTPS downgrade) stays on.
func allowInternalUpstreams() {
	allowInternal.Do(func() {
		policy := fiberproxy.DefaultSecurityPolicy()
		policy.AllowPrivateIPs = true
		fiberproxy.WithSecurityPolicy(policy)
	})
}

// To returns a handler that forwards the request unchanged — same method,
// path, query, headers and body — to the service at baseURL (from config,
// never from the request), and relays its response. The request ID is passed
// along so upstream logs can be matched with the gateway's. An unreachable or
// slow upstream yields 502.
func To(baseURL string, timeout time.Duration) fiber.Handler {
	allowInternalUpstreams()

	return func(c fiber.Ctx) error {
		id, _ := c.Locals(sharedmw.RequestIDLocalsKey).(string)
		if id != "" {
			c.Request().Header.Set(sharedmw.RequestIDHeader, id)
		}

		url := baseURL + string(c.Request().URI().RequestURI())
		err := fiberproxy.DoTimeout(c, url, timeout)

		// The upstream's response replaces ours wholesale, so restore the
		// request ID the caller should see.
		if id != "" {
			c.Set(sharedmw.RequestIDHeader, id)
		}
		if err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(response.Fail[any]("upstream service unavailable"))
		}

		return nil
	}
}
