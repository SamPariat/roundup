package domain

import "fmt"

// Sentinel errors are used as stable identities for errors.Is comparisons.
// Callers should match against these values, never against error strings.
var (
	// ErrInvalidPhone is returned when a phone number is not in E.164 format.
	ErrInvalidPhone = fmt.Errorf("phone must be in E.164 format, e.g. +919876543210")

	// ErrTooManyRequests is returned when a phone has requested too many codes
	// within the rate-limit window.
	ErrTooManyRequests = fmt.Errorf("too many code requests, try again later")

	// ErrInvalidCode is returned when a code is wrong, expired, already used,
	// or was never requested. One error for all cases avoids telling an
	// attacker which part was wrong.
	ErrInvalidCode = fmt.Errorf("invalid or expired code")

	// ErrTooManyAttempts is returned when the current code has been guessed
	// wrongly too many times. The user must request a new code.
	ErrTooManyAttempts = fmt.Errorf("too many wrong attempts, request a new code")

	// ErrUserServiceUnavailable is returned when user-service cannot be reached
	// or answers unexpectedly during login. The code is left unused so the
	// user can retry.
	ErrUserServiceUnavailable = fmt.Errorf("user service unavailable")

	// ErrInvalidRefreshToken is returned for a refresh token that is unknown,
	// expired, already used, or belongs to a revoked session.
	ErrInvalidRefreshToken = fmt.Errorf("invalid refresh token")
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
func NewDomainError(sentinel error, detail string) *DomainError {
	return &DomainError{Sentinel: sentinel, Detail: detail}
}
