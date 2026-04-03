package dtos

import "time"

// favorite_dtos.go defines request and response structs for the favorite endpoints.

// AddFavoriteRequest carries the JSON body for saving a venue as a favorite.
type AddFavoriteRequest struct {
	UserID  string `json:"userID"`
	PlaceID string `json:"placeID"`
	Name    string `json:"name"`
}

// SavedVenueResponse is the JSON representation of a domain.SavedVenue.
type SavedVenueResponse struct {
	ID      int64     `json:"id"`
	PlaceID string    `json:"placeID"`
	Name    string    `json:"name"`
	UserID  string    `json:"userID"`
	SavedAt time.Time `json:"savedAt"`
}

// FavoritesResponse wraps a slice of SavedVenueResponse for the list endpoint.
type FavoritesResponse struct {
	Favorites []SavedVenueResponse `json:"favorites"`
}

// IsFavoriteResponse carries the result of the IsFavorite check.
type IsFavoriteResponse struct {
	IsFavorite bool `json:"isFavorite"`
}
