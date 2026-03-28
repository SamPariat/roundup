package domain

import "time"

// venue.go defines the core value types for venue search: Review, Venue, VenueDetail,
// and SearchParams. These are the canonical representations used throughout all layers;
// adapters map their provider-specific types into these structs at the boundary.

// PhotoRef holds a provider photo reference alongside its original dimensions.
// Width and Height let callers compute aspect ratios before fetching the image.
type PhotoRef struct {
	Reference string
	Width     int
	Height    int
}

// Review is a single user review returned by the place provider.
type Review struct {
	// AuthorName is the display name of the reviewer.
	AuthorName string
	// Rating is the reviewer's score, normalized to 1–5.
	Rating float32
	// Text is the body of the review.
	Text string
	// PublishedAt is when the review was posted.
	PublishedAt time.Time
}

// Venue is the lightweight representation returned by a nearby search.
// For full details including reviews and opening hours, use VenueDetail.
type Venue struct {
	// PlaceID is the provider-assigned unique identifier (e.g., Google Place ID).
	PlaceID string
	// Name is the display name of the venue.
	Name string
	// Address is the formatted street address.
	Address string
	// Latitude is the WGS-84 latitude of the venue.
	Latitude float64
	// Longitude is the WGS-84 longitude of the venue.
	Longitude float64
	// AverageRating is the provider's aggregate rating, normalized to 1–5.
	AverageRating float32
	// PriceLevel is provider-normalized to 0–4 (0 = free, 4 = very expensive).
	PriceLevel int
	// Types are the provider's category tags (e.g. "bar", "restaurant", "rooftop").
	Types []string
	// IsOpen is the current open/closed status. Nil means the provider did not return it.
	IsOpen *bool
	// PhotoRefs are opaque provider references used to fetch photos.
	PhotoRefs []PhotoRef
}

// VenueDetail extends Venue with the full information returned by a detail lookup.
type VenueDetail struct {
	Venue
	// PhoneNumber is the venue's contact number in international format.
	PhoneNumber string
	// Website is the venue's official URL.
	Website string
	// OpeningHours is a human-readable list of opening periods (e.g. "Monday: 9am – 11pm").
	OpeningHours []string
	// Reviews is a sample of recent user reviews returned by the provider.
	Reviews []Review
	// EditorialSummary is a short provider-supplied description of the venue.
	EditorialSummary string
}

// SearchParams carries the filters for a nearby venue search.
type SearchParams struct {
	// Latitude is the WGS-84 latitude of the search origin.
	Latitude float64
	// Longitude is the WGS-84 longitude of the search origin.
	Longitude float64
	// RadiusInMeters is the search radius. The provider may cap this value.
	RadiusInMeters int
	// Query is a free-text keyword (e.g. "rooftop bar").
	Query string
	// Type is a provider category filter (e.g. "bar", "restaurant").
	Type string
}
