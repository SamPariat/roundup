package domain

import "time"

// favourite.go defines the types and command structs for the favourites feature.
// SavedVenue is the persisted record; AddFavouriteCommand and RemoveFavouriteCommand
// carry the input data for the corresponding use-case operations.

// SavedVenue is a venue that a user has added to their favourites within a squad.
// It is persisted in the database and returned by the FavouriteUseCase.
type SavedVenue struct {
	// ID is the database-assigned primary key.
	ID int64
	// SquadID is the squad the favourite belongs to.
	SquadID string
	// UserID is the user who saved the venue.
	UserID string
	// PlaceID is the provider-assigned venue identifier (e.g. Google Place ID).
	PlaceID string
	// Name is the display name of the venue at the time it was saved.
	Name string
	// SavedAt is the UTC timestamp when the user added the venue to their favourites.
	SavedAt time.Time
}
