// Package domainerr adds context to a service's sentinel errors without
// changing their identity. Every service declares its own sentinels (e.g.
// domain.ErrUserNotFound) and wraps them with New, so errors.Is keeps matching
// the sentinel while the message carries the detail.
package domainerr

import "fmt"

// Error wraps a sentinel error with additional context. It preserves the
// sentinel's identity through the error chain so that errors.Is continues to
// work correctly at the call site.
type Error struct {
	// Sentinel is the stable, matchable root error.
	Sentinel error

	// Detail provides human-readable context without altering the sentinel identity.
	Detail string
}

// New constructs an Error that wraps sentinel with the supplied detail.
//
// Example:
//
//	err := domainerr.New(domain.ErrUserNotFound, "id: 3f1c2a9e-...")
//	errors.Is(err, domain.ErrUserNotFound) // true
func New(sentinel error, detail string) *Error {
	return &Error{Sentinel: sentinel, Detail: detail}
}

// Error implements the error interface, formatting the sentinel and detail together.
func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Sentinel, e.Detail)
}

// Unwrap returns the sentinel so that errors.Is can traverse the chain.
func (e *Error) Unwrap() error {
	return e.Sentinel
}
