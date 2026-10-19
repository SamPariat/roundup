package domain

import (
	"fmt"
	"math"
	"time"

	"github.com/SamPariat/roundup/shared/domainerr"
)

// PhotoRef holds a provider photo reference alongside its original dimensions.
type PhotoRef struct {
	// Reference is the opaque provider token used to fetch the photo.
	Reference string

	// Width is the original image width in pixels.
	Width int

	// Height is the original image height in pixels.
	Height int
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

// Search limits enforced by SearchParams.Validate.
const (
	// MaxLatitude bounds latitude to [-MaxLatitude, MaxLatitude] degrees.
	MaxLatitude = 90

	// MaxLongitude bounds longitude to [-MaxLongitude, MaxLongitude] degrees.
	MaxLongitude = 180

	// MaxRadiusInMeters is the largest radius the place provider accepts.
	MaxRadiusInMeters = 50000
)

// SearchParams carries the filters for a nearby venue search.
type SearchParams struct {
	// Latitude is the WGS-84 latitude of the search origin.
	Latitude float64

	// Longitude is the WGS-84 longitude of the search origin.
	Longitude float64

	// RadiusInMeters is the search radius, 1–MaxRadiusInMeters.
	RadiusInMeters int

	// Query is an optional free-text keyword (e.g. "rooftop bar").
	Query string

	// Type is an optional provider category filter (e.g. "bar", "restaurant").
	Type string
}

// Validate reports whether p describes a searchable area. It returns a
// DomainError wrapping ErrInvalidSearch, with a detail naming the bad field,
// so the transport layer can map it to 400 without calling the provider.
func (p SearchParams) Validate() error {
	// NaN fails every comparison, so it would slip past the range checks below.
	if math.IsNaN(p.Latitude) || math.IsNaN(p.Longitude) {
		return domainerr.New(ErrInvalidSearch, "latitude and longitude must be numbers")
	}

	if p.Latitude < -MaxLatitude || p.Latitude > MaxLatitude {
		return domainerr.New(ErrInvalidSearch, fmt.Sprintf("latitude must be between -%d and %d", MaxLatitude, MaxLatitude))
	}

	if p.Longitude < -MaxLongitude || p.Longitude > MaxLongitude {
		return domainerr.New(ErrInvalidSearch, fmt.Sprintf("longitude must be between -%d and %d", MaxLongitude, MaxLongitude))
	}

	if p.RadiusInMeters <= 0 || p.RadiusInMeters > MaxRadiusInMeters {
		return domainerr.New(ErrInvalidSearch, fmt.Sprintf("radius must be between 1 and %d", MaxRadiusInMeters))
	}

	return nil
}
