package domain

import "fmt"

// Sentinel errors are used as stable identities for errors.Is comparisons.
// Callers should match against these values, never against error strings.
var (
	// ErrUserNotFound is returned when no active user matches the lookup.
	ErrUserNotFound = fmt.Errorf("user not found")

	// ErrPhoneTaken is returned when creating a user whose phone number already
	// belongs to an active user.
	ErrPhoneTaken = fmt.Errorf("phone number already registered")

	// ErrInvalidPhone is returned when a phone number is not in E.164 format.
	ErrInvalidPhone = fmt.Errorf("phone must be in E.164 format, e.g. +919876543210")

	// ErrInvalidName is returned when a profile update sets an empty name.
	ErrInvalidName = fmt.Errorf("name must not be empty")

	// ErrInvalidPlatform is returned when a device token's platform is not iOS or Android.
	ErrInvalidPlatform = fmt.Errorf("platform must be iOS or Android")

	// ErrInvalidToken is returned when a device token is empty.
	ErrInvalidToken = fmt.Errorf("device token must not be empty")

	// ErrDeviceTokenNotFound is returned when removing a token the user does not own.
	ErrDeviceTokenNotFound = fmt.Errorf("device token not found")
)

// DomainError wraps a sentinel error with additional context. It preserves
// the sentinel's identity through the error chain so that errors.Is continues
// to work correctly at the call site.
type DomainError struct {
	// Sentinel is the stable, matchable root error (one of the vars above).
	Sentinel error
	// Detail provides human-readable context without altering the sentinel identity.
	Detail string
}

// Error implements the error interface, formatting the sentinel and detail together.
func (e *DomainError) Error() string {
	return fmt.Sprintf("%s: %s", e.Sentinel, e.Detail)
}

// Unwrap returns the sentinel so that errors.Is can traverse the chain.
func (e *DomainError) Unwrap() error { return e.Sentinel }

// NewDomainError constructs a DomainError that wraps sentinel with the supplied
// detail string.
//
// Example:
//
//	err := domain.NewDomainError(domain.ErrUserNotFound, "id: 3f1c2a9e-...")
//	errors.Is(err, domain.ErrUserNotFound) // true
func NewDomainError(sentinel error, detail string) *DomainError {
	return &DomainError{Sentinel: sentinel, Detail: detail}
}
