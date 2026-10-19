package domain

import "context"

// PlaceProvider abstracts the external place-search API (Google Maps, Foursquare, etc.).
// Swap the provider by writing a new adapter; domain and application layers stay untouched.
type PlaceProvider interface {
	// SearchNearby returns venues matching params. An area with no matches
	// returns an empty slice, not an error.
	SearchNearby(ctx context.Context, params SearchParams) ([]Venue, error)

	// GetDetail returns full information for one venue. Returns ErrVenueNotFound
	// for an unknown or malformed placeID, and ErrProviderUnavailable when the
	// provider rejects or cannot serve the request.
	GetDetail(ctx context.Context, placeID string) (*VenueDetail, error)
}

// VenueCache abstracts the caching layer (Redis, in-memory, etc.).
// All cache operations are best-effort — callers ignore their errors.
// Getters return nil, nil on a miss.
type VenueCache interface {
	// GetNearby returns the cached search results stored under key.
	GetNearby(ctx context.Context, key string) ([]Venue, error)

	// SetNearby stores search results under key.
	SetNearby(ctx context.Context, key string, venues []Venue) error

	// GetDetail returns the cached detail for placeID.
	GetDetail(ctx context.Context, placeID string) (*VenueDetail, error)

	// SetDetail stores the detail for placeID.
	SetDetail(ctx context.Context, placeID string, detail *VenueDetail) error
}

// VenueRepository abstracts the persistence layer (Postgres, etc.).
type VenueRepository interface {
	// AddFavorite returns ErrAlreadySaved if the user already saved this venue in the squad.
	AddFavorite(ctx context.Context, cmd AddFavoriteCommand) error

	// RemoveFavorite returns ErrNotSaved if the user has not saved this venue in the squad.
	RemoveFavorite(ctx context.Context, cmd RemoveFavoriteCommand) error

	// ListFavorites returns every member's favorites in the squad, newest first.
	ListFavorites(ctx context.Context, squadID string) ([]SavedVenue, error)

	// IsFavorite reports whether the user has saved this venue in the squad.
	IsFavorite(ctx context.Context, squadID, userID, placeID string) (bool, error)

	// RecordVisit stores one visit. Returns ErrVisitAlreadyRecorded if the squad
	// already recorded this venue for this event.
	RecordVisit(ctx context.Context, cmd RecordVisitCommand) error

	// GetVisitHistory returns one summary per venue the squad has visited,
	// most recently visited first.
	GetVisitHistory(ctx context.Context, squadID string) ([]VisitSummary, error)
}
