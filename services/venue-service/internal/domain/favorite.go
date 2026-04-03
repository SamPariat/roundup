package domain

import "time"

// favorite.go defines the types and command structs for the favorites feature.
// SavedVenue is the persisted record; AddFavoriteCommand and RemoveFavoriteCommand
// carry the input data for the corresponding use-case operations.

// SavedVenue is a venue that a user has added to their favorites within a squad.
// It is persisted in the database and returned by the FavoriteUseCase.
type SavedVenue struct {
	// ID is the database-assigned primary key.
	ID int64
	// SquadID is the squad the favorite belongs to.
	SquadID string
	// UserID is the user who saved the venue.
	UserID string
	// PlaceID is the provider-assigned venue identifier (e.g., Google Place ID).
	PlaceID string
	// Name is the display name of the venue at the time it was saved.
	Name string
	// SavedAt is the UTC timestamp when the user added the venue to their favorites.
	SavedAt time.Time
}

// AddFavoriteCommand carries the data needed to save a venue as a favorite.
type AddFavoriteCommand struct {
	// SquadID is the squad the favorite belongs to.
	SquadID string
	// UserID is the user saving the favorite.
	UserID string
	// PlaceID is the provider-assigned venue identifier.
	PlaceID string
	// Name is the display name of the venue at the time it was saved.
	Name string
}

// RemoveFavoriteCommand carries the data needed to remove a saved favorite.
type RemoveFavoriteCommand struct {
	// SquadID is the squad the favorite belongs to.
	SquadID string
	// UserID is the user removing the favorite.
	UserID string
	// PlaceID is the provider-assigned venue identifier.
	PlaceID string
}
