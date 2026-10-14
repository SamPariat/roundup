package dtos

import "time"

// RecordVisitRequest carries the JSON body for recording a venue visit.
type RecordVisitRequest struct {
	// EventID is the event during which the visit happened.
	EventID string `json:"eventID"`
	// PlaceID is the Google Place ID of the venue.
	PlaceID string `json:"placeID"`
	// Name is the venue's display name.
	Name string `json:"name"`
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
	// Name is the venue's display name.
	Name string `json:"name"`
	// VisitCount is how many visits the squad has recorded here.
	VisitCount int64 `json:"visitCount"`
	// LastVisitedAt is the most recent visit, as an RFC 3339 timestamp.
	LastVisitedAt time.Time `json:"lastVisitedAt"`
}

// VisitHistoryResponse is the body of the history endpoint.
type VisitHistoryResponse struct {
	// History holds one entry per venue, most recently visited first.
	// Empty, never null, when there are none.
	History []VisitSummaryResponse `json:"history"`
}
