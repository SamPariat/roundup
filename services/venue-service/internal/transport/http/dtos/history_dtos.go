package dtos

import "time"

// history_dtos.go defines request and response structs for the visit history endpoints.

// RecordVisitRequest carries the JSON body for recording a venue visit.
type RecordVisitRequest struct {
	EventID         string    `json:"eventID"`
	PlaceID         string    `json:"placeID"`
	Name            string    `json:"name"`
	VisitedAt       time.Time `json:"visitedAt"`
	AvgSpendInPaise int64     `json:"avgSpendInPaise"`
}

// VisitSummaryResponse is the JSON representation of a domain.VisitSummary.
type VisitSummaryResponse struct {
	PlaceID       string    `json:"placeID"`
	Name          string    `json:"name"`
	VisitCount    int64     `json:"visitCount"`
	LastVisitedAt time.Time `json:"lastVisitedAt"`
}

// VisitHistoryResponse wraps a slice of VisitSummaryResponse for the history endpoint.
type VisitHistoryResponse struct {
	History []VisitSummaryResponse `json:"history"`
}
