package application

import (
	"context"

	"github.com/SamPariat/roundup/services/venue-service/internal/domain"
)

// FavoriteUseCase handles the favorites feature for a squad. Every method
// first checks that the caller is a member of the squad, then delegates to
// the VenueRepository port.
type FavoriteUseCase struct {
	favorites domain.VenueRepository
	squads    domain.SquadDirectory
}

// NewFavoriteUseCase constructs a FavoriteUseCase with the given repository
// and squad directory.
func NewFavoriteUseCase(favorites domain.VenueRepository, squads domain.SquadDirectory) *FavoriteUseCase {
	return &FavoriteUseCase{favorites: favorites, squads: squads}
}

// AddFavorite saves a venue as a favorite for the user within their squad.
// Returns domain.ErrSquadNotFound if the user is not in the squad and
// domain.ErrAlreadySaved if the venue is already in their favorites.
func (u *FavoriteUseCase) AddFavorite(ctx context.Context, cmd domain.AddFavoriteCommand) error {
	if err := u.squads.EnsureMember(ctx, cmd.SquadID, cmd.UserID); err != nil {
		return err
	}

	return u.favorites.AddFavorite(ctx, cmd)
}

// RemoveFavorite removes a saved venue from the user's favorites within their squad.
// Returns domain.ErrSquadNotFound if the user is not in the squad and
// domain.ErrNotSaved if the venue is not in their favorites.
func (u *FavoriteUseCase) RemoveFavorite(ctx context.Context, cmd domain.RemoveFavoriteCommand) error {
	if err := u.squads.EnsureMember(ctx, cmd.SquadID, cmd.UserID); err != nil {
		return err
	}

	return u.favorites.RemoveFavorite(ctx, cmd)
}

// ListFavorites returns all venues saved as favorites by any member of the squad.
// Returns domain.ErrSquadNotFound if userID is not in the squad.
func (u *FavoriteUseCase) ListFavorites(ctx context.Context, squadID, userID string) ([]domain.SavedVenue, error) {
	if err := u.squads.EnsureMember(ctx, squadID, userID); err != nil {
		return nil, err
	}

	return u.favorites.ListFavorites(ctx, squadID)
}

// IsFavorite reports whether the given venue is in the user's favorites for the squad.
// Returns domain.ErrSquadNotFound if userID is not in the squad.
func (u *FavoriteUseCase) IsFavorite(ctx context.Context, squadID, userID, placeID string) (bool, error) {
	if err := u.squads.EnsureMember(ctx, squadID, userID); err != nil {
		return false, err
	}

	return u.favorites.IsFavorite(ctx, squadID, userID, placeID)
}
