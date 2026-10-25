package http

import (
	"testing"

	"github.com/SamPariat/roundup/shared/api"
	"github.com/SamPariat/roundup/shared/api/apitest"
	"github.com/gofiber/fiber/v3"
)

func TestRoutesAreDocumented(t *testing.T) {
	app := fiber.New()
	RegisterRoutes(app, &Handlers{Event: &EventHandler{}})

	apitest.AssertRoutesDocumented(t, app, api.EventService)
}
