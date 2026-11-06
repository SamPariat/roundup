package middleware

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/SamPariat/roundup/shared/response"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// maxLoggedBody caps how many bytes of a request or response body go into a
// log line; longer bodies are cut and marked as truncated.
const maxLoggedBody = 8 << 10

// Logging logs method, path, status, latency, and request ID for every
// request, as one line. Register it after RequestID so the request ID is
// available by the time a request completes. An error saved by
// response.Error is included, and 5xx responses are logged at error level.
//
// With logBodies, JSON request and response bodies are added as req_body and
// resp_body, compacted onto the same line, with sensitive keys redacted (see
// redact). Non-JSON bodies (e.g. streamed photos) are never read. The gateway
// passes false; internal services pass true.
func Logging(log *zap.Logger, logBodies bool) fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		status := c.Response().StatusCode()
		fields := []zap.Field{
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", status),
			zap.Duration("latency", time.Since(start)),
			zap.String("request_id", RequestIDFrom(c)),
		}

		if logBodies {
			if isJSON(c.Get(fiber.HeaderContentType)) {
				fields = append(fields, bodyField("req_body", c.Body()))
			}

			// Checked before Body(): on a streamed response, Body() would read
			// the whole stream into memory.
			if isJSON(string(c.Response().Header.ContentType())) {
				fields = append(fields, bodyField("resp_body", c.Response().Body()))
			}
		}

		if handlerErr := response.ErrorFrom(c); handlerErr != nil {
			fields = append(fields, zap.Error(handlerErr))
		}

		if status >= fiber.StatusInternalServerError {
			log.Error("request", fields...)
		} else {
			log.Info("request", fields...)
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
// else is logged as an escaped string. Bodies over maxLoggedBody are cut and
// logged as a string. An empty body adds nothing.
func bodyField(key string, b []byte) zap.Field {
	if len(b) == 0 {
		return zap.Skip()
	}

	out, ok := redact(b)
	if !ok {
		return zap.ByteString(key, truncate(b))
	}

	if len(out) > maxLoggedBody {
		return zap.ByteString(key, truncate(out))
	}

	// Reflect, not Any: it goes through encoding/json, which writes a
	// RawMessage verbatim, so the body is nested JSON rather than a quoted
	// string full of \" escapes.
	return zap.Reflect(key, json.RawMessage(out))
}

// truncate cuts b to maxLoggedBody bytes, noting how much was dropped.
func truncate(b []byte) []byte {
	if len(b) <= maxLoggedBody {
		return b
	}

	return fmt.Appendf(b[:maxLoggedBody:maxLoggedBody], "…(truncated %d bytes)", len(b)-maxLoggedBody)
}
