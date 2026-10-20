package http

import (
	"testing"

	"github.com/SamPariat/roundup/shared/api/apitest"
	"github.com/gofiber/fiber/v3"
)

func TestRoutesAreDocumented(t *testing.T) {
	app := fiber.New()
	RegisterRoutes(app, &Handlers{Venue: &VenueHandler{}, Favorite: &FavoriteHandler{}, History: &HistoryHandler{}})

	apitest.AssertRoutesDocumented(t, app, "/api/v1/venues")
	apitest.AssertRoutesDocumented(t, app, "/api/v1/squads")
}
