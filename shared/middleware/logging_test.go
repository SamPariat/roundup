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

// newLoggedApp returns an app with RequestID and Logging mounted, and a
// function that returns the fields of the one line Logging wrote. Lines are
// captured through a real JSON encoder, so tests see what production logs
// contain rather than raw field values.
func newLoggedApp(t *testing.T, logBodies bool) (*fiber.App, func() map[string]any) {
	t.Helper()

	var buf bytes.Buffer

	core := zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(&buf), zap.InfoLevel)

	app := fiber.New()
	app.Use(RequestID())
	app.Use(Logging(zap.New(core), logBodies))

	logLine := func() map[string]any {
		t.Helper()

		line := strings.TrimSuffix(buf.String(), "\n")
		require.NotEmpty(t, line, "request was logged")
		require.NotContains(t, line, "\n", "request is logged as exactly one line")

		var fields map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &fields))

		return fields
	}

	return app, logLine
}

func TestLoggingRecordsRequestFields(t *testing.T) {
	app, logLine := newLoggedApp(t, false)
	app.Get("/venues", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	_, err := app.Test(httptest.NewRequest("GET", "/venues", nil))
	require.NoError(t, err)

	fields := logLine()
	assert.Equal(t, "GET", fields["method"])
	assert.Equal(t, "/venues", fields["path"])
	assert.EqualValues(t, fiber.StatusOK, fields["status"])
	assert.NotEmpty(t, fields["request_id"])
}

// postJSON sends body as a JSON POST to /echo, which replies with resp, and
// returns the fields of the log line.
func postJSON(t *testing.T, logBodies bool, body, resp string) map[string]any {
	t.Helper()

	app, logLine := newLoggedApp(t, logBodies)
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
	fields := postJSON(t, true, "{\n  \"phone\": \"+919876543210\"\n}", `{"success":true}`)

	assert.Equal(t, map[string]any{"phone": "+919876543210"}, fields["req_body"])
	assert.Equal(t, map[string]any{"success": true}, fields["resp_body"])
}

func TestLoggingLeavesBodiesOutWhenDisabled(t *testing.T) {
	fields := postJSON(t, false, `{"phone":"+919876543210"}`, `{"success":true}`)

	assert.NotContains(t, fields, "req_body")
	assert.NotContains(t, fields, "resp_body")
}

func TestLoggingRedactsSensitiveKeys(t *testing.T) {
	fields := postJSON(t, true,
		`{"phone":"+919876543210","code":"123456"}`,
		`{"data":[{"accessToken":"a","refreshToken":{"v":"r"},"userID":"u"}]}`,
	)

	assert.Equal(t, map[string]any{"phone": "+919876543210", "code": redacted}, fields["req_body"])
	assert.Equal(t,
		map[string]any{"data": []any{map[string]any{"accessToken": redacted, "refreshToken": redacted, "userID": "u"}}},
		fields["resp_body"],
	)
}

func TestLoggingKeepsInvalidJSONAsString(t *testing.T) {
	fields := postJSON(t, true, `{"phone":`, `{"a":1} trailing`)

	assert.Equal(t, `{"phone":`, fields["req_body"])
	assert.Equal(t, `{"a":1} trailing`, fields["resp_body"])
}

func TestLoggingSkipsNonJSONResponse(t *testing.T) {
	app, logLine := newLoggedApp(t, true)
	app.Get("/photo", func(c fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, "image/jpeg")

		return c.SendStream(strings.NewReader("\xff\xd8\xff"))
	})

	_, err := app.Test(httptest.NewRequest("GET", "/photo", nil))
	require.NoError(t, err)

	assert.NotContains(t, logLine(), "resp_body")
}

func TestLoggingTruncatesLargeBody(t *testing.T) {
	big := `{"note":"` + strings.Repeat("x", maxLoggedBody) + `"}`
	fields := postJSON(t, true, big, `{}`)

	body, ok := fields["req_body"].(string)
	require.True(t, ok, "oversized body is logged as a string")
	assert.Contains(t, body, "…(truncated ")
}
