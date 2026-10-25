package domain

import "time"

// FavoriteGroup is one venue in a squad's favorites as stored: who saved it,
// by user ID. The use case swaps the IDs for member names.
type FavoriteGroup struct {
	// PlaceID is the provider-assigned venue identifier (e.g., Google Place ID).
	PlaceID string

	// VenueSnapshot holds the display fields from the most recent save.
	VenueSnapshot

	// SavedBy holds the IDs of the users who saved it, most recent first.
	SavedBy []string

	// LastSavedAt is when the most recent save happened.
	LastSavedAt time.Time
}

// SquadFavorite is one venue in a squad's favorites, however many members saved it.
type SquadFavorite struct {
	// PlaceID is the provider-assigned venue identifier.
	PlaceID string

	// VenueSnapshot holds the display fields from the most recent save.
	VenueSnapshot

	// SavedBy lists the current squad members who saved it, most recent first.
	SavedBy []SquadMember

	// SavedByMe reports whether the caller is one of them.
	SavedByMe bool

	// LastSavedAt is when the most recent save happened.
	LastSavedAt time.Time
}

// SquadSaved reports whether the caller saved a venue in one of their squads.
type SquadSaved struct {
	// SquadID identifies the squad.
	SquadID string

	// SquadName is the squad's display name.
	SquadName string

	// Saved reports whether the caller saved the venue in this squad.
	Saved bool
}

// AddFavoriteCommand carries the data needed to save a venue as a favorite.
// The venue's display fields are looked up by the server, never sent by the client.
type AddFavoriteCommand struct {
	// SquadID is the squad the favorite belongs to.
	SquadID string

	// UserID is the user saving the favorite.
	UserID string

	// PlaceID is the provider-assigned venue identifier.
	PlaceID string
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
