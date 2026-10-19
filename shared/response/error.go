package response

import "github.com/gofiber/fiber/v3"

// ErrorLocalsKey is the fiber.Locals key the original error is stored under,
// so the Logging middleware can include it in the request log.
const ErrorLocalsKey = "error"

// Error writes a failed envelope with the given status and err's message,
// including driver errors and internal URLs on 5xx: the full error is shown
// to the client on purpose. It is also kept in Locals for the request log.
func Error(c fiber.Ctx, status int, err error) error {
	c.Locals(ErrorLocalsKey, err)
	return c.Status(status).JSON(Fail[any](err.Error()))
}
