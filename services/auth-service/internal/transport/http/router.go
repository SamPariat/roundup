package http

import (
	"github.com/SamPariat/roundup/shared/api"
	"github.com/gofiber/fiber/v3"
)

// Handlers group all HTTP handlers for dependency injection into RegisterRoutes.
// Add new handlers here as features are built — one field per handler.
type Handlers struct {
	Auth *AuthHandler
}

// RegisterRoutes mounts all route groups onto the Fiber app.
// All routes are versioned under /api/v1 and are public.
func RegisterRoutes(app *fiber.App, h *Handlers) {
	auth := app.Group(api.V1 + "/auth")

	auth.Post("/otp/request", h.Auth.RequestOTP)
	auth.Post("/otp/verify", h.Auth.VerifyOTP)
	auth.Post("/refresh", h.Auth.Refresh)
	auth.Post("/logout", h.Auth.Logout)
}
