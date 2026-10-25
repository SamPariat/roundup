package dtos

import "time"

// RecordVisitRequest carries the JSON body for recording a venue visit.
type RecordVisitRequest struct {
	// EventID is the event during which the visit happened.
	EventID string `json:"eventID"`

	// PlaceID is the Google Place ID of the venue.
	PlaceID string `json:"placeID"`

	// VisitedAt is when the squad arrived, as an RFC 3339 timestamp.
	VisitedAt time.Time `json:"visitedAt"`

	// AvgSpendInPaise is the average spend per person in paise (₹1 = 100).
	// Omit or send 0 when unknown.
	AvgSpendInPaise int64 `json:"avgSpendInPaise"`
}

// VisitSummaryResponse is the JSON representation of a domain.VisitSummary.
type VisitSummaryResponse struct {
	// PlaceID is the Google Place ID of the venue.
	PlaceID string `json:"placeID"`

	// Name is the venue's display name at the latest visit.
	Name string `json:"name"`

	// Address is the formatted street address at the latest visit.
	Address string `json:"address"`

	// Photo can be loaded directly; null when the venue has none.
	Photo *PhotoResponse `json:"photo"`

	// VisitCount is how many visits the squad has recorded here.
	VisitCount int64 `json:"visitCount"`

	// LastVisitedAt is the most recent visit, as an RFC 3339 timestamp.
	LastVisitedAt time.Time `json:"lastVisitedAt"`

	// AvgSpendInPaise averages the visits with a known spend per person, in
	// paise (₹1 = 100); null when none were known.
	AvgSpendInPaise *int64 `json:"avgSpendInPaise"`
}

// VisitHistoryResponse is the body of the history endpoint.
type VisitHistoryResponse struct {
	// History holds one entry per venue, most recently visited first.
	// Empty, never null, when there are none.
	History []VisitSummaryResponse `json:"history"`
}
