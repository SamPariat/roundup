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

// Logging logs two lines per request, joined by request_id: a "request" line
// when it arrives (method, path, query) and a "response" line when it
// completes (method, path, query, status, latency). path is the route's path
// alone, so lines group by route; query holds its parameters, with locations
// rounded and sensitive keys redacted (see queryField). Register it after RequestID so the request
// ID is available on arrival. An error saved by response.Error goes on the
// response line, and 5xx responses are logged at error level.
//
// With logBodies, the JSON request body is added to the request line as
// req_body and the JSON response body to the response line as resp_body,
// compacted onto one line each, with sensitive keys redacted (see redact).
// Non-JSON bodies (e.g. streamed photos) are never read. The gateway passes
// false; internal services pass true.
func Logging(log *zap.Logger, logBodies bool) fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		base := []zap.Field{
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			queryField(c),
			zap.String("request_id", RequestIDFrom(c)),
		}

		// Full slice expressions cap base, so each append copies it instead of
		// the request and response lines writing into one shared array.
		reqFields := base[:len(base):len(base)]
		if logBodies && isJSON(c.Get(fiber.HeaderContentType)) {
			reqFields = append(reqFields, bodyField("req_body", c.Body()))
		}

		log.Info("request", reqFields...)

		err := c.Next()

		status := c.Response().StatusCode()
		respFields := append(base[:len(base):len(base)],
			zap.Int("status", status),
			zap.Duration("latency", time.Since(start)),
		)

		// Checked before Body(): on a streamed response, Body() would read the
		// whole stream into memory.
		if logBodies && isJSON(string(c.Response().Header.ContentType())) {
			respFields = append(respFields, bodyField("resp_body", c.Response().Body()))
		}

		if handlerErr := response.ErrorFrom(c); handlerErr != nil {
			respFields = append(respFields, zap.Error(handlerErr))
		}

		if status >= fiber.StatusInternalServerError {
			log.Error("response", respFields...)
		} else {
			log.Info("response", respFields...)
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
	return truncateTo(b, maxLoggedBody)
}

// truncateTo cuts b to limit bytes, noting how much was dropped.
func truncateTo(b []byte, limit int) []byte {
	if len(b) <= limit {
		return b
	}

	return fmt.Appendf(b[:limit:limit], "…(truncated %d bytes)", len(b)-limit)
}
