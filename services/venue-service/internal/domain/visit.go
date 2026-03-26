package domain

import "time"

// VenueVisit is a record of a squad visiting a venue during an event.
// Visits are recorded automatically when an event is confirmed.
type VenueVisit struct {
	// ID is the database-assigned primary key.
	ID int64
	// SquadID identifies the squad that made the visit.
	SquadID int64
	// EventID identifies the event during which the visit occurred.
	EventID int64
	// PlaceID is the provider-assigned identifier for the venue.
	PlaceID string
	// VisitedAt is the time the squad arrived at the venue.
	VisitedAt time.Time
	// AvgSpendInPaise is the average spend per person in paise (1 INR = 100 paise).
	AvgSpendInPaise int64
}

// RecordVisitCommand carries the data needed to record a new venue visit.
type RecordVisitCommand struct {
	// SquadID identifies the squad that made the visit.
	SquadID int64
	// EventID identifies the event during which the visit occurred.
	EventID int64
	// PlaceID is the provider-assigned identifier for the venue.
	PlaceID string
	// VisitedAt is the time the squad arrived at the venue.
	VisitedAt time.Time
	// AvgSpendInPaise is the average spend per person in paise (1 INR = 100 paise).
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
	VisitCount int
	// LastVisitedAt is the most recent visit time across all events at this venue.
	LastVisitedAt time.Time
}
