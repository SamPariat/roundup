package application

import (
	"context"

	"github.com/SamPariat/roundup/services/venue-service/internal/domain"
)

// HistoryUseCase handles the visit history feature for a squad.
// It delegates all persistence operations to the VenueRepository port.
type HistoryUseCase struct {
	repo domain.VenueRepository
}

// NewHistoryUseCase constructs a HistoryUseCase with the given repository.
func NewHistoryUseCase(repo domain.VenueRepository) *HistoryUseCase {
	return &HistoryUseCase{repo: repo}
}

// RecordVisit stores a new venue visit for the squad.
// Returns domain.ErrVisitAlreadyRecorded if this event's visit is already recorded.
func (u *HistoryUseCase) RecordVisit(ctx context.Context, cmd domain.RecordVisitCommand) error {
	return u.repo.RecordVisit(ctx, cmd)
}

// GetVisitHistory returns aggregated visit summaries for all venues visited by the squad.
func (u *HistoryUseCase) GetVisitHistory(ctx context.Context, squadID string) ([]domain.VisitSummary, error) {
	return u.repo.GetVisitHistory(ctx, squadID)
}
