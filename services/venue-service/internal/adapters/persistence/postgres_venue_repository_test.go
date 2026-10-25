package persistence

import (
	"fmt"
	"testing"
	"time"

	"github.com/SamPariat/roundup/services/venue-service/internal/domain"
	"github.com/SamPariat/roundup/shared/pagination"
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
	dev        = "dddddddd-dddd-dddd-dddd-dddddddddddd"
)

// day returns 20:00 UTC on the given day of October 2026.
func day(d int) time.Time {
	return time.Date(2026, 10, d, 20, 0, 0, 0, time.UTC)
}

// firstPage is the default first page, for tests with fewer rows than a page.
var firstPage = pagination.Page{Limit: pagination.DefaultLimit}

// members is the current member list most tests pass to ListSquadFavorites.
var members = []string{alice, bob}

func newRepo(t *testing.T) (*PostgresVenueRepository, *pgxpool.Pool) {
	t.Helper()
	pool := postgrestest.New(t, migrationsDir)

	return NewPostgresVenueRepository(pool), pool
}

// snapshotOf returns a fully populated snapshot for placeID, so round trips
// show every column is stored and read back.
func snapshotOf(placeID string) domain.VenueSnapshot {
	return domain.VenueSnapshot{
		Name:          "Venue " + placeID,
		Address:       "12 Church St, Bengaluru",
		AverageRating: 4.5,
		RatingCount:   1240,
		PriceLevel:    2,
		Photo:         domain.PhotoRef{Reference: "ref-" + placeID, Width: 4032, Height: 3024},
	}
}

func addFavorite(t *testing.T, repo *PostgresVenueRepository, squadID, userID, placeID string) {
	t.Helper()
	require.NoError(t, repo.AddFavorite(t.Context(), domain.AddFavoriteCommand{
		SquadID: squadID, UserID: userID, PlaceID: placeID,
	}, snapshotOf(placeID)))
}

// --- favorites ---

func TestAddFavoriteStoresSnapshot(t *testing.T) {
	repo, _ := newRepo(t)
	addFavorite(t, repo, squad, alice, "place-1")

	favs, _, err := repo.ListSquadFavorites(t.Context(), squad, members, firstPage)
	require.NoError(t, err)

	require.Len(t, favs, 1)
	assert.Equal(t, snapshotOf("place-1"), favs[0].VenueSnapshot)
}

func TestSavedInSquadsIsPerUserAndLimitedToGivenSquads(t *testing.T) {
	ctx := t.Context()
	repo, _ := newRepo(t)
	addFavorite(t, repo, squad, alice, "place-1")
	addFavorite(t, repo, otherSquad, bob, "place-1")

	got, err := repo.SavedInSquads(ctx, alice, "place-1", []string{squad, otherSquad})
	require.NoError(t, err)
	assert.Equal(t, []string{squad}, got, "favorites are per user")

	got, err = repo.SavedInSquads(ctx, alice, "place-1", []string{otherSquad})
	require.NoError(t, err)
	assert.Empty(t, got, "squads outside the list are not reported")
}

func TestAddFavoriteTwiceReturnsErrAlreadySaved(t *testing.T) {
	repo, _ := newRepo(t)
	addFavorite(t, repo, squad, alice, "place-1")

	err := repo.AddFavorite(t.Context(), domain.AddFavoriteCommand{SquadID: squad, UserID: alice, PlaceID: "place-1"}, snapshotOf("again"))

	assert.ErrorIs(t, err, domain.ErrAlreadySaved)
}

func TestSamePlaceCanBeSavedByEachMemberAndSquad(t *testing.T) {
	repo, _ := newRepo(t)

	addFavorite(t, repo, squad, alice, "place-1")
	addFavorite(t, repo, squad, bob, "place-1")
	addFavorite(t, repo, otherSquad, alice, "place-1")
}

func TestRemoveFavoriteOnlyByOwner(t *testing.T) {
	ctx := t.Context()
	repo, _ := newRepo(t)
	addFavorite(t, repo, squad, alice, "place-1")

	err := repo.RemoveFavorite(ctx, domain.RemoveFavoriteCommand{SquadID: squad, UserID: bob, PlaceID: "place-1"})
	require.ErrorIs(t, err, domain.ErrNotSaved, "bob cannot remove alice's favorite")

	require.NoError(t, repo.RemoveFavorite(ctx, domain.RemoveFavoriteCommand{SquadID: squad, UserID: alice, PlaceID: "place-1"}))

	favs, _, err := repo.ListSquadFavorites(ctx, squad, members, firstPage)
	require.NoError(t, err)
	assert.Empty(t, favs)
}

