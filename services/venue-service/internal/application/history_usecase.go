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
	venues VenueLookup
}

// NewHistoryUseCase constructs a HistoryUseCase with the given repository,
// squad directory and venue lookup.
func NewHistoryUseCase(visits domain.VenueRepository, squads domain.SquadDirectory, venues VenueLookup) *HistoryUseCase {
	return &HistoryUseCase{visits: visits, squads: squads, venues: venues}
}

// Record stores a new venue visit for the squad, with a snapshot of the
// venue's display fields. Returns domain.ErrSquadNotFound if cmd.UserID is
// not in the squad, domain.ErrVenueNotFound for an unknown place and
// domain.ErrVisitAlreadyRecorded if this event's visit is already recorded.
func (u *HistoryUseCase) Record(ctx context.Context, cmd domain.RecordVisitCommand) error {
	if _, err := u.squads.Members(ctx, cmd.SquadID, cmd.UserID); err != nil {
		return err
	}

	detail, err := u.venues.GetDetail(ctx, cmd.PlaceID)
	if err != nil {
		return err
	}

	return u.visits.RecordVisit(ctx, cmd, detail.Snapshot())
}

// List returns aggregated visit summaries for all venues visited by the squad.
// Returns domain.ErrSquadNotFound if userID is not in the squad.
func (u *HistoryUseCase) List(ctx context.Context, squadID, userID string) ([]domain.VisitSummary, error) {
	if _, err := u.squads.Members(ctx, squadID, userID); err != nil {
		return nil, err
	}

	return u.visits.GetVisitHistory(ctx, squadID)
}
