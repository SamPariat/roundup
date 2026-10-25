package domain

import (
	"errors"

	"github.com/SamPariat/roundup/shared/userclient"
)

// Sentinel errors are used as stable identities for errors.Is comparisons.
// Callers should match against these values, never against error strings.
// Adapters add context with domainerr.New, which keeps errors.Is working.
var (
	// ErrVenueNotFound is returned when a venue lookup finds no matching record.
	ErrVenueNotFound = errors.New("venue not found")

	// ErrPhotoNotFound is returned when the provider has no photo for a reference.
	ErrPhotoNotFound = errors.New("photo not found")

	// ErrProviderUnavailable is returned when the external place provider
	// (e.g., Google Maps) cannot be reached or returns an unrecoverable error.
	ErrProviderUnavailable = errors.New("place provider unavailable")

	// ErrAlreadySaved is returned when a user attempts to save a venue they
	// have already added to their favorites.
	ErrAlreadySaved = errors.New("venue already saved")

	// ErrNotSaved is returned when a user attempts to remove a venue that is
	// not present in their favorites.
	ErrNotSaved = errors.New("venue not saved by this user")

	// ErrInvalidSearch is returned when nearby-search parameters are missing
	// or out of range. Callers must fix the request; retrying will not help.
	ErrInvalidSearch = errors.New("invalid search parameters")

	// ErrVisitAlreadyRecorded is returned when the squad already has a visit
	// recorded for this venue during this event.
	ErrVisitAlreadyRecorded = errors.New("visit already recorded for this event")

	// ErrSquadNotFound is returned when the squad does not exist or the caller
	// is not a member. user-service treats the two the same, and so do we.
	ErrSquadNotFound = userclient.ErrSquadNotFound

	// ErrUserServiceUnavailable is returned when user-service cannot answer a
	// membership check.
	ErrUserServiceUnavailable = userclient.ErrUserServiceUnavailable
)
