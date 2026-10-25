package application

import (
	"context"

	"github.com/SamPariat/roundup/services/venue-service/internal/domain"
	"github.com/SamPariat/roundup/shared/pagination"
)

// FavoriteUseCase handles the favorites feature for a squad. Every
// squad-scoped method first checks that the caller is a member of the squad,
// then delegates to the VenueRepository port.
type FavoriteUseCase struct {
	favorites domain.VenueRepository
	squads    domain.SquadDirectory
	venues    VenueLookup
}

// NewFavoriteUseCase constructs a FavoriteUseCase with the given repository,
// squad directory and venue lookup.
func NewFavoriteUseCase(favorites domain.VenueRepository, squads domain.SquadDirectory, venues VenueLookup) *FavoriteUseCase {
	return &FavoriteUseCase{favorites: favorites, squads: squads, venues: venues}
}

// Add saves a venue as a favorite for the user within their squad, with a
// snapshot of its display fields. Returns domain.ErrSquadNotFound if the user
// is not in the squad, domain.ErrVenueNotFound for an unknown place and
// domain.ErrAlreadySaved if the venue is already in their favorites.
func (u *FavoriteUseCase) Add(ctx context.Context, cmd domain.AddFavoriteCommand) error {
	snapshot, err := memberSnapshot(ctx, u.squads, u.venues, cmd.SquadID, cmd.UserID, cmd.PlaceID)
	if err != nil {
		return err
	}

	return u.favorites.AddFavorite(ctx, cmd, snapshot)
}

// Remove removes a saved venue from the user's favorites within their squad.
// Returns domain.ErrSquadNotFound if the user is not in the squad and
// domain.ErrNotSaved if the venue is not in their favorites.
func (u *FavoriteUseCase) Remove(ctx context.Context, cmd domain.RemoveFavoriteCommand) error {
	if _, err := u.squads.Members(ctx, cmd.SquadID, cmd.UserID); err != nil {
		return err
	}

	return u.favorites.RemoveFavorite(ctx, cmd)
}

// List returns one page of the squad's favorites, one entry per venue with
// the members who saved it, and the next page's cursor ("" on the last page).
// Returns domain.ErrSquadNotFound if userID is not in the squad.
func (u *FavoriteUseCase) List(ctx context.Context, squadID, userID string, page pagination.Page) ([]domain.SquadFavorite, string, error) {
	members, err := u.squads.Members(ctx, squadID, userID)
	if err != nil {
		return nil, "", err
	}

	memberIDs := make([]string, 0, len(members))
	for _, m := range members {
		memberIDs = append(memberIDs, m.UserID)
	}

	groups, next, err := u.favorites.ListSquadFavorites(ctx, squadID, memberIDs, page)
	if err != nil {
		return nil, "", err
	}

	return withMembers(groups, members, userID), next, nil
}

// SavedIn reports, for each of the caller's squads, whether they saved
// placeID there. A caller with no squads gets an empty slice.
func (u *FavoriteUseCase) SavedIn(ctx context.Context, userID, placeID string) ([]domain.SquadSaved, error) {
	squads, err := u.squads.Squads(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]domain.SquadSaved, 0, len(squads))
	if len(squads) == 0 {
		return result, nil
	}

	ids := make([]string, 0, len(squads))
	for _, s := range squads {
		ids = append(ids, s.ID)
	}

	savedIDs, err := u.favorites.SavedInSquads(ctx, userID, placeID, ids)
	if err != nil {
		return nil, err
	}

	saved := make(map[string]bool, len(savedIDs))
	for _, id := range savedIDs {
		saved[id] = true
	}

	for _, s := range squads {
		result = append(result, domain.SquadSaved{SquadID: s.ID, SquadName: s.Name, Saved: saved[s.ID]})
	}

	return result, nil
}

// withMembers swaps each group's user IDs for the members' names and avatars
// and marks the groups callerID saved. The database already grouped, ordered
// and filtered to current members; an ID missing from members (someone who
// left between the two calls) is skipped.
func withMembers(groups []domain.FavoriteGroup, members []domain.SquadMember, callerID string) []domain.SquadFavorite {
	byID := make(map[string]domain.SquadMember, len(members))
	for _, m := range members {
		byID[m.UserID] = m
	}

	result := make([]domain.SquadFavorite, 0, len(groups))
	for _, g := range groups {
		favorite := domain.SquadFavorite{
			PlaceID:       g.PlaceID,
			VenueSnapshot: g.VenueSnapshot,
			SavedBy:       make([]domain.SquadMember, 0, len(g.SavedBy)),
			LastSavedAt:   g.LastSavedAt,
		}

		for _, id := range g.SavedBy {
			if member, ok := byID[id]; ok {
				favorite.SavedBy = append(favorite.SavedBy, member)
			}

			if id == callerID {
				favorite.SavedByMe = true
			}
		}

		result = append(result, favorite)
	}

	return result
}
