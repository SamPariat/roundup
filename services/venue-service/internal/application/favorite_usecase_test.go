package application

import (
	"context"
	"testing"

	"github.com/SamPariat/roundup/services/venue-service/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// favorite_usecase_test.go tests FavoriteUseCase using a hand-rolled fake for the
// VenueRepository port. No real database calls are made.

type fakeVenueRepo struct {
	addFavoriteFn     func(ctx context.Context, cmd domain.AddFavoriteCommand) error
	removeFavoriteFn  func(ctx context.Context, cmd domain.RemoveFavoriteCommand) error
	listFavoritesFn   func(ctx context.Context, squadID string) ([]domain.SavedVenue, error)
	isFavoriteFn      func(ctx context.Context, squadID, userID, placeID string) (bool, error)
	recordVisitFn     func(ctx context.Context, cmd domain.RecordVisitCommand) error
	getVisitHistoryFn func(ctx context.Context, squadID string) ([]domain.VisitSummary, error)
}

func (f *fakeVenueRepo) AddFavorite(ctx context.Context, cmd domain.AddFavoriteCommand) error {
	if f.addFavoriteFn != nil {
		return f.addFavoriteFn(ctx, cmd)
	}

	return nil
}

func (f *fakeVenueRepo) RemoveFavorite(ctx context.Context, cmd domain.RemoveFavoriteCommand) error {
	if f.removeFavoriteFn != nil {
		return f.removeFavoriteFn(ctx, cmd)
	}

	return nil
}

func (f *fakeVenueRepo) ListFavorites(ctx context.Context, squadID string) ([]domain.SavedVenue, error) {
	if f.listFavoritesFn != nil {
		return f.listFavoritesFn(ctx, squadID)
	}

	return nil, nil
}

func (f *fakeVenueRepo) IsFavorite(ctx context.Context, squadID, userID, placeID string) (bool, error) {
	if f.isFavoriteFn != nil {
		return f.isFavoriteFn(ctx, squadID, userID, placeID)
	}

	return false, nil
}

func (f *fakeVenueRepo) RecordVisit(ctx context.Context, cmd domain.RecordVisitCommand) error {
	if f.recordVisitFn != nil {
		return f.recordVisitFn(ctx, cmd)
	}

	return nil
}

func (f *fakeVenueRepo) GetVisitHistory(ctx context.Context, squadID string) ([]domain.VisitSummary, error) {
	if f.getVisitHistoryFn != nil {
		return f.getVisitHistoryFn(ctx, squadID)
	}

	return nil, nil
}

// fakeSquadDirectory is a domain.SquadDirectory that admits everyone unless err is set.
type fakeSquadDirectory struct {
	err error
}

func (f fakeSquadDirectory) EnsureMember(_ context.Context, _, _ string) error {
	return f.err
}

// mustNotTouchRepo returns a fakeVenueRepo that fails the test on any call.
func mustNotTouchRepo(t *testing.T) *fakeVenueRepo {
	t.Helper()

	fail := func() { t.Fatal("the repository must not be called for a non-member") }

	return &fakeVenueRepo{
		addFavoriteFn: func(_ context.Context, _ domain.AddFavoriteCommand) error {
			fail()

			return nil
		},
		removeFavoriteFn: func(_ context.Context, _ domain.RemoveFavoriteCommand) error {
			fail()

			return nil
		},
		listFavoritesFn: func(_ context.Context, _ string) ([]domain.SavedVenue, error) {
			fail()

			return nil, nil
		},
		isFavoriteFn: func(_ context.Context, _, _, _ string) (bool, error) {
			fail()

			return false, nil
		},
		recordVisitFn: func(_ context.Context, _ domain.RecordVisitCommand) error {
			fail()

			return nil
		},
		getVisitHistoryFn: func(_ context.Context, _ string) ([]domain.VisitSummary, error) {
			fail()

			return nil, nil
		},
	}
}

func TestNonMembersNeverReachTheRepository(t *testing.T) {
	notMember := fakeSquadDirectory{err: domain.ErrSquadNotFound}
	favorites := NewFavoriteUseCase(mustNotTouchRepo(t), notMember)
	history := NewHistoryUseCase(mustNotTouchRepo(t), notMember)
	ctx := t.Context()

	_, listErr := favorites.ListFavorites(ctx, "squad-1", "stranger")
	_, checkErr := favorites.IsFavorite(ctx, "squad-1", "stranger", "abc")
	_, historyErr := history.GetVisitHistory(ctx, "squad-1", "stranger")

	for name, err := range map[string]error{
		"AddFavorite":     favorites.AddFavorite(ctx, domain.AddFavoriteCommand{SquadID: "squad-1", UserID: "stranger"}),
		"RemoveFavorite":  favorites.RemoveFavorite(ctx, domain.RemoveFavoriteCommand{SquadID: "squad-1", UserID: "stranger"}),
		"ListFavorites":   listErr,
		"IsFavorite":      checkErr,
		"RecordVisit":     history.RecordVisit(ctx, domain.RecordVisitCommand{SquadID: "squad-1", UserID: "stranger"}),
		"GetVisitHistory": historyErr,
	} {
		assert.ErrorIs(t, err, domain.ErrSquadNotFound, name)
	}
}

func TestAddFavoriteSuccess(t *testing.T) {
	uc := NewFavoriteUseCase(&fakeVenueRepo{}, fakeSquadDirectory{})

	err := uc.AddFavorite(t.Context(), domain.AddFavoriteCommand{
		SquadID: "squad-1", UserID: "user-1", PlaceID: "abc", Name: "The Rooftop",
	})

	require.NoError(t, err)
}

func TestAddFavoriteAlreadySaved(t *testing.T) {
	uc := NewFavoriteUseCase(&fakeVenueRepo{
		addFavoriteFn: func(_ context.Context, _ domain.AddFavoriteCommand) error {
			return domain.ErrAlreadySaved
		},
	}, fakeSquadDirectory{})

	err := uc.AddFavorite(t.Context(), domain.AddFavoriteCommand{})

	assert.ErrorIs(t, err, domain.ErrAlreadySaved)
}

func TestRemoveFavoriteSuccess(t *testing.T) {
	uc := NewFavoriteUseCase(&fakeVenueRepo{}, fakeSquadDirectory{})

	err := uc.RemoveFavorite(t.Context(), domain.RemoveFavoriteCommand{
		SquadID: "squad-1", UserID: "user-1", PlaceID: "abc",
	})
	require.NoError(t, err)
}

func TestRemoveFavoriteNotSaved(t *testing.T) {
	uc := NewFavoriteUseCase(&fakeVenueRepo{
		removeFavoriteFn: func(_ context.Context, _ domain.RemoveFavoriteCommand) error {
			return domain.ErrNotSaved
		},
	}, fakeSquadDirectory{})

	err := uc.RemoveFavorite(t.Context(), domain.RemoveFavoriteCommand{})
	assert.ErrorIs(t, err, domain.ErrNotSaved)
}

func TestListFavoritesReturnsSavedVenues(t *testing.T) {
	saved := []domain.SavedVenue{{PlaceID: "abc", Name: "The Rooftop"}}
	uc := NewFavoriteUseCase(&fakeVenueRepo{
		listFavoritesFn: func(_ context.Context, _ string) ([]domain.SavedVenue, error) {
			return saved, nil
		},
	}, fakeSquadDirectory{})

	got, err := uc.ListFavorites(t.Context(), "squad-1", "user-1")

	require.NoError(t, err)
	assert.Equal(t, saved, got)
}

func TestIsFavoriteTrue(t *testing.T) {
	uc := NewFavoriteUseCase(&fakeVenueRepo{
		isFavoriteFn: func(_ context.Context, _, _, _ string) (bool, error) {
			return true, nil
		},
	}, fakeSquadDirectory{})

	got, err := uc.IsFavorite(t.Context(), "squad-1", "user-1", "abc")
	require.NoError(t, err)
	assert.True(t, got)
}

func TestIsFavoriteFalse(t *testing.T) {
	uc := NewFavoriteUseCase(&fakeVenueRepo{}, fakeSquadDirectory{})

	got, err := uc.IsFavorite(t.Context(), "squad-1", "user-1", "abc")
	require.NoError(t, err)
	assert.False(t, got)
}
