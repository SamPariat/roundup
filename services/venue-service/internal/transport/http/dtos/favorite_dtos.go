package dtos

import "time"

// AddFavoriteRequest carries the JSON body for saving a venue as a favorite.
// The user is taken from the X-User-ID header, never from the body, and the
// venue's display fields are looked up by the server.
type AddFavoriteRequest struct {
	// PlaceID is the Google Place ID of the venue.
	PlaceID string `json:"placeID"`
}

// MemberResponse is a squad member as shown next to a favorite.
type MemberResponse struct {
	// UserID is the member's user ID.
	UserID string `json:"userID"`

	// Name is "" until the member completes profile setup.
	Name string `json:"name"`

	// AvatarURL is "" when no avatar is set.
	AvatarURL string `json:"avatarURL"`
}

// FavoriteResponse is one venue in a squad's favorites.
type FavoriteResponse struct {
	// PlaceID is the Google Place ID of the venue.
	PlaceID string `json:"placeID"`

	// Name is the venue's display name at the latest save.
	Name string `json:"name"`

	// Address is the formatted street address at the latest save.
	Address string `json:"address"`

	// AverageRating is Google's aggregate rating, 1–5; 0 when unrated.
	AverageRating float32 `json:"averageRating"`

	// RatingCount is how many ratings AverageRating is based on.
	RatingCount int `json:"ratingCount"`

	// PriceLevel is 0 (free) to 4 (very expensive).
	PriceLevel int `json:"priceLevel"`

	// Photo can be loaded directly; null when the venue has none.
	Photo *PhotoResponse `json:"photo"`

	// SavedBy lists the members who saved it, most recent first. Never null.
	SavedBy []MemberResponse `json:"savedBy"`

	// SavedByMe reports whether the caller saved it.
	SavedByMe bool `json:"savedByMe"`

	// LastSavedAt is the most recent save, as an RFC 3339 timestamp.
	LastSavedAt time.Time `json:"lastSavedAt"`
}

// FavoritesResponse is the body of the list endpoint.
type FavoritesResponse struct {
	// Favorites holds one entry per venue, most recently saved first.
	// Empty, never null, when there are none.
	Favorites []FavoriteResponse `json:"favorites"`

	// NextCursor fetches the next page as ?cursor=; null on the last page.
	NextCursor *string `json:"nextCursor"`
}
