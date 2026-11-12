package middleware

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/SamPariat/roundup/shared/response"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// LogMode is how much Logging writes per request.
type LogMode int

const (
	// LogAccess writes one "access" line per request when it completes, with
	// Traefik's access log field names (see accessFields): who asked, the
	// request, the status and size of the answer, the route and upstream
	// that served it, and how long it took. The gateway uses it; the
	// services behind it log the bodies.
	LogAccess LogMode = iota

	// LogFull writes a "request" line when a request arrives and a
	// "response" line when it completes, with the JSON request and response
	// bodies in full. The internal services use it.
	LogFull
)

// Logging logs each request as mode says, so register it after RequestID.
// Lines carry the query (see queryField: locations rounded, sensitive keys
// redacted) and the request ID, and the line written on completion is logged
// at error level for a 5xx, with any error saved by response.Error.
//
// In LogFull, every line has the method and the path (the route alone, so
// lines group by route); the response line adds the status and latency. The
// JSON request body goes on the request line as reqBody and the JSON
// response body on the response line as respBody, each whole and compacted
// onto one line, with sensitive keys redacted (see redact). Non-JSON bodies
// (e.g. streamed files) are never read.
func Logging(log *zap.Logger, mode LogMode) fiber.Handler {
	if mode == LogAccess {
		return accessLog(log)
	}

	return func(c fiber.Ctx) error {
		start := time.Now()
		base := []zap.Field{
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			queryField(c),
			zap.String("requestID", RequestIDFrom(c)),
		}

		// Full slice expressions cap base, so each append copies it instead
		// of the request and response lines writing into one shared array.
		reqFields := base[:len(base):len(base)]
		if isJSON(c.Get(fiber.HeaderContentType)) {
			reqFields = append(reqFields, bodyField("reqBody", c.Body()))
		}

		log.Info("request", reqFields...)

		err := c.Next()

		status := c.Response().StatusCode()
		fields := append(base[:len(base):len(base)],
			zap.Int("status", status),
			zap.Duration("latency", time.Since(start)),
		)

		// Checked before Body(): on a streamed response, Body() would read the
		// whole stream into memory.
		if isJSON(string(c.Response().Header.ContentType())) {
			fields = append(fields, bodyField("respBody", c.Response().Body()))
		}

		logCompleted(log, c, "response", status, fields)

		return err
	}
}

// accessLog is Logging in LogAccess: one "access" line per request, when it
// completes, with accessFields.
func accessLog(log *zap.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		status := c.Response().StatusCode()
		logCompleted(log, c, "access", status, accessFields(c, status, time.Since(start)))

		return err
	}
}

// accessFields are the access line's fields, named as in Traefik's JSON access
// log. clientHost is the address the connection came from (behind a load
// balancer, the balancer); the client's own address, as the balancer saw it,
// is in requestXForwardedFor when present. routerName is the route that
// matched, and serviceURL the upstream the gateway's proxy sent the request
// to (see SetServiceURL), absent when the gateway answered itself.
func accessFields(c fiber.Ctx, status int, took time.Duration) []zap.Field {
	fields := []zap.Field{
		zap.String("clientHost", c.IP()),
		zap.String("requestMethod", c.Method()),
		zap.String("requestPath", c.Path()),
		queryField(c),
		zap.String("requestProtocol", c.Protocol()),
		zap.Int("downstreamStatus", status),
		zap.Int("downstreamContentSize", len(c.Response().Body())),
		zap.String("requestUserAgent", c.Get(fiber.HeaderUserAgent)),
		zap.String("routerName", c.Route().Path),
		zap.Duration("duration", took),
		zap.String("requestID", RequestIDFrom(c)),
	}

	if forwarded := c.Get(fiber.HeaderXForwardedFor); forwarded != "" {
		fields = append(fields, zap.String("requestXForwardedFor", forwarded))
	}

	if url, ok := c.Locals(serviceURLKey).(string); ok {
		fields = append(fields, zap.String("serviceURL", url))
	}

	return fields
}

// logCompleted writes the line for a completed request, with the error saved
// by response.Error if any, at error level for a 5xx.
func logCompleted(log *zap.Logger, c fiber.Ctx, msg string, status int, fields []zap.Field) {
	if handlerErr := response.ErrorFrom(c); handlerErr != nil {
		fields = append(fields, zap.Error(handlerErr))
	}

	if status >= fiber.StatusInternalServerError {
		log.Error(msg, fields...)
	} else {
		log.Info(msg, fields...)
	}
}

// serviceURLKey is the Locals key SetServiceURL stores the upstream under.
const serviceURLKey = "middleware.serviceURL"

// SetServiceURL records which upstream a proxied request went to, for the
// access line's serviceURL. The gateway's proxy calls it.
func SetServiceURL(c fiber.Ctx, url string) {
	c.Locals(serviceURLKey, url)
}

// isJSON reports whether a Content-Type header names JSON, ignoring
// parameters such as "; charset=utf-8".
func isJSON(contentType string) bool {
	return strings.HasPrefix(strings.ToLower(contentType), fiber.MIMEApplicationJSON)
}

// bodyField turns a body into a log field that keeps the line single-line:
// valid JSON is redacted and compacted and logged as nested JSON; anything
// else is logged whole as an escaped string. An empty body adds nothing.
func bodyField(key string, b []byte) zap.Field {
	if len(b) == 0 {
		return zap.Skip()
	}

	out, ok := redact(b)
	if !ok {
		return zap.ByteString(key, b)
	}

	// Reflect, not Any: it goes through encoding/json, which writes a
	// RawMessage verbatim, so the body is nested JSON rather than a quoted
	// string full of \" escapes.
	return zap.Reflect(key, json.RawMessage(out))
}
