// Package response defines the standard JSON envelope used by every Go
// service's HTTP API, so clients always receive a consistent shape
// regardless of which service answered the request.
package response

// Response is the standard JSON envelope for all API responses.
// On success, Data is populated and Code and Error are empty. On failure, the
// inverse: Code is stable for clients to branch on, Error is for people.
type Response[T any] struct {
	Success bool   `json:"success"`
	Data    T      `json:"data,omitempty"`
	Code    string `json:"code,omitempty"`
	Error   string `json:"error,omitempty"`
}

// OK wraps data in a successful Response envelope.
func OK[T any](data T) Response[T] {
	return Response[T]{Success: true, Data: data}
}

// Fail returns a failed Response envelope with the given error message and
// the generic "ERROR" code. Prefer FailWithCode.
func Fail[T any](msg string) Response[T] {
	return FailWithCode[T]("ERROR", msg)
}

// FailWithCode returns a failed Response envelope with a machine-readable code.
func FailWithCode[T any](code, msg string) Response[T] {
	return Response[T]{Success: false, Code: code, Error: msg}
}
