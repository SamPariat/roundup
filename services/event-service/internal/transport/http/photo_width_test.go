package http

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SamPariat/roundup/shared/middleware"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// photo_width_test.go checks that the routes answering with places' photos
// reject a photo width the client may not ask for with 400, before reaching
// the use case (the handler here has none). For a route that changes the
// event, that means a bad width changes nothing.

func TestPhotoRoutesRejectBadPhotoWidth(t *testing.T) {
	app := fiber.New()
	RegisterRoutes(app, &Handlers{Event: &EventHandler{}})

	requests := []struct {
		method string
		path   string
		body   string
	}{
		{"GET", "/api/v1/squads/s1/events?photoWidth=700", ""},
		{"GET", "/api/v1/events/e1?photoWidth=4800", ""},
		{"GET", "/api/v1/events/e1?photoWidth=wide", ""},
		{"PUT", "/api/v1/events/e1/rsvp?photoWidth=700", `{"status":"GOING"}`},
	}

	for _, r := range requests {
		req := httptest.NewRequest(r.method, r.path, strings.NewReader(r.body))
		req.Header.Set(middleware.UserIDHeader, "7dc92d84-1950-43d6-b676-debacf84d830")
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode, r.method+" "+r.path)
	}
}
