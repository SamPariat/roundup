package domain

import "fmt"

// errors.go defines the sentinel error variables, the DomainError wrapper type, and its
// constructor. All cross-layer error matching uses errors.Is against the sentinels here —
// callers must never compare against error strings.

// Sentinel errors are used as stable identities for errors.Is comparisons.
// Callers should match against these values, never against error strings.
var (
	// ErrVenueNotFound is returned when a venue lookup finds no matching record.
	ErrVenueNotFound = fmt.Errorf("venue not found")

	// ErrProviderUnavailable is returned when the external place provider
	// (e.g., Google Maps) cannot be reached or returns an unrecoverable error.
	ErrProviderUnavailable = fmt.Errorf("place provider unavailable")

	// ErrAlreadySaved is returned when a user attempts to save a venue they
	// have already added to their favorites.
	ErrAlreadySaved = fmt.Errorf("venue already saved")

	// ErrNotSaved is returned when a user attempts to remove a venue that is
	// not present in their favorites.
	ErrNotSaved = fmt.Errorf("venue not saved by this user")
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
//	err := domain.NewDomainError(domain.ErrVenueNotFound, "placeId: ChIJN1t2eZgVkFQR")
//	errors.Is(err, domain.ErrVenueNotFound) // true
func NewDomainError(sentinel error, detail string) *DomainError {
	return &DomainError{Sentinel: sentinel, Detail: detail}
}
