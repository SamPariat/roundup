package domain

import "time"

// RecordVisitCommand carries the data needed to record a new venue visit.
// A squad can record each venue only once per event. The venue's display
// fields are looked up by the server, never sent by the client.
type RecordVisitCommand struct {
	// SquadID identifies the squad that made the visit.
	SquadID string

	// UserID is the member recording the visit. Used for the membership
	// check; not stored.
	UserID string

	// EventID identifies the event during which the visit occurred.
	EventID string

	// PlaceID is the provider-assigned identifier for the venue.
	PlaceID string

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

	// Name is the display name of the venue at the latest visit.
	Name string

	// Address is the formatted street address at the latest visit.
	Address string

	// Photo is the venue's photo at the latest visit; an empty Reference means none.
	Photo PhotoRef

	// VisitCount is the total number of times the squad has visited this venue.
	VisitCount int64

	// LastVisitedAt is the most recent visit time across all events at this venue.
	LastVisitedAt time.Time

	// AvgSpendInPaise averages the visits with a known spend; 0 means none were known.
	AvgSpendInPaise int64
}
