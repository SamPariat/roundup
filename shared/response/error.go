package response

import (
	"errors"

	"github.com/gofiber/fiber/v3"
)

// errorKey is the fiber.Locals key the original error is stored under;
// unexported so only Error can set it.
type errorKey struct{}

// coder is implemented by domainerr.Coded sentinels.
type coder interface {
	error
	Code() string
}

// Error writes a failed envelope with the given status and err's message,
// including driver errors and internal URLs on 5xx: the full error is shown
// to the client on purpose. It is also kept in Locals for the request log.
// The code comes from the first coded sentinel in err's chain, else from status.
func Error(c fiber.Ctx, status int, err error) error {
	c.Locals(errorKey{}, err)

	return c.Status(status).JSON(FailWithCode[any](CodeOf(err, status), err.Error()))
}

// ErrorFrom returns the error saved by Error during this request, or nil.
func ErrorFrom(c fiber.Ctx) error {
	err, _ := c.Locals(errorKey{}).(error)

	return err
}

// CodeOf returns the code of the first coded error in err's chain, or the
// generic code for status when none carries one (e.g. a malformed body).
func CodeOf(err error, status int) string {
	if c, ok := errors.AsType[coder](err); ok {
		return c.Code()
	}

	return StatusCode(status)
}

// StatusCode is the generic code for an HTTP status, used for errors that have
// no domain-specific code.
func StatusCode(status int) string {
	switch status {
	case fiber.StatusBadRequest:
		return "VALIDATION"
	case fiber.StatusUnauthorized:
		return "UNAUTHORIZED"
	case fiber.StatusForbidden:
		return "FORBIDDEN"
	case fiber.StatusNotFound:
		return "NOT_FOUND"
	case fiber.StatusConflict:
		return "CONFLICT"
	case fiber.StatusTooManyRequests:
		return "RATE_LIMITED"
	case fiber.StatusNotImplemented:
		return "NOT_IMPLEMENTED"
	case fiber.StatusBadGateway:
		return "BAD_GATEWAY"
	case fiber.StatusServiceUnavailable:
		return "UNAVAILABLE"
	default:
		if status >= fiber.StatusInternalServerError {
			return "INTERNAL"
		}

		return "ERROR"
	}
}
