package middleware

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// loggedLines holds the fields of the lines Logging writes per request: in
// LogFull the request and response lines, in LogAccess only the access line,
// held in resp (req is nil).
type loggedLines struct {
	req, resp map[string]any
}

// newLoggedApp returns an app with RequestID and Logging mounted in mode, and
// a function that returns the lines Logging wrote, checking there are as many
// as mode writes. Lines are captured through a real JSON encoder, so tests see
// what production logs contain rather than raw field values.
func newLoggedApp(t *testing.T, mode LogMode) (*fiber.App, func() loggedLines) {
	t.Helper()

	var buf bytes.Buffer

	core := zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(&buf), zap.InfoLevel)

	app := fiber.New()
	app.Use(RequestID())
	app.Use(Logging(zap.New(core), mode))

	logLine := func() loggedLines {
		t.Helper()

		lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")

		parsed := make([]map[string]any, len(lines))
		for i, line := range lines {
			require.NoError(t, json.Unmarshal([]byte(line), &parsed[i]))
		}

		if mode == LogAccess {
			require.Len(t, parsed, 1, "a request is logged as one access line")
			require.Equal(t, "access", parsed[0]["msg"])

			return loggedLines{resp: parsed[0]}
		}

		require.Len(t, parsed, 2, "a request is logged as one request line and one response line")
		require.Equal(t, "request", parsed[0]["msg"])
		require.Equal(t, "response", parsed[1]["msg"])
		require.Equal(t, parsed[0]["requestID"], parsed[1]["requestID"], "both lines share the request ID")

		return loggedLines{req: parsed[0], resp: parsed[1]}
	}

	return app, logLine
}

func TestLoggingRecordsRequestFields(t *testing.T) {
	app, logLine := newLoggedApp(t, LogFull)
	app.Get("/venues", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	_, err := app.Test(httptest.NewRequest("GET", "/venues", nil))
	require.NoError(t, err)

	lines := logLine()
	assert.Equal(t, "GET", lines.req["method"])
	assert.Equal(t, "/venues", lines.req["path"])
	assert.NotEmpty(t, lines.req["requestID"])
	assert.NotContains(t, lines.req, "status")

	assert.Equal(t, "GET", lines.resp["method"])
	assert.Equal(t, "/venues", lines.resp["path"])
	assert.EqualValues(t, fiber.StatusOK, lines.resp["status"])
	assert.Contains(t, lines.resp, "latency")
}

// postJSON sends body as a JSON POST to /echo, which replies with resp, and
// returns the lines Logging wrote in mode.
func postJSON(t *testing.T, mode LogMode, body, resp string) loggedLines {
	t.Helper()

	app, logLine := newLoggedApp(t, mode)
	app.Post("/echo", func(c fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		return c.SendString(resp)
	})

	req := httptest.NewRequest("POST", "/echo", strings.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	_, err := app.Test(req)
	require.NoError(t, err)

	return logLine()
}

func TestLoggingAddsCompactedBodies(t *testing.T) {
	lines := postJSON(t, LogFull, "{\n  \"phone\": \"+919876543210\"\n}", `{"success":true}`)

	assert.Equal(t, map[string]any{"phone": "+919876543210"}, lines.req["reqBody"])
	assert.Equal(t, map[string]any{"success": true}, lines.resp["respBody"])
	assert.NotContains(t, lines.req, "respBody")
	assert.NotContains(t, lines.resp, "reqBody")
}

func TestLoggingAccessWritesOneLineWithoutBodies(t *testing.T) {
	lines := postJSON(t, LogAccess, `{"phone":"+919876543210"}`, `{"success":true}`)

	assert.Equal(t, "POST", lines.resp["requestMethod"])
	assert.Equal(t, "/echo", lines.resp["requestPath"])
	assert.Equal(t, "HTTP/1.1", lines.resp["requestProtocol"])
	assert.EqualValues(t, fiber.StatusOK, lines.resp["downstreamStatus"])
	assert.EqualValues(t, len(`{"success":true}`), lines.resp["downstreamContentSize"])
	assert.Equal(t, "/echo", lines.resp["routerName"])
	assert.Contains(t, lines.resp, "duration")
	assert.Contains(t, lines.resp, "clientHost")
	assert.NotEmpty(t, lines.resp["requestID"])
	assert.NotContains(t, lines.resp, "serviceURL", "the app answered itself, so no upstream")
	assert.NotContains(t, lines.resp, "reqBody")
	assert.NotContains(t, lines.resp, "respBody")
}

func TestLoggingAccessNamesTheUpstreamAndTheClient(t *testing.T) {
	app, logLine := newLoggedApp(t, LogAccess)
	app.Get("/api/v1/venues/:placeID", func(c fiber.Ctx) error {
		SetServiceURL(c, "http://venue-service:3006")

		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/api/v1/venues/p1", nil)
	req.Header.Set(fiber.HeaderUserAgent, "Roundup/1.0 (iPhone)")
	req.Header.Set(fiber.HeaderXForwardedFor, "203.0.113.7, 10.0.0.1")

	_, err := app.Test(req)
	require.NoError(t, err)

	lines := logLine()
	assert.Equal(t, "http://venue-service:3006", lines.resp["serviceURL"])
	assert.Equal(t, "/api/v1/venues/:placeID", lines.resp["routerName"], "the route pattern, not the path")
	assert.Equal(t, "Roundup/1.0 (iPhone)", lines.resp["requestUserAgent"])
	assert.Equal(t, "203.0.113.7, 10.0.0.1", lines.resp["requestXForwardedFor"])
}

func TestLoggingRedactsSensitiveKeys(t *testing.T) {
	lines := postJSON(t, LogFull,
		`{"phone":"+919876543210","code":"123456"}`,
		`{"data":[{"accessToken":"a","refreshToken":{"v":"r"},"userID":"u"}]}`,
	)

	assert.Equal(t, map[string]any{"phone": "+919876543210", "code": redacted}, lines.req["reqBody"])
	assert.Equal(t,
		map[string]any{"data": []any{map[string]any{"accessToken": redacted, "refreshToken": redacted, "userID": "u"}}},
		lines.resp["respBody"],
	)
}

func TestLoggingKeepsInvalidJSONAsString(t *testing.T) {
	lines := postJSON(t, LogFull, `{"phone":`, `{"a":1} trailing`)

	assert.Equal(t, `{"phone":`, lines.req["reqBody"])
	assert.Equal(t, `{"a":1} trailing`, lines.resp["respBody"])
}

func TestLoggingSkipsNonJSONResponse(t *testing.T) {
	app, logLine := newLoggedApp(t, LogFull)
	app.Get("/photo", func(c fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, "image/jpeg")

		return c.SendStream(strings.NewReader("\xff\xd8\xff"))
	})

	_, err := app.Test(httptest.NewRequest("GET", "/photo", nil))
	require.NoError(t, err)

	assert.NotContains(t, logLine().resp, "respBody")
}

func TestLoggingKeepsLargeBodiesWhole(t *testing.T) {
	note := strings.Repeat("x", 64<<10)
	lines := postJSON(t, LogFull, `{"note":"`+note+`"}`, `{}`)

	assert.Equal(t, map[string]any{"note": note}, lines.req["reqBody"])
}
