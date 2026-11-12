package http

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SamPariat/roundup/shared/api"
	"github.com/SamPariat/roundup/shared/middleware"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// photo_options_test.go checks that every route returning photos rejects a
// photo width or count the client may not ask for with 400, before reaching a
// use case (the handlers here have none) and so before any Google call.

func TestPhotoRoutesRejectBadPhotoOptions(t *testing.T) {
	app := fiber.New()
	RegisterRoutes(app, &Handlers{Venue: &VenueHandler{}, Favorite: &FavoriteHandler{}, History: &HistoryHandler{}})

	gets := []string{
		"/api/v1/venues?lat=12.97&lng=77.59&radius=3000&photoWidth=700",
		"/api/v1/venues/p1?photoWidth=4800",
		"/api/v1/venues/p1?photoCount=6",
		"/api/v1/venues/p1?photoCount=-1",
		"/api/v1/squads/s1/favorites?photoWidth=100",
		"/api/v1/squads/s1/history?photoWidth=1000",
	}

	for _, path := range gets {
		req := httptest.NewRequest("GET", path, nil)
		req.Header.Set(middleware.UserIDHeader, "7dc92d84-1950-43d6-b676-debacf84d830")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode, path)
	}

	req := httptest.NewRequest("POST", api.VenuePhotoURLsPath, strings.NewReader(`{"refs":["r1"],"maxWidth":700}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode, api.VenuePhotoURLsPath)
}
