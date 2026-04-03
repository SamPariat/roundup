package dtos

import "time"

// venue_dtos.go defines the request and response structs for venue-related HTTP endpoints.
// Domain types are mapped into these structs at the handler boundary — JSON and query tags
// never appear in the domain layer.

// SearchRequest carries the query parameters for a nearby venue search.
type SearchRequest struct {
	Lat    float64 `query:"lat"`
	Lng    float64 `query:"lng"`
	Radius int     `query:"radius"`
	Query  string  `query:"query"`
	Type   string  `query:"type"`
}

// PhotoRefResponse is the JSON representation of a provider photo reference.
type PhotoRefResponse struct {
	Reference string `json:"reference"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
}

// ReviewResponse is the JSON representation of a single user review.
type ReviewResponse struct {
	AuthorName  string    `json:"author_name"`
	Rating      float32   `json:"rating"`
	Text        string    `json:"text"`
	PublishedAt time.Time `json:"published_at"`
}

// VenueResponse is the JSON representation of a domain.Venue.
// Returned as elements of SearchResponse and embedded in VenueDetailResponse.
type VenueResponse struct {
	PlaceID       string             `json:"placeID"`
	Name          string             `json:"name"`
	Address       string             `json:"address"`
	Latitude      float64            `json:"latitude"`
	Longitude     float64            `json:"longitude"`
	AverageRating float32            `json:"averageRating"`
	PriceLevel    int                `json:"priceLevel"`
	Types         []string           `json:"types"`
	IsOpen        *bool              `json:"isOpen"`
	PhotoRefs     []PhotoRefResponse `json:"photoRefs"`
}

// VenueDetailResponse is the JSON representation of a domain.VenueDetail.
// Embeds VenueResponse and adds the full detail fields.
type VenueDetailResponse struct {
	VenueResponse
	PhoneNumber      string           `json:"phoneNumber"`
	Website          string           `json:"website"`
	OpeningHours     []string         `json:"openingHours"`
	Reviews          []ReviewResponse `json:"reviews"`
	EditorialSummary string           `json:"editorialSummary"`
}

// SearchResponse wraps a slice of VenueResponse for the search endpoint.
type SearchResponse struct {
	Venues []VenueResponse `json:"venues"`
}
