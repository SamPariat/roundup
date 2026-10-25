package application

import (
	"context"

	"github.com/SamPariat/roundup/services/venue-service/internal/domain"
)

// memberSnapshot checks that userID is in squadID, then returns placeID's
// display fields from the cached venue details. Saving a favorite and
// recording a visit both store this snapshot. Returns domain.ErrSquadNotFound
// for a non-member (before any venue lookup) and domain.ErrVenueNotFound for
// an unknown place.
func memberSnapshot(ctx context.Context, squads domain.SquadDirectory, venues VenueLookup, squadID, userID, placeID string) (domain.VenueSnapshot, error) {
	if _, err := squads.Members(ctx, squadID, userID); err != nil {
		return domain.VenueSnapshot{}, err
	}

	detail, err := venues.GetDetail(ctx, placeID)
	if err != nil {
		return domain.VenueSnapshot{}, err
	}

	return detail.Snapshot(), nil
}
