package application

import (
	"context"

	"github.com/SamPariatIL/roundup/services/venue-service/internal/domain"
)

// history_usecase.go implements HistoryUseCase, which handles recording venue visits
// and fetching aggregated visit history for a squad. All operations are delegated directly
// to VenueRepository — no caching or provider calls are made here.

// HistoryUseCase handles the visit history feature for a squad.
// It delegates all persistence operations to the VenueRepository port.
type HistoryUseCase struct {
	repo domain.VenueRepository
}

// NewHistoryUseCase constructs a HistoryUseCase with the given repository.
func NewHistoryUseCase(repo domain.VenueRepository) *HistoryUseCase {
	return &HistoryUseCase{repo: repo}
}

// RecordVisit persists as a new venue visit for the squad.
// Typically called when an event is confirmed.
func (u *HistoryUseCase) RecordVisit(ctx context.Context, cmd domain.RecordVisitCommand) error {
	return u.repo.RecordVisit(ctx, cmd)
}

// GetVisitHistory returns aggregated visit summaries for all venues visited by the squad.
func (u *HistoryUseCase) GetVisitHistory(ctx context.Context, squadID string) ([]domain.VisitSummary, error) {
	return u.repo.GetVisitHistory(ctx, squadID)
}
