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
	// LogAccess writes one "access" line per request when it completes:
	// method, path, query, status, latency and request ID, like a proxy's
	// access log. The gateway uses it; the services behind it log the rest.
	LogAccess LogMode = iota

	// LogFull writes a "request" line when a request arrives and a
	// "response" line when it completes, with the JSON request and response
	// bodies in full. The internal services use it.
	LogFull
)

// Logging logs each request as mode says. Every line has the method, the
// path (the route alone, so lines group by route), the query (see
// queryField: locations rounded, sensitive keys redacted) and the request ID,
// so register it after RequestID. The line written on completion adds the
// status, the latency and any error saved by response.Error, and is logged at
// error level for a 5xx.
//
// In LogFull, the JSON request body goes on the request line as req_body and
// the JSON response body on the response line as resp_body, each whole and
// compacted onto one line, with sensitive keys redacted (see redact).
// Non-JSON bodies (e.g. streamed files) are never read.
func Logging(log *zap.Logger, mode LogMode) fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		base := []zap.Field{
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			queryField(c),
			zap.String("request_id", RequestIDFrom(c)),
		}

		if mode == LogFull {
			// Full slice expressions cap base, so each append copies it
			// instead of the request and response lines writing into one
			// shared array.
			reqFields := base[:len(base):len(base)]
			if isJSON(c.Get(fiber.HeaderContentType)) {
				reqFields = append(reqFields, bodyField("req_body", c.Body()))
			}

			log.Info("request", reqFields...)
		}

		err := c.Next()

		status := c.Response().StatusCode()
		fields := append(base[:len(base):len(base)],
			zap.Int("status", status),
			zap.Duration("latency", time.Since(start)),
		)

		// Checked before Body(): on a streamed response, Body() would read the
		// whole stream into memory.
		if mode == LogFull && isJSON(string(c.Response().Header.ContentType())) {
			fields = append(fields, bodyField("resp_body", c.Response().Body()))
		}

		if handlerErr := response.ErrorFrom(c); handlerErr != nil {
			fields = append(fields, zap.Error(handlerErr))
		}

		msg := "response"
		if mode == LogAccess {
			msg = "access"
		}

		if status >= fiber.StatusInternalServerError {
			log.Error(msg, fields...)
		} else {
			log.Info(msg, fields...)
		}

		return err
	}
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
