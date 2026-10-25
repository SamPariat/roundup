package application

import (
	"context"
	"testing"
	"time"

	"github.com/SamPariat/roundup/services/venue-service/internal/domain"
	"github.com/SamPariat/roundup/shared/pagination"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// favorite_usecase_test.go tests FavoriteUseCase using hand-rolled fakes for the
// VenueRepository, SquadDirectory and VenueLookup ports. No real database or
// network calls are made.

type fakeVenueRepo struct {
	addFavoriteFn     func(ctx context.Context, cmd domain.AddFavoriteCommand, snapshot domain.VenueSnapshot) error
	removeFavoriteFn  func(ctx context.Context, cmd domain.RemoveFavoriteCommand) error
	listFavoritesFn   func(ctx context.Context, squadID string, memberIDs []string, page pagination.Page) ([]domain.FavoriteGroup, string, error)
	savedInSquadsFn   func(ctx context.Context, userID, placeID string, squadIDs []string) ([]string, error)
	recordVisitFn     func(ctx context.Context, cmd domain.RecordVisitCommand, snapshot domain.VenueSnapshot) error
	getVisitHistoryFn func(ctx context.Context, squadID string, page pagination.Page) ([]domain.VisitSummary, string, error)
}

// firstPage is a default first page.
var firstPage = pagination.Page{Limit: pagination.DefaultLimit}

// day returns 20:00 UTC on the given day of October 2026.
func day(d int) time.Time {
	return time.Date(2026, 10, d, 20, 0, 0, 0, time.UTC)
}

// mustNotTouchRepo returns a fakeVenueRepo that fails the test on any call.
func mustNotTouchRepo(t *testing.T) *fakeVenueRepo {
	t.Helper()

	fail := func() { t.Fatal("the repository must not be called") }

	return &fakeVenueRepo{
		addFavoriteFn: func(_ context.Context, _ domain.AddFavoriteCommand, _ domain.VenueSnapshot) error {
			fail()

			return nil
		},
		removeFavoriteFn: func(_ context.Context, _ domain.RemoveFavoriteCommand) error {
			fail()

			return nil
		},
		listFavoritesFn: func(_ context.Context, _ string, _ []string, _ pagination.Page) ([]domain.FavoriteGroup, string, error) {
			fail()

			return nil, "", nil
		},
		savedInSquadsFn: func(_ context.Context, _, _ string, _ []string) ([]string, error) {
			fail()

			return nil, nil
		},
		recordVisitFn: func(_ context.Context, _ domain.RecordVisitCommand, _ domain.VenueSnapshot) error {
			fail()

			return nil
		},
		getVisitHistoryFn: func(_ context.Context, _ string, _ pagination.Page) ([]domain.VisitSummary, string, error) {
			fail()

			return nil, "", nil
		},
	}
}

func (f *fakeVenueRepo) AddFavorite(ctx context.Context, cmd domain.AddFavoriteCommand, snapshot domain.VenueSnapshot) error {
	if f.addFavoriteFn != nil {
		return f.addFavoriteFn(ctx, cmd, snapshot)
	}

	return nil
}

func (f *fakeVenueRepo) RemoveFavorite(ctx context.Context, cmd domain.RemoveFavoriteCommand) error {
	if f.removeFavoriteFn != nil {
		return f.removeFavoriteFn(ctx, cmd)
	}

	return nil
}

func (f *fakeVenueRepo) ListSquadFavorites(ctx context.Context, squadID string, memberIDs []string, page pagination.Page) ([]domain.FavoriteGroup, string, error) {
	if f.listFavoritesFn != nil {
		return f.listFavoritesFn(ctx, squadID, memberIDs, page)
	}

	return nil, "", nil
}

func (f *fakeVenueRepo) SavedInSquads(ctx context.Context, userID, placeID string, squadIDs []string) ([]string, error) {
	if f.savedInSquadsFn != nil {
		return f.savedInSquadsFn(ctx, userID, placeID, squadIDs)
	}

	return nil, nil
}

func (f *fakeVenueRepo) RecordVisit(ctx context.Context, cmd domain.RecordVisitCommand, snapshot domain.VenueSnapshot) error {
	if f.recordVisitFn != nil {
		return f.recordVisitFn(ctx, cmd, snapshot)
	}

	return nil
}

func (f *fakeVenueRepo) GetVisitHistory(ctx context.Context, squadID string, page pagination.Page) ([]domain.VisitSummary, string, error) {
	if f.getVisitHistoryFn != nil {
		return f.getVisitHistoryFn(ctx, squadID, page)
	}

	return nil, "", nil
}

// fakeSquadDirectory is a domain.SquadDirectory that admits everyone unless
// err is set, answering with members and squads.
type fakeSquadDirectory struct {
	members []domain.SquadMember
	squads  []domain.Squad
	err     error
}

func (f fakeSquadDirectory) Members(_ context.Context, _, _ string) ([]domain.SquadMember, error) {
	return f.members, f.err
}

func (f fakeSquadDirectory) Squads(_ context.Context, _ string) ([]domain.Squad, error) {
	return f.squads, f.err
}

// fakeVenueLookup is a VenueLookup that returns detail, or err when set.
type fakeVenueLookup struct {
	detail domain.VenueDetail
	err    error
}

func (f fakeVenueLookup) GetDetail(_ context.Context, _ string) (*domain.VenueDetail, error) {
	if f.err != nil {
		return nil, f.err
	}

	return &f.detail, nil
}

func TestNonMembersNeverReachTheRepository(t *testing.T) {
	notMember := fakeSquadDirectory{err: domain.ErrSquadNotFound}
	favorites := NewFavoriteUseCase(mustNotTouchRepo(t), notMember, fakeVenueLookup{})
	history := NewHistoryUseCase(mustNotTouchRepo(t), notMember, fakeVenueLookup{})
	ctx := t.Context()

	_, _, listErr := favorites.List(ctx, "squad-1", "stranger", firstPage)
	_, _, historyErr := history.List(ctx, "squad-1", "stranger", firstPage)

	for name, err := range map[string]error{
		"favorites.Add":    favorites.Add(ctx, domain.AddFavoriteCommand{SquadID: "squad-1", UserID: "stranger"}),
		"favorites.Remove": favorites.Remove(ctx, domain.RemoveFavoriteCommand{SquadID: "squad-1", UserID: "stranger"}),
		"favorites.List":   listErr,
		"history.Record":   history.Record(ctx, domain.RecordVisitCommand{SquadID: "squad-1", UserID: "stranger"}),
		"history.List":     historyErr,
	} {
		assert.ErrorIs(t, err, domain.ErrSquadNotFound, name)
	}
}

func TestAddFavoriteStoresSnapshotOfDetail(t *testing.T) {
	var got domain.VenueSnapshot

	detail := domain.VenueDetail{
		Name: "The Rooftop", Address: "12 Church St", AverageRating: 4.5, RatingCount: 1240, PriceLevel: 2,
		PhotoRefs: []domain.PhotoRef{{Reference: "first", Width: 800, Height: 600}, {Reference: "second"}},
	}
	uc := NewFavoriteUseCase(&fakeVenueRepo{
		addFavoriteFn: func(_ context.Context, _ domain.AddFavoriteCommand, snapshot domain.VenueSnapshot) error {
			got = snapshot

			return nil
		},
	}, fakeSquadDirectory{}, fakeVenueLookup{detail: detail})

	err := uc.Add(t.Context(), domain.AddFavoriteCommand{SquadID: "squad-1", UserID: "user-1", PlaceID: "abc"})

	require.NoError(t, err)
	assert.Equal(t, domain.VenueSnapshot{
		Name: "The Rooftop", Address: "12 Church St", AverageRating: 4.5, RatingCount: 1240, PriceLevel: 2,
		Photo: domain.PhotoRef{Reference: "first", Width: 800, Height: 600},
	}, got, "only the first photo is kept")
}

func TestAddFavoriteUnknownVenueIsNotSaved(t *testing.T) {
	uc := NewFavoriteUseCase(&fakeVenueRepo{
		addFavoriteFn: func(_ context.Context, _ domain.AddFavoriteCommand, _ domain.VenueSnapshot) error {
			t.Fatal("an unknown venue must not be saved")

			return nil
		},
	}, fakeSquadDirectory{}, fakeVenueLookup{err: domain.ErrVenueNotFound})

	err := uc.Add(t.Context(), domain.AddFavoriteCommand{PlaceID: "nope"})

	assert.ErrorIs(t, err, domain.ErrVenueNotFound)
}

func TestAddFavoriteAlreadySaved(t *testing.T) {
	uc := NewFavoriteUseCase(&fakeVenueRepo{
		addFavoriteFn: func(_ context.Context, _ domain.AddFavoriteCommand, _ domain.VenueSnapshot) error {
			return domain.ErrAlreadySaved
		},
	}, fakeSquadDirectory{}, fakeVenueLookup{})

	err := uc.Add(t.Context(), domain.AddFavoriteCommand{})

	assert.ErrorIs(t, err, domain.ErrAlreadySaved)
}

func TestRemoveFavoriteSuccess(t *testing.T) {
	uc := NewFavoriteUseCase(&fakeVenueRepo{}, fakeSquadDirectory{}, fakeVenueLookup{})

	err := uc.Remove(t.Context(), domain.RemoveFavoriteCommand{
		SquadID: "squad-1", UserID: "user-1", PlaceID: "abc",
	})
	require.NoError(t, err)
}

func TestRemoveFavoriteNotSaved(t *testing.T) {
	uc := NewFavoriteUseCase(&fakeVenueRepo{
		removeFavoriteFn: func(_ context.Context, _ domain.RemoveFavoriteCommand) error {
			return domain.ErrNotSaved
		},
	}, fakeSquadDirectory{}, fakeVenueLookup{})

	err := uc.Remove(t.Context(), domain.RemoveFavoriteCommand{})
	assert.ErrorIs(t, err, domain.ErrNotSaved)
}

func TestListFavoritesAttachesMembers(t *testing.T) {
	alice := domain.SquadMember{UserID: "alice", Name: "Alice"}
	bob := domain.SquadMember{UserID: "bob", Name: "Bob"}

	var (
		gotMemberIDs []string
		gotPage      pagination.Page
	)

	// Grouped by the database; "dev" left between the two calls.
	groups := []domain.FavoriteGroup{
		{PlaceID: "toit", Name: "Toit Brewpub", SavedBy: []string{"bob", "dev", "alice"}, LastSavedAt: day(20)},
		{PlaceID: "byg", Name: "Byg Brewski", SavedBy: []string{"alice"}, LastSavedAt: day(18)},
	}
	uc := NewFavoriteUseCase(&fakeVenueRepo{
		listFavoritesFn: func(_ context.Context, _ string, memberIDs []string, page pagination.Page) ([]domain.FavoriteGroup, string, error) {
			gotMemberIDs = memberIDs
			gotPage = page

			return groups, "next-page", nil
		},
	}, fakeSquadDirectory{members: []domain.SquadMember{alice, bob}}, fakeVenueLookup{})

	page := pagination.Page{Limit: 10, Cursor: "abc"}
	got, next, err := uc.List(t.Context(), "squad-1", "bob", page)

	require.NoError(t, err)
	assert.Equal(t, []string{"alice", "bob"}, gotMemberIDs, "the database counts only current members")
	assert.Equal(t, page, gotPage, "the page is passed through")
	assert.Equal(t, "next-page", next)
	assert.Equal(t, []domain.SquadFavorite{
		{
			PlaceID:     "toit",
			Name:        "Toit Brewpub",
			SavedBy:     []domain.SquadMember{bob, alice},
			SavedByMe:   true,
			LastSavedAt: day(20),
		},
		{
			PlaceID:     "byg",
			Name:        "Byg Brewski",
			SavedBy:     []domain.SquadMember{alice},
			LastSavedAt: day(18),
		},
	}, got, "IDs become members, in order; an unknown ID is skipped")
}

func TestListFavoritesEmptyIsNotNil(t *testing.T) {
	uc := NewFavoriteUseCase(&fakeVenueRepo{}, fakeSquadDirectory{}, fakeVenueLookup{})

	got, next, err := uc.List(t.Context(), "squad-1", "alice", firstPage)

	require.NoError(t, err)
	assert.Empty(t, next)
	assert.NotNil(t, got)
	assert.Empty(t, got)
}

func TestSavedInMarksEachOfTheCallersSquads(t *testing.T) {
	var gotIDs []string

	uc := NewFavoriteUseCase(&fakeVenueRepo{
		savedInSquadsFn: func(_ context.Context, _, _ string, squadIDs []string) ([]string, error) {
			gotIDs = squadIDs

			return []string{"squad-2"}, nil
		},
	}, fakeSquadDirectory{squads: []domain.Squad{{ID: "squad-1", Name: "Friday crew"}, {ID: "squad-2", Name: "Office"}}}, fakeVenueLookup{})

	got, err := uc.SavedIn(t.Context(), "user-1", "abc")

	require.NoError(t, err)
	assert.Equal(t, []string{"squad-1", "squad-2"}, gotIDs, "one query covers every squad")
	assert.Equal(t, []domain.SquadSaved{
		{SquadID: "squad-1", SquadName: "Friday crew", Saved: false},
		{SquadID: "squad-2", SquadName: "Office", Saved: true},
	}, got)
}

func TestSavedInWithNoSquadsSkipsTheRepository(t *testing.T) {
	uc := NewFavoriteUseCase(mustNotTouchRepo(t), fakeSquadDirectory{}, fakeVenueLookup{})

	got, err := uc.SavedIn(t.Context(), "user-1", "abc")

	require.NoError(t, err)
	assert.NotNil(t, got)
	assert.Empty(t, got)
}

func TestSavedInUserServiceDown(t *testing.T) {
	uc := NewFavoriteUseCase(mustNotTouchRepo(t), fakeSquadDirectory{err: domain.ErrUserServiceUnavailable}, fakeVenueLookup{})

	_, err := uc.SavedIn(t.Context(), "user-1", "abc")

	assert.ErrorIs(t, err, domain.ErrUserServiceUnavailable)
}
