package application

import (
	"context"
	"testing"

	"github.com/SamPariatIL/roundup/services/venue-service/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// favorite_usecase_test.go tests FavoriteUseCase using a hand-rolled fake for the
// VenueRepository port. No real database calls are made.

type fakeRepo struct {
	addFavoriteFn     func(ctx context.Context, cmd domain.AddFavoriteCommand) error
	removeFavoriteFn  func(ctx context.Context, cmd domain.RemoveFavoriteCommand) error
	listFavoritesFn   func(ctx context.Context, squadID string) ([]domain.SavedVenue, error)
	isFavoriteFn      func(ctx context.Context, squadID, userID, placeID string) (bool, error)
	recordVisitFn     func(ctx context.Context, cmd domain.RecordVisitCommand) error
	getVisitHistoryFn func(ctx context.Context, squadID string) ([]domain.VisitSummary, error)
}

func (f *fakeRepo) AddFavorite(ctx context.Context, cmd domain.AddFavoriteCommand) error {
	if f.addFavoriteFn != nil {
		return f.addFavoriteFn(ctx, cmd)
	}

	return nil
}

func (f *fakeRepo) RemoveFavorite(ctx context.Context, cmd domain.RemoveFavoriteCommand) error {
	if f.removeFavoriteFn != nil {
		return f.removeFavoriteFn(ctx, cmd)
	}

	return nil
}

func (f *fakeRepo) ListFavorites(ctx context.Context, squadID string) ([]domain.SavedVenue, error) {
	if f.listFavoritesFn != nil {
		return f.listFavoritesFn(ctx, squadID)
	}

	return nil, nil
}

func (f *fakeRepo) IsFavorite(ctx context.Context, squadID, userID, placeID string) (bool, error) {
	if f.isFavoriteFn != nil {
		return f.isFavoriteFn(ctx, squadID, userID, placeID)
	}

	return false, nil
}

func (f *fakeRepo) RecordVisit(ctx context.Context, cmd domain.RecordVisitCommand) error {
	if f.recordVisitFn != nil {
		return f.recordVisitFn(ctx, cmd)
	}

	return nil
}

func (f *fakeRepo) GetVisitHistory(ctx context.Context, squadID string) ([]domain.VisitSummary, error) {
	if f.getVisitHistoryFn != nil {
		return f.getVisitHistoryFn(ctx, squadID)
	}

	return nil, nil
}

func TestAddFavoriteSuccess(t *testing.T) {
	uc := NewFavoriteUseCase(&fakeRepo{})

	err := uc.AddFavorite(context.Background(), domain.AddFavoriteCommand{
		SquadID: "squad-1", UserID: "user-1", PlaceID: "abc", Name: "The Rooftop",
	})

	require.NoError(t, err)
}

func TestAddFavoriteAlreadySaved(t *testing.T) {
	uc := NewFavoriteUseCase(&fakeRepo{
		addFavoriteFn: func(_ context.Context, _ domain.AddFavoriteCommand) error {
			return domain.ErrAlreadySaved
		},
	})

	err := uc.AddFavorite(context.Background(), domain.AddFavoriteCommand{})

	assert.ErrorIs(t, err, domain.ErrAlreadySaved)
}

func TestRemoveFavoriteSuccess(t *testing.T) {
	uc := NewFavoriteUseCase(&fakeRepo{})

	err := uc.RemoveFavorite(context.Background(), domain.RemoveFavoriteCommand{
		SquadID: "squad-1", UserID: "user-1", PlaceID: "abc",
	})
	require.NoError(t, err)
}

func TestRemoveFavoriteNotSaved(t *testing.T) {
	uc := NewFavoriteUseCase(&fakeRepo{
		removeFavoriteFn: func(_ context.Context, _ domain.RemoveFavoriteCommand) error {
			return domain.ErrNotSaved
		},
	})

	err := uc.RemoveFavorite(context.Background(), domain.RemoveFavoriteCommand{})
	assert.ErrorIs(t, err, domain.ErrNotSaved)
}

func TestListFavoritesReturnsSavedVenues(t *testing.T) {
	saved := []domain.SavedVenue{{PlaceID: "abc", Name: "The Rooftop"}}
	uc := NewFavoriteUseCase(&fakeRepo{
		listFavoritesFn: func(_ context.Context, _ string) ([]domain.SavedVenue, error) {
			return saved, nil
		},
	})

	got, err := uc.ListFavorites(context.Background(), "squad-1")

	require.NoError(t, err)
	assert.Equal(t, saved, got)
}

func TestIsFavoriteTrue(t *testing.T) {
	uc := NewFavoriteUseCase(&fakeRepo{
		isFavoriteFn: func(_ context.Context, _, _, _ string) (bool, error) {
			return true, nil
		},
	})

	got, err := uc.IsFavorite(context.Background(), "squad-1", "user-1", "abc")
	require.NoError(t, err)
	assert.True(t, got)
}

func TestIsFavoriteFalse(t *testing.T) {
	uc := NewFavoriteUseCase(&fakeRepo{})

	got, err := uc.IsFavorite(context.Background(), "squad-1", "user-1", "abc")
	require.NoError(t, err)
	assert.False(t, got)
}
