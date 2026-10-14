package persistence

import (
	"context"
	"testing"
	"time"

	"github.com/SamPariat/roundup/services/venue-service/internal/domain"
	"github.com/SamPariat/roundup/shared/postgres/postgrestest"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// postgres_venue_repository_test.go runs the repository against a real,
// freshly migrated Postgres database. Tests are skipped unless
// TEST_POSTGRES_DSN is set (see shared/postgres/postgrestest).

const migrationsDir = "../../../db/migrations"

const (
	squad      = "11111111-1111-1111-1111-111111111111"
	otherSquad = "22222222-2222-2222-2222-222222222222"
	alice      = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	bob        = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
)

func newRepo(t *testing.T) (*PostgresVenueRepository, *pgxpool.Pool) {
	t.Helper()
	pool := postgrestest.New(t, migrationsDir)
	return NewPostgresVenueRepository(pool), pool
}

func addFavorite(t *testing.T, repo *PostgresVenueRepository, squadID, userID, placeID string) {
	t.Helper()
	require.NoError(t, repo.AddFavorite(context.Background(), domain.AddFavoriteCommand{
		SquadID: squadID, UserID: userID, PlaceID: placeID, Name: "Venue " + placeID,
	}))
}

// --- favorites ---

func TestAddFavoriteThenIsFavorite(t *testing.T) {
	ctx := context.Background()
	repo, _ := newRepo(t)

	addFavorite(t, repo, squad, alice, "place-1")

	fav, err := repo.IsFavorite(ctx, squad, alice, "place-1")
	require.NoError(t, err)
	assert.True(t, fav)

	fav, err = repo.IsFavorite(ctx, squad, bob, "place-1")
	require.NoError(t, err)
	assert.False(t, fav, "favorites are per user")
}

func TestAddFavoriteTwiceReturnsErrAlreadySaved(t *testing.T) {
	repo, _ := newRepo(t)
	addFavorite(t, repo, squad, alice, "place-1")

	err := repo.AddFavorite(context.Background(), domain.AddFavoriteCommand{SquadID: squad, UserID: alice, PlaceID: "place-1", Name: "again"})

	assert.ErrorIs(t, err, domain.ErrAlreadySaved)
}

func TestSamePlaceCanBeSavedByEachMemberAndSquad(t *testing.T) {
	repo, _ := newRepo(t)

	addFavorite(t, repo, squad, alice, "place-1")
	addFavorite(t, repo, squad, bob, "place-1")
	addFavorite(t, repo, otherSquad, alice, "place-1")
}

func TestRemoveFavoriteOnlyByOwner(t *testing.T) {
	ctx := context.Background()
	repo, _ := newRepo(t)
	addFavorite(t, repo, squad, alice, "place-1")

	err := repo.RemoveFavorite(ctx, domain.RemoveFavoriteCommand{SquadID: squad, UserID: bob, PlaceID: "place-1"})
	assert.ErrorIs(t, err, domain.ErrNotSaved, "bob cannot remove alice's favorite")

	require.NoError(t, repo.RemoveFavorite(ctx, domain.RemoveFavoriteCommand{SquadID: squad, UserID: alice, PlaceID: "place-1"}))

	fav, err := repo.IsFavorite(ctx, squad, alice, "place-1")
	require.NoError(t, err)
	assert.False(t, fav)
}

func TestListFavoritesIsSquadScopedNewestFirst(t *testing.T) {
	repo, _ := newRepo(t)
	addFavorite(t, repo, squad, alice, "place-1")
	time.Sleep(10 * time.Millisecond) // saved_at defaults to NOW(); keep the order deterministic
	addFavorite(t, repo, squad, bob, "place-2")
	addFavorite(t, repo, otherSquad, alice, "place-3")

	favs, err := repo.ListFavorites(context.Background(), squad)
	require.NoError(t, err)

	require.Len(t, favs, 2)
	assert.Equal(t, "place-2", favs[0].PlaceID)
	assert.Equal(t, bob, favs[0].UserID)
	assert.Equal(t, "place-1", favs[1].PlaceID)
}

func TestListFavoritesEmptySquad(t *testing.T) {
	repo, _ := newRepo(t)

	favs, err := repo.ListFavorites(context.Background(), squad)

	require.NoError(t, err)
	assert.Empty(t, favs)
}

// --- visit history ---

func recordVisit(t *testing.T, repo *PostgresVenueRepository, eventID, placeID string, at time.Time, spend int64) {
	t.Helper()
	require.NoError(t, repo.RecordVisit(context.Background(), domain.RecordVisitCommand{
		SquadID: squad, EventID: eventID, PlaceID: placeID, Name: "Venue " + placeID, VisitedAt: at, AvgSpendInPaise: spend,
	}))
}

func TestVisitHistoryAggregatesPerVenueMostRecentFirst(t *testing.T) {
	repo, _ := newRepo(t)
	day := func(d int) time.Time { return time.Date(2026, 10, d, 20, 0, 0, 0, time.UTC) }

	recordVisit(t, repo, "event-1", "toit", day(1), 150000)
	recordVisit(t, repo, "event-2", "toit", day(5), 0)
	recordVisit(t, repo, "event-3", "byg", day(9), 90000)

	history, err := repo.GetVisitHistory(context.Background(), squad)
	require.NoError(t, err)

	require.Len(t, history, 2)
	assert.Equal(t, "byg", history[0].PlaceID, "most recently visited first")
	assert.Equal(t, int64(1), history[0].VisitCount)
	assert.Equal(t, "toit", history[1].PlaceID)
	assert.Equal(t, int64(2), history[1].VisitCount)
	assert.True(t, day(5).Equal(history[1].LastVisitedAt), "last visit is the latest of the two")
}

func TestRecordVisitTwiceForSameEventReturnsErrVisitAlreadyRecorded(t *testing.T) {
	repo, _ := newRepo(t)
	at := time.Date(2026, 10, 1, 20, 0, 0, 0, time.UTC)
	recordVisit(t, repo, "event-1", "toit", at, 0)

	err := repo.RecordVisit(context.Background(), domain.RecordVisitCommand{
		SquadID: squad, EventID: "event-1", PlaceID: "toit", Name: "Toit", VisitedAt: at,
	})

	assert.ErrorIs(t, err, domain.ErrVisitAlreadyRecorded)
}

func TestRecordVisitStoresZeroSpendAsNull(t *testing.T) {
	repo, pool := newRepo(t)
	recordVisit(t, repo, "event-1", "toit", time.Now(), 0)
	recordVisit(t, repo, "event-2", "byg", time.Now(), 90000)

	var nulls, set int
	err := pool.QueryRow(context.Background(),
		"SELECT count(*) FILTER (WHERE avg_spend_in_paise IS NULL), count(*) FILTER (WHERE avg_spend_in_paise = 90000) FROM venue_visits",
	).Scan(&nulls, &set)
	require.NoError(t, err)

	assert.Equal(t, 1, nulls, "0 means unknown and is stored as NULL")
	assert.Equal(t, 1, set)
}
