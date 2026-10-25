package domain

import (
	"context"

	"github.com/SamPariat/roundup/shared/pagination"
	"github.com/SamPariat/roundup/shared/userclient"
)

// PlaceProvider abstracts the external place-search API (Google Maps, Foursquare, etc.).
// Swap the provider by writing a new adapter; domain and application layers stay untouched.
type PlaceProvider interface {
	// SearchNearby returns one page of venues matching params. An area with
	// no matches returns an empty page, not an error.
	SearchNearby(ctx context.Context, params SearchParams) (*SearchResult, error)

	// GetDetail returns full information for one venue. Returns ErrVenueNotFound
	// for an unknown or malformed placeID, and ErrProviderUnavailable when the
	// provider rejects or cannot serve the request.
	GetDetail(ctx context.Context, placeID string) (*VenueDetail, error)

	// Photo streams a venue photo at most maxWidth pixels wide. Returns
	// ErrPhotoNotFound for an unknown reference.
	Photo(ctx context.Context, reference string, maxWidth int) (*Photo, error)
}

// VenueCache abstracts the caching layer (Redis, in-memory, etc.).
// All cache operations are best-effort — callers ignore their errors.
// Getters return nil, nil on a miss.
type VenueCache interface {
	// GetNearby returns the cached search page stored under key.
	GetNearby(ctx context.Context, key string) (*SearchResult, error)

	// SetNearby stores a search page under key.
	SetNearby(ctx context.Context, key string, result *SearchResult) error

	// GetDetail returns the cached detail for placeID.
	GetDetail(ctx context.Context, placeID string) (*VenueDetail, error)

	// SetDetail stores the detail for placeID.
	SetDetail(ctx context.Context, placeID string, detail *VenueDetail) error
}

// VenueRepository abstracts the persistence layer (Postgres, etc.).
type VenueRepository interface {
	// AddFavorite stores the favorite with snapshot as its display fields. Returns
	// ErrAlreadySaved if the user already saved this venue in the squad.
	AddFavorite(ctx context.Context, cmd AddFavoriteCommand, snapshot VenueSnapshot) error

	// RemoveFavorite returns ErrNotSaved if the user has not saved this venue in the squad.
	RemoveFavorite(ctx context.Context, cmd RemoveFavoriteCommand) error

	// ListSquadFavorites returns one page of groups, one per venue saved in
	// the squad, most recently saved first, counting only saves by memberIDs,
	// and the next page's cursor ("" on the last). Returns
	// pagination.ErrInvalid for a cursor it did not issue.
	ListSquadFavorites(ctx context.Context, squadID string, memberIDs []string, page pagination.Page) ([]FavoriteGroup, string, error)

	// SavedInSquads returns the IDs of the squads, among squadIDs, in which
	// the user saved this venue.
	SavedInSquads(ctx context.Context, userID, placeID string, squadIDs []string) ([]string, error)

	// RecordVisit stores one visit with snapshot as its display fields. Returns
	// ErrVisitAlreadyRecorded if the squad already recorded this venue for this event.
	RecordVisit(ctx context.Context, cmd RecordVisitCommand, snapshot VenueSnapshot) error

	// GetVisitHistory returns one page of summaries, one per venue the squad
	// has visited, most recently visited first, and the next page's cursor
	// ("" on the last). Returns pagination.ErrInvalid for a cursor it did not
	// issue.
	GetVisitHistory(ctx context.Context, squadID string, page pagination.Page) ([]VisitSummary, string, error)
}

// SquadMember is a squad member as user-service describes them.
type SquadMember = userclient.SquadMember

// Squad is one of the caller's squads as user-service describes it.
type Squad = userclient.Squad

// SquadDirectory answers who belongs to which squad. Squads live in
// user-service; implemented by userclient.UserServiceClient.
type SquadDirectory interface {
	// Members returns squadID's members if userID is one of them,
	// ErrSquadNotFound if they are not (or the squad does not exist), and
	// ErrUserServiceUnavailable if user-service cannot answer.
	Members(ctx context.Context, squadID, userID string) ([]SquadMember, error)

	// Squads returns every squad userID belongs to, and
	// ErrUserServiceUnavailable if user-service cannot answer.
	Squads(ctx context.Context, userID string) ([]Squad, error)
}
