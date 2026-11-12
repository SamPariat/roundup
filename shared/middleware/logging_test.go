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
		require.Equal(t, parsed[0]["request_id"], parsed[1]["request_id"], "both lines share the request ID")

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
	assert.NotEmpty(t, lines.req["request_id"])
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

	assert.Equal(t, map[string]any{"phone": "+919876543210"}, lines.req["req_body"])
	assert.Equal(t, map[string]any{"success": true}, lines.resp["resp_body"])
	assert.NotContains(t, lines.req, "resp_body")
	assert.NotContains(t, lines.resp, "req_body")
}

func TestLoggingAccessWritesOneLineWithoutBodies(t *testing.T) {
	lines := postJSON(t, LogAccess, `{"phone":"+919876543210"}`, `{"success":true}`)

	assert.Equal(t, "POST", lines.resp["method"])
	assert.Equal(t, "/echo", lines.resp["path"])
	assert.EqualValues(t, fiber.StatusOK, lines.resp["status"])
	assert.Contains(t, lines.resp, "latency")
	assert.NotEmpty(t, lines.resp["request_id"])
	assert.NotContains(t, lines.resp, "req_body")
	assert.NotContains(t, lines.resp, "resp_body")
}

func TestLoggingRedactsSensitiveKeys(t *testing.T) {
	lines := postJSON(t, LogFull,
		`{"phone":"+919876543210","code":"123456"}`,
		`{"data":[{"accessToken":"a","refreshToken":{"v":"r"},"userID":"u"}]}`,
	)

	assert.Equal(t, map[string]any{"phone": "+919876543210", "code": redacted}, lines.req["req_body"])
	assert.Equal(t,
		map[string]any{"data": []any{map[string]any{"accessToken": redacted, "refreshToken": redacted, "userID": "u"}}},
		lines.resp["resp_body"],
	)
}

func TestLoggingKeepsInvalidJSONAsString(t *testing.T) {
	lines := postJSON(t, LogFull, `{"phone":`, `{"a":1} trailing`)

	assert.Equal(t, `{"phone":`, lines.req["req_body"])
	assert.Equal(t, `{"a":1} trailing`, lines.resp["resp_body"])
}

func TestLoggingSkipsNonJSONResponse(t *testing.T) {
	app, logLine := newLoggedApp(t, LogFull)
	app.Get("/photo", func(c fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, "image/jpeg")

		return c.SendStream(strings.NewReader("\xff\xd8\xff"))
	})

	_, err := app.Test(httptest.NewRequest("GET", "/photo", nil))
	require.NoError(t, err)

	assert.NotContains(t, logLine().resp, "resp_body")
}

func TestLoggingKeepsLargeBodiesWhole(t *testing.T) {
	note := strings.Repeat("x", 64<<10)
	lines := postJSON(t, LogFull, `{"note":"`+note+`"}`, `{}`)

	assert.Equal(t, map[string]any{"note": note}, lines.req["req_body"])
}
