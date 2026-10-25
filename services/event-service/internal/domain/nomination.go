package domain

import "time"

// Place is a venue's display fields, copied from venue-service when it is
// suggested, so the event screens need no provider call.
type Place struct {
	PlaceID     string
	Name        string
	Address     string
	Lat         float64
	Lng         float64
	Rating      float32
	PriceLevel  int
	PhotoURL    string
	PhotoWidth  int
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
	EventID string
	UserID  string
	PlaceID string
}
