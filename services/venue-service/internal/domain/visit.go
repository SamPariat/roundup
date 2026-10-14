package domain

import "time"

// RecordVisitCommand carries the data needed to record a new venue visit.
// A squad can record each venue only once per event.
type RecordVisitCommand struct {
	// SquadID identifies the squad that made the visit.
	SquadID string
	// EventID identifies the event during which the visit occurred.
	EventID string
	// PlaceID is the provider-assigned identifier for the venue.
	PlaceID string
	// Name is the display name of the venue.
	Name string
	// VisitedAt is the time the squad arrived at the venue.
	VisitedAt time.Time
	// AvgSpendInPaise is the average expenditure per person in paise (1 INR = 100 paise).
	// Zero means unknown and is stored as NULL.
	AvgSpendInPaise int64
}

// VisitSummary is an aggregated view of a squad's visits to a single venue.
// Used to populate the venue history screen.
type VisitSummary struct {
	// PlaceID is the provider-assigned identifier for the venue.
	PlaceID string
	// Name is the display name of the venue.
	Name string
	// VisitCount is the total number of times the squad has visited this venue.
	VisitCount int64
	// LastVisitedAt is the most recent visit time across all events at this venue.
	LastVisitedAt time.Time
}
