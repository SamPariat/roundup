package application

import (
	"context"

	"github.com/SamPariat/roundup/services/venue-service/internal/domain"
)

// HistoryUseCase handles the visit history feature for a squad. Every method
// first checks that the caller is a member of the squad, then delegates to
// the VenueRepository port.
type HistoryUseCase struct {
	visits domain.VenueRepository
	squads domain.SquadDirectory
}

// NewHistoryUseCase constructs a HistoryUseCase with the given repository
// and squad directory.
func NewHistoryUseCase(visits domain.VenueRepository, squads domain.SquadDirectory) *HistoryUseCase {
	return &HistoryUseCase{visits: visits, squads: squads}
}

// RecordVisit stores a new venue visit for the squad.
// Returns domain.ErrSquadNotFound if cmd.UserID is not in the squad and
// domain.ErrVisitAlreadyRecorded if this event's visit is already recorded.
func (u *HistoryUseCase) RecordVisit(ctx context.Context, cmd domain.RecordVisitCommand) error {
	if err := u.squads.EnsureMember(ctx, cmd.SquadID, cmd.UserID); err != nil {
		return err
	}

	return u.visits.RecordVisit(ctx, cmd)
}

// GetVisitHistory returns aggregated visit summaries for all venues visited by the squad.
// Returns domain.ErrSquadNotFound if userID is not in the squad.
func (u *HistoryUseCase) GetVisitHistory(ctx context.Context, squadID, userID string) ([]domain.VisitSummary, error) {
	if err := u.squads.EnsureMember(ctx, squadID, userID); err != nil {
		return nil, err
	}

	return u.visits.GetVisitHistory(ctx, squadID)
}
