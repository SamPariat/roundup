package dtos

// response.go defines the standard JSON envelope used by all venue-service endpoints.
// Every handler returns Response[T] so clients always receive a consistent shape.

// Response is the standard JSON envelope for all API responses.
// On success, Data is populated and Error is empty. On failure, the inverse.
type Response[T any] struct {
	Success bool   `json:"success"`
	Data    T      `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

// OK wraps data in a successful Response envelope.
func OK[T any](data T) Response[T] {
	return Response[T]{Success: true, Data: data}
}

// Fail returns a failed Response envelope with the given error message.
func Fail[T any](error string) Response[T] {
	return Response[T]{Success: false, Error: error}
}
