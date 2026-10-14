package dtos

import "time"

// AddFavoriteRequest carries the JSON body for saving a venue as a favorite.
// The user is taken from the X-User-ID header, never from the body.
type AddFavoriteRequest struct {
	// PlaceID is the Google Place ID of the venue.
	PlaceID string `json:"placeID"`
	// Name is the venue's display name, stored so lists render without a provider call.
	Name string `json:"name"`
}

// SavedVenueResponse is the JSON representation of a domain.SavedVenue.
type SavedVenueResponse struct {
	// ID is the server-assigned ID of the favorite.
	ID int64 `json:"id"`
	// PlaceID is the Google Place ID of the venue.
	PlaceID string `json:"placeID"`
	// Name is the venue's display name at the time it was saved.
	Name string `json:"name"`
	// UserID is the squad member who saved it.
	UserID string `json:"userID"`
	// SavedAt is when it was saved, as an RFC 3339 timestamp.
	SavedAt time.Time `json:"savedAt"`
}

// FavoritesResponse is the body of the list endpoint.
type FavoritesResponse struct {
	// Favorites holds every member's favorites in the squad, newest first.
	// Empty, never null, when there are none.
	Favorites []SavedVenueResponse `json:"favorites"`
}

// IsFavoriteResponse is the body of the check endpoint.
type IsFavoriteResponse struct {
	// IsFavorite reports whether the caller has saved the venue in this squad.
	IsFavorite bool `json:"isFavorite"`
}
