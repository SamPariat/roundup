# transport/http/middleware

Fiber middleware applied globally to all routes via `server.go`.

## Files

| File | Purpose |
|---|---|
| `request_id.go` | Injects a unique `X-Request-ID` header into every request (generates one if not already present). Propagated to structured logs so a full request trace can be reconstructed from logs alone. |
| `logging.go` | Structured request logging using `log/slog`. Logs method, path, status code, latency, and request ID on every response. |

## Adding middleware

New middleware should be registered in `server.go` using `app.Use(...)`. Apply it before the route groups if it should cover all routes, or scope it to a specific group if not.

## Auth middleware

JWT validation is handled by the API gateway before requests reach this service. The venue-service trusts the `X-User-ID` and `X-Squad-ID` headers injected by the gateway and does not perform its own token verification.
