package response

import "github.com/gofiber/fiber/v3"

// errorKey is the fiber.Locals key the original error is stored under;
// unexported so only Error can set it.
type errorKey struct{}

// Error writes a failed envelope with the given status and err's message,
// including driver errors and internal URLs on 5xx: the full error is shown
// to the client on purpose. It is also kept in Locals for the request log.
func Error(c fiber.Ctx, status int, err error) error {
	c.Locals(errorKey{}, err)

	return c.Status(status).JSON(Fail[any](err.Error()))
}

// ErrorFrom returns the error saved by Error during this request, or nil.
func ErrorFrom(c fiber.Ctx) error {
	err, _ := c.Locals(errorKey{}).(error)

	return err
}
