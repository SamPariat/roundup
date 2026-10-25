package http

import (
	"net/http/httptest"
	"testing"

	"github.com/SamPariat/roundup/shared/middleware"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pagination_test.go checks that list routes reject a bad limit with 400
// before reaching a use case (the handlers here have none).

func TestListRoutesRejectBadLimit(t *testing.T) {
	app := fiber.New()
	RegisterRoutes(app, &Handlers{User: &UserHandler{}, DeviceToken: &DeviceTokenHandler{}, Squad: &SquadHandler{}})

	for _, path := range []string{"/api/v1/squads"} {
		for _, limit := range []string{"0", "101", "ten"} {
			req := httptest.NewRequest("GET", path+"?limit="+limit, nil)
			req.Header.Set(middleware.UserIDHeader, "7dc92d84-1950-43d6-b676-debacf84d830")

			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode, path+"?limit="+limit)
		}
	}
}
