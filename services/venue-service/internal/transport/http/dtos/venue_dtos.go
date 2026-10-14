package dtos

import "time"

// SearchRequest carries the query parameters for a nearby venue search.
type SearchRequest struct {
	// Lat is the latitude of the search centre, -90 to 90.
	Lat float64 `query:"lat"`
	// Lng is the longitude of the search centre, -180 to 180.
	Lng float64 `query:"lng"`
	// Radius is the search radius in metres, 1–50000. Required.
	Radius int `query:"radius"`
	// Query is an optional keyword, e.g. "rooftop bar".
	Query string `query:"query"`
	// Type is an optional Google place type, e.g. "bar" or "restaurant".
	Type string `query:"type"`
}

// PhotoRefResponse is the JSON representation of a provider photo reference.
type PhotoRefResponse struct {
	// Reference is the opaque Google photo reference used to fetch the image.
	Reference string `json:"reference"`
	// Width is the original image width in pixels.
	Width int `json:"width"`
	// Height is the original image height in pixels.
	Height int `json:"height"`
}

// ReviewResponse is the JSON representation of a single user review.
type ReviewResponse struct {
	// AuthorName is the reviewer's display name.
	AuthorName string `json:"authorName"`
	// Rating is the reviewer's score, 1–5.
	Rating float32 `json:"rating"`
	// Text is the review body.
	Text string `json:"text"`
	// PublishedAt is when the review was posted, as an RFC 3339 timestamp.
	PublishedAt time.Time `json:"publishedAt"`
}

// VenueResponse is the JSON representation of a domain.Venue.
// Returned as elements of SearchResponse and embedded in VenueDetailResponse.
type VenueResponse struct {
	// PlaceID is the Google Place ID; pass it to GET /api/v1/venues/:placeID.
	PlaceID string `json:"placeID"`
	// Name is the venue's display name.
	Name string `json:"name"`
	// Address is the formatted street address.
	Address string `json:"address"`
	// Latitude is the venue's latitude.
	Latitude float64 `json:"latitude"`
	// Longitude is the venue's longitude.
	Longitude float64 `json:"longitude"`
	// AverageRating is Google's aggregate rating, 1–5; 0 when unrated.
	AverageRating float32 `json:"averageRating"`
	// PriceLevel is 0 (free) to 4 (very expensive).
	PriceLevel int `json:"priceLevel"`
	// Types are Google's category tags, e.g. ["bar", "restaurant"].
	Types []string `json:"types"`
	// IsOpen is whether the venue is open now; null when Google does not say.
	IsOpen *bool `json:"isOpen"`
	// PhotoRefs are references for fetching the venue's photos.
	PhotoRefs []PhotoRefResponse `json:"photoRefs"`
}

// VenueDetailResponse is the JSON representation of a domain.VenueDetail.
// The embedded VenueResponse fields appear at the top level of the JSON object.
type VenueDetailResponse struct {
	VenueResponse
	// PhoneNumber is the venue's phone number as Google formats it; "" if unknown.
	PhoneNumber string `json:"phoneNumber"`
	// Website is the venue's official URL; "" if unknown.
	Website string `json:"website"`
	// OpeningHours has one display line per weekday, e.g. "Monday: 12:00 – 11:00 PM".
	OpeningHours []string `json:"openingHours"`
	// Reviews is a sample of recent reviews from Google (at most 5).
	Reviews []ReviewResponse `json:"reviews"`
	// EditorialSummary is Google's short description; "" if none.
	EditorialSummary string `json:"editorialSummary"`
}

// SearchResponse is the body of the search endpoint.
type SearchResponse struct {
	// Venues holds the matches; empty, never null, when there are none.
	Venues []VenueResponse `json:"venues"`
}
