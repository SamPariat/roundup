package application

import (
	"context"

	"github.com/SamPariatIL/roundup/services/venue-service/internal/domain"
)

// favorite_usecase.go implements FavoriteUseCase, which handles saving, removing,
// and listing venue favorites for a squad. All operations are delegated directly to
// VenueRepository — no caching or provider calls are made here.

// FavoriteUseCase handles the favorites feature for a squad.
// It delegates all persistence operations to the VenueRepository port.
type FavoriteUseCase struct {
	repo domain.VenueRepository
}

// NewFavoriteUseCase constructs a FavoriteUseCase with the given repository.
func NewFavoriteUseCase(repo domain.VenueRepository) *FavoriteUseCase {
	return &FavoriteUseCase{repo: repo}
}

// AddFavorite saves a venue as a favorite for the user within their squad.
// Returns domain.ErrAlreadySaved if the venue is already in the user's favorites.
func (u *FavoriteUseCase) AddFavorite(ctx context.Context, cmd domain.AddFavoriteCommand) error {
	return u.repo.AddFavorite(ctx, cmd)
}

// RemoveFavorite removes a saved venue from the user's favorites within their squad.
// Returns domain.ErrNotSaved if the venue is not in the user's favorites.
func (u *FavoriteUseCase) RemoveFavorite(ctx context.Context, cmd domain.RemoveFavoriteCommand) error {
	return u.repo.RemoveFavorite(ctx, cmd)
}

// ListFavorites returns all venues saved as favorites by any member of the squad.
func (u *FavoriteUseCase) ListFavorites(ctx context.Context, squadID string) ([]domain.SavedVenue, error) {
	return u.repo.ListFavorites(ctx, squadID)
}

// IsFavorite reports whether the given venue is in the user's favorites for the squad.
func (u *FavoriteUseCase) IsFavorite(ctx context.Context, squadID, userID, placeID string) (bool, error) {
	return u.repo.IsFavorite(ctx, squadID, userID, placeID)
}
