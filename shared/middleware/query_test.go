package middleware

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getLogged sends GET target to an app that answers 200 and returns the lines
// Logging wrote.
func getLogged(t *testing.T, target string) loggedLines {
	t.Helper()

	app, logLine := newLoggedApp(t, LogFull)
	app.Get("/venues", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	_, err := app.Test(httptest.NewRequest("GET", target, nil))
	require.NoError(t, err)

	return logLine()
}

func TestLoggingRecordsTheQueryOnBothLinesBesideThePath(t *testing.T) {
	lines := getLogged(t, "/venues?radius=3000&type=bar&photoWidth=800")

	want := map[string]any{"radius": "3000", "type": "bar", "photoWidth": "800"}
	assert.Equal(t, want, lines.req["query"])
	assert.Equal(t, want, lines.resp["query"])
	assert.Equal(t, "/venues", lines.req["path"], "the path stays the route alone, so lines group by route")
}

func TestLoggingRoundsLocationsToAboutAKilometre(t *testing.T) {
	lines := getLogged(t, "/venues?lat=12.971598&lng=77.594566")

	assert.Equal(t, map[string]any{"lat": "12.97", "lng": "77.59"}, lines.req["query"])
}

func TestLoggingRedactsALocationThatIsNotANumber(t *testing.T) {
	lines := getLogged(t, "/venues?lat=my-home-address")

	assert.Equal(t, map[string]any{"lat": redacted}, lines.req["query"])
}

func TestLoggingRedactsSensitiveQueryKeys(t *testing.T) {
	lines := getLogged(t, "/venues?code=123456&refreshToken=abc")

	assert.Equal(t, map[string]any{"code": redacted, "refreshToken": redacted}, lines.req["query"])
}

func TestLoggingKeepsLongQueryValuesWhole(t *testing.T) {
	cursor := strings.Repeat("x", 4<<10)
	lines := getLogged(t, "/venues?cursor="+cursor)

	assert.Equal(t, map[string]any{"cursor": cursor}, lines.req["query"])
}

func TestLoggingAccessLineHasTheQuery(t *testing.T) {
	app, logLine := newLoggedApp(t, LogAccess)
	app.Get("/venues", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusUnauthorized)
	})

	_, err := app.Test(httptest.NewRequest("GET", "/venues?lat=12.971598&radius=3000", nil))
	require.NoError(t, err)

	lines := logLine()
	assert.Equal(t, map[string]any{"lat": "12.97", "radius": "3000"}, lines.resp["query"])
	assert.EqualValues(t, fiber.StatusUnauthorized, lines.resp["downstreamStatus"])
}

func TestLoggingAddsNoQueryWithoutOne(t *testing.T) {
	lines := getLogged(t, "/venues")

	assert.NotContains(t, lines.req, "query")
	assert.NotContains(t, lines.resp, "query")
}
