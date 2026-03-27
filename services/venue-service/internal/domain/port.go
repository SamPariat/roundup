package domain

import "context"

// port.go declares the three driven-port interfaces the application layer depends on:
// PlaceProvider (external search API), VenueCache (caching layer), and VenueRepository
// (persistence layer). Implementations live in internal/adapters/.

// PlaceProvider abstracts the external place-search API (Google Maps, Foursquare, etc.).
// Swap the provider by writing a new adapter; domain and application layers stay untouched.
type PlaceProvider interface {
	SearchNearby(ctx context.Context, params SearchParams) ([]Venue, error)
	GetDetail(ctx context.Context, placeID string) (*VenueDetail, error)
}

// VenueCache abstracts the caching layer (Redis, in-memory, etc.).
// All cache operations are best-effort — callers ignore write errors.
type VenueCache interface {
	GetNearby(ctx context.Context, key string) ([]Venue, error)
	SetNearby(ctx context.Context, key string, venues []Venue) error
	GetDetail(ctx context.Context, placeID string) (*VenueDetail, error)
	SetDetail(ctx context.Context, placeID string, detail *VenueDetail) error
	InvalidateDetail(ctx context.Context, placeID string) error
}

// VenueRepository abstracts the persistence layer (Postgres, etc.).
type VenueRepository interface {
	SaveFavourite(ctx context.Context) error
	DeleteFavourite(ctx context.Context) error
	ListFavourites(ctx context.Context, squadID string) ([]SavedVenue, error)
	IsFavourite(ctx context.Context, squadID, userID, placeID string) (bool, error)
	RecordVisit(ctx context.Context, cmd RecordVisitCommand) error
	GetVisitHistory(ctx context.Context, squadID string) ([]VisitSummary, error)
}
