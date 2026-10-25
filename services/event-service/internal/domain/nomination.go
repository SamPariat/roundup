package domain

import "time"

// Place is a venue's display fields, copied from venue-service when it is
// suggested, so the event screens need no provider call.
type Place struct {
	// PlaceID is the Google Place ID.
	PlaceID string

	// Name is the place's display name.
	Name string

	// Address is the formatted street address.
	Address string

	// Lat is the WGS-84 latitude.
	Lat float64

	// Lng is the WGS-84 longitude.
	Lng float64

	// Rating is the average rating, 1–5; 0 when unrated.
	Rating float32

	// PriceLevel is 0–4 (0 = free or unknown, 4 = very expensive).
	PriceLevel int

	// PhotoURL is venue-service's URL for the first photo; "" if it had none.
	PhotoURL string

	// PhotoWidth is the photo's original width in pixels.
	PhotoWidth int

	// PhotoHeight is the photo's original height in pixels.
	PhotoHeight int
}

// Nomination is a place suggested for an event.
type Nomination struct {
	Place

	// NominatedBy is the member who suggested it. They may since have left.
	NominatedBy string

	// NominatedAt breaks ties: the earliest suggestion wins a tied vote.
	NominatedAt time.Time
}

// Vote is one member's pick for one event. Votes are open: everyone sees
// who voted for what.
type Vote struct {
	// EventID is the event voted in.
	EventID string

	// UserID is the member who voted. They may since have left.
	UserID string

	// PlaceID is the suggested place they picked.
	PlaceID string
}