// saveAt saves placeID for userID in squad with a fixed saved_at, so the
// grouping tests don't depend on NOW().
func saveAt(t *testing.T, pool *pgxpool.Pool, userID, placeID, name string, at time.Time) {
	t.Helper()

	_, err := pool.Exec(t.Context(),
		"INSERT INTO saved_venues (squad_id, user_id, place_id, name, saved_at) VALUES ($1, $2, $3, $4, $5)",
		squad, userID, placeID, name, at)
	require.NoError(t, err)
}

func TestListSquadFavoritesGroupsPerVenueNewestFirst(t *testing.T) {
	repo, pool := newRepo(t)

	saveAt(t, pool, alice, "toit", "Toit", day(15))
	saveAt(t, pool, bob, "toit", "Toit Brewpub", day(20))
	saveAt(t, pool, alice, "byg", "Byg Brewski", day(18))
	addFavorite(t, repo, otherSquad, alice, "elsewhere")

	favs, _, err := repo.ListSquadFavorites(t.Context(), squad, members, firstPage)
	require.NoError(t, err)

	require.Len(t, favs, 2, "one row per venue, only this squad's")
	assert.Equal(t, "toit", favs[0].PlaceID, "most recently saved first")
	assert.Equal(t, "Toit Brewpub", favs[0].Name, "display fields come from the latest save")
	assert.Equal(t, []string{bob, alice}, favs[0].SavedBy, "savers most recent first")
	assert.True(t, day(20).Equal(favs[0].LastSavedAt))
	assert.Equal(t, "byg", favs[1].PlaceID)
	assert.Equal(t, []string{alice}, favs[1].SavedBy)
}

func TestListSquadFavoritesSkipsFormerMembers(t *testing.T) {
	repo, pool := newRepo(t)

	saveAt(t, pool, dev, "toit", "Toit (old name)", day(25))
	saveAt(t, pool, alice, "toit", "Toit", day(15))
	saveAt(t, pool, dev, "arbor", "Arbor", day(5))

	favs, _, err := repo.ListSquadFavorites(t.Context(), squad, members, firstPage)
	require.NoError(t, err)

	require.Len(t, favs, 1, "a venue saved only by someone who left is not listed")
	assert.Equal(t, []string{alice}, favs[0].SavedBy)
	assert.Equal(t, "Toit", favs[0].Name, "a former member's save doesn't decide the display fields")
	assert.True(t, day(15).Equal(favs[0].LastSavedAt))
}

func TestListSquadFavoritesPagesWithoutGapsOrRepeats(t *testing.T) {
	repo, pool := newRepo(t)

	for i := range 25 {
		saveAt(t, pool, alice, fmt.Sprintf("place-%02d", i), "Venue", day(1).Add(time.Duration(i)*time.Hour))
	}

	saveAt(t, pool, bob, "place-tie-a", "Same time", day(10))
	saveAt(t, pool, alice, "place-tie-b", "Same time", day(10))

	var (
		seen   []string
		sizes  []int
		cursor string
	)

	for {
		favs, next, err := repo.ListSquadFavorites(t.Context(), squad, members, pagination.Page{Limit: 10, Cursor: cursor})
		require.NoError(t, err)

		sizes = append(sizes, len(favs))
		for _, f := range favs {
			seen = append(seen, f.PlaceID)
		}

		if next == "" {
			break
		}

		cursor = next
	}

	assert.Equal(t, []int{10, 10, 7}, sizes)
	assert.Len(t, seen, 27)
	assert.ElementsMatch(t, seen, unique(seen), "no venue appears on two pages, including ties on last_saved_at")
	assert.Equal(t, []string{"place-tie-b", "place-tie-a"}, seen[:2], "ties broken by place_id")
}

func TestListSquadFavoritesRejectsForeignCursor(t *testing.T) {
	repo, _ := newRepo(t)

	_, _, err := repo.ListSquadFavorites(t.Context(), squad, members, pagination.Page{Limit: 10, Cursor: "not-ours"})

	assert.ErrorIs(t, err, pagination.ErrInvalid)
}

func TestListSquadFavoritesEmptySquad(t *testing.T) {
	repo, _ := newRepo(t)

	favs, _, err := repo.ListSquadFavorites(t.Context(), squad, members, firstPage)

	require.NoError(t, err)
	assert.Empty(t, favs)
}

// --- visit history ---

func recordVisit(t *testing.T, repo *PostgresVenueRepository, eventID, placeID string, at time.Time, spend int64) {
	t.Helper()
	require.NoError(t, repo.RecordVisit(t.Context(), domain.RecordVisitCommand{
		SquadID: squad, EventID: eventID, PlaceID: placeID, VisitedAt: at, AvgSpendInPaise: spend,
	}, snapshotOf(placeID)))
}

