package domain

import "github.com/SamPariat/roundup/shared/domainerr"

// Sentinel errors are used as stable identities for errors.Is comparisons.
// Callers should match against these values, never against error strings.
// Adapters add context with domainerr.New, which keeps errors.Is working.
var (
	// ErrVenueNotFound is returned when a venue lookup finds no matching record.
	ErrVenueNotFound = domainerr.Sentinel("VENUE_NOT_FOUND", "venue not found")

	// ErrProviderUnavailable is returned when the external place provider
	// (e.g., Google Maps) cannot be reached or returns an unrecoverable error.
	ErrProviderUnavailable = domainerr.Sentinel("PROVIDER_UNAVAILABLE", "place provider unavailable")

	// ErrAlreadySaved is returned when a user attempts to save a venue they
	// have already added to their favorites.
	ErrAlreadySaved = domainerr.Sentinel("ALREADY_SAVED", "venue already saved")

	// ErrNotSaved is returned when a user attempts to remove a venue that is
	// not present in their favorites.
	ErrNotSaved = domainerr.Sentinel("NOT_SAVED", "venue not saved by this user")

	// ErrInvalidSearch is returned when nearby-search parameters are missing
	// or out of range. Callers must fix the request; retrying will not help.
	ErrInvalidSearch = domainerr.Sentinel("INVALID_SEARCH", "invalid search parameters")

	// ErrVisitAlreadyRecorded is returned when the squad already has a visit
	// recorded for this venue during this event.
	ErrVisitAlreadyRecorded = domainerr.Sentinel("VISIT_ALREADY_RECORDED", "visit already recorded for this event")

	// ErrSquadNotFound is returned when the squad does not exist or the caller
	// is not a member. user-service treats the two the same, and so do we.
	ErrSquadNotFound = domainerr.Sentinel("SQUAD_NOT_FOUND", "squad not found")

	// ErrUserServiceUnavailable is returned when user-service cannot answer a
	// membership check.
	ErrUserServiceUnavailable = domainerr.Sentinel("USER_SERVICE_UNAVAILABLE", "user service unavailable")
)