func TestVisitHistoryAggregatesPerVenueMostRecentFirst(t *testing.T) {
	repo, _ := newRepo(t)

	recordVisit(t, repo, "event-1", "toit", day(1), 150000)
	recordVisit(t, repo, "event-2", "toit", day(5), 0)
	recordVisit(t, repo, "event-3", "byg", day(9), 90000)

	history, _, err := repo.GetVisitHistory(t.Context(), squad, firstPage)
	require.NoError(t, err)

	require.Len(t, history, 2)
	assert.Equal(t, "byg", history[0].PlaceID, "most recently visited first")
	assert.Equal(t, int64(1), history[0].VisitCount)
	assert.Equal(t, "toit", history[1].PlaceID)
	assert.Equal(t, int64(2), history[1].VisitCount)
	assert.True(t, day(5).Equal(history[1].LastVisitedAt), "last visit is the latest of the two")
	assert.Equal(t, int64(150000), history[1].AvgSpendInPaise, "visits with an unknown spend are left out of the average")
	assert.Equal(t, "12 Church St, Bengaluru", history[1].Address)
	assert.Equal(t, snapshotOf("toit").Photo, history[1].Photo)
}

func TestVisitHistoryAvgSpendIsZeroWhenNoneKnown(t *testing.T) {
	repo, _ := newRepo(t)
	recordVisit(t, repo, "event-1", "toit", time.Now(), 0)

	history, _, err := repo.GetVisitHistory(t.Context(), squad, firstPage)
	require.NoError(t, err)

	require.Len(t, history, 1)
	assert.Zero(t, history[0].AvgSpendInPaise)
}

func TestVisitHistoryUsesLatestSnapshot(t *testing.T) {
	repo, _ := newRepo(t)

	renamed := snapshotOf("toit")
	renamed.Name = "Toit Brewpub"

	recordVisit(t, repo, "event-1", "toit", day(1), 0)
	require.NoError(t, repo.RecordVisit(t.Context(), domain.RecordVisitCommand{
		SquadID: squad, EventID: "event-2", PlaceID: "toit", VisitedAt: day(5),
	}, renamed))

	history, _, err := repo.GetVisitHistory(t.Context(), squad, firstPage)
	require.NoError(t, err)

	require.Len(t, history, 1, "a renamed venue is still one row")
	assert.Equal(t, "Toit Brewpub", history[0].Name)
}

func TestVisitHistoryPages(t *testing.T) {
	repo, _ := newRepo(t)

	for i := range 12 {
		recordVisit(t, repo, fmt.Sprintf("event-%02d", i), fmt.Sprintf("place-%02d", i), day(1).Add(time.Duration(i)*time.Hour), 0)
	}

	first, next, err := repo.GetVisitHistory(t.Context(), squad, pagination.Page{Limit: 5})
	require.NoError(t, err)
	require.Len(t, first, 5)
	require.NotEmpty(t, next)
	assert.Equal(t, "place-11", first[0].PlaceID, "most recent first")

	second, next, err := repo.GetVisitHistory(t.Context(), squad, pagination.Page{Limit: 5, Cursor: next})
	require.NoError(t, err)
	assert.Equal(t, "place-06", second[0].PlaceID, "continues right after the first page")

	last, next, err := repo.GetVisitHistory(t.Context(), squad, pagination.Page{Limit: 5, Cursor: next})
	require.NoError(t, err)
	assert.Len(t, last, 2)
	assert.Empty(t, next, "no cursor after the last page")
}

func TestRecordVisitTwiceForSameEventReturnsErrVisitAlreadyRecorded(t *testing.T) {
	repo, _ := newRepo(t)
	at := time.Date(2026, 10, 1, 20, 0, 0, 0, time.UTC)
	recordVisit(t, repo, "event-1", "toit", at, 0)

	err := repo.RecordVisit(t.Context(), domain.RecordVisitCommand{
		SquadID: squad, EventID: "event-1", PlaceID: "toit", VisitedAt: at,
	}, snapshotOf("toit"))

	assert.ErrorIs(t, err, domain.ErrVisitAlreadyRecorded)
}

func TestRecordVisitStoresZeroSpendAsNull(t *testing.T) {
	repo, pool := newRepo(t)
	recordVisit(t, repo, "event-1", "toit", time.Now(), 0)
	recordVisit(t, repo, "event-2", "byg", time.Now(), 90000)

	var nulls, set int

	err := pool.QueryRow(t.Context(),
		"SELECT count(*) FILTER (WHERE avg_spend_in_paise IS NULL), count(*) FILTER (WHERE avg_spend_in_paise = 90000) FROM venue_visits",
	).Scan(&nulls, &set)
	require.NoError(t, err)

	assert.Equal(t, 1, nulls, "0 means unknown and is stored as NULL")
	assert.Equal(t, 1, set)
}

// unique returns ids without duplicates, keeping order.
func unique(ids []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(ids))

	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}

	return out
}
