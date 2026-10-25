package persistence

import (
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/SamPariat/roundup/services/event-service/internal/domain"
	"github.com/SamPariat/roundup/shared/pagination"
	"github.com/SamPariat/roundup/shared/postgres/postgrestest"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// postgres_event_repository_test.go runs the event repository against a real,
// freshly migrated Postgres database: the status-guarded writes, the winner
// rule, the list order and paging, and the sweeps under concurrency.

const migrationsDir = "../../../db/migrations"

type eventFixture struct {
	pool   *pgxpool.Pool
	events *PostgresEventRepository
	squad  string
	alice  string
	bob    string
	carol  string
}

func newEventFixture(t *testing.T) *eventFixture {
	t.Helper()

	pool := postgrestest.New(t, migrationsDir)

	return &eventFixture{
		pool:   pool,
		events: NewPostgresEventRepository(pool),
		squad:  uuid.NewString(),
		alice:  uuid.NewString(),
		bob:    uuid.NewString(),
		carol:  uuid.NewString(),
	}
}

// event creates a DRAFT event starting at startsAt, created by alice.
func (f *eventFixture) event(t *testing.T, startsAt time.Time) *domain.Event {
	t.Helper()

	e, err := f.events.Create(t.Context(), domain.NewEvent{
		SquadID:   f.squad,
		CreatedBy: f.alice,
		Name:      "Friday drinks",
		Type:      domain.TypeDrinks,
		Schedule:  domain.Schedule{StartsAt: startsAt, VotingDeadline: startsAt.Add(-time.Hour)},
	})
	require.NoError(t, err)

	return e
}

func (f *eventFixture) nominate(t *testing.T, eventID, placeID string) {
	t.Helper()

	added, err := f.events.AddNomination(t.Context(), eventID, domain.Nomination{
		PlaceID: placeID, Name: "Place " + placeID, Lat: 19.07, Lng: 72.87, PhotoURL: "/p/" + placeID,
		NominatedBy: f.alice,
	})
	require.NoError(t, err)
	require.True(t, added)
}

func (f *eventFixture) vote(t *testing.T, eventID, userID, placeID string) {
	t.Helper()

	cast, err := f.events.CastVote(t.Context(), eventID, userID, placeID)
	require.NoError(t, err)
	require.True(t, cast)
}

// voting creates an event, suggests the places in order and opens voting.
func (f *eventFixture) voting(t *testing.T, places ...string) *domain.Event {
	t.Helper()

	e := f.event(t, time.Now().Add(48*time.Hour))
	for _, p := range places {
		f.nominate(t, e.ID, p)
	}

	started, err := f.events.StartVoting(t.Context(), e.ID)
	require.NoError(t, err)
	require.True(t, started)

	return e
}

// setStatus forces a status, for tests about what each status allows.
func (f *eventFixture) setStatus(t *testing.T, eventID string, status domain.Status) {
	t.Helper()

	_, err := f.pool.Exec(t.Context(), "UPDATE events SET status = $2 WHERE id = $1", eventID, string(status))
	require.NoError(t, err)
}

func (f *eventFixture) status(t *testing.T, eventID string) domain.Status {
	t.Helper()

	e, err := f.events.Get(t.Context(), eventID)
	require.NoError(t, err)

	return e.Status
}

func TestCreateAndGet(t *testing.T) {
	f := newEventFixture(t)
	startsAt := time.Now().Add(48 * time.Hour).Truncate(time.Microsecond)
	created := f.event(t, startsAt)

	got, err := f.events.Get(t.Context(), created.ID)
	require.NoError(t, err)

	assert.Equal(t, domain.StatusDraft, got.Status)
	assert.Equal(t, f.squad, got.SquadID)
	assert.True(t, startsAt.Equal(got.StartsAt))
	assert.Nil(t, got.Winner)
	assert.Nil(t, got.ClosedAt)

	_, err = f.events.Get(t.Context(), uuid.NewString())
	require.ErrorIs(t, err, domain.ErrEventNotFound)

	_, err = f.events.Get(t.Context(), "not-a-uuid")
	assert.ErrorIs(t, err, domain.ErrEventNotFound)
}

func TestRSVPUpsertAndClosedGuard(t *testing.T) {
	f := newEventFixture(t)
	e := f.event(t, time.Now().Add(48*time.Hour))

	for _, status := range []domain.RSVPStatus{domain.RSVPMaybe, domain.RSVPGoing} {
		set, err := f.events.SetRSVP(t.Context(), e.ID, f.bob, status)
		require.NoError(t, err)
		require.True(t, set)
	}

	rsvps, err := f.events.RSVPs(t.Context(), []string{e.ID})
	require.NoError(t, err)
	assert.Equal(t, []domain.RSVP{{EventID: e.ID, UserID: f.bob, Status: domain.RSVPGoing}}, rsvps, "a second answer replaces the first")

	f.setStatus(t, e.ID, domain.StatusClosed)

	set, err := f.events.SetRSVP(t.Context(), e.ID, f.bob, domain.RSVPNotGoing)
	require.NoError(t, err)
	assert.False(t, set, "a closed event takes no answers")
}

func TestNominationsOnlyWhileDraft(t *testing.T) {
	f := newEventFixture(t)
	e := f.event(t, time.Now().Add(48*time.Hour))
	f.nominate(t, e.ID, "p1")

	_, err := f.events.AddNomination(t.Context(), e.ID, domain.Nomination{PlaceID: "p1", NominatedBy: f.bob})
	require.ErrorIs(t, err, domain.ErrAlreadyNominated)

	f.setStatus(t, e.ID, domain.StatusVoting)

	added, err := f.events.AddNomination(t.Context(), e.ID, domain.Nomination{PlaceID: "p2", NominatedBy: f.bob})
	require.NoError(t, err)
	assert.False(t, added)

	removed, err := f.events.RemoveNomination(t.Context(), e.ID, "p1")
	require.NoError(t, err)
	assert.False(t, removed)

	f.setStatus(t, e.ID, domain.StatusDraft)

	removed, err = f.events.RemoveNomination(t.Context(), e.ID, "p1")
	require.NoError(t, err)
	assert.True(t, removed)
}

func TestVotesNeedVotingAndANomination(t *testing.T) {
	f := newEventFixture(t)
	e := f.event(t, time.Now().Add(48*time.Hour))
	f.nominate(t, e.ID, "p1")
	f.nominate(t, e.ID, "p2")

	cast, err := f.events.CastVote(t.Context(), e.ID, f.bob, "p1")
	require.NoError(t, err)
	assert.False(t, cast, "no votes on a draft")

	started, err := f.events.StartVoting(t.Context(), e.ID)
	require.NoError(t, err)
	require.True(t, started)

	cast, err = f.events.CastVote(t.Context(), e.ID, f.bob, "nowhere")
	require.NoError(t, err)
	assert.False(t, cast, "only suggested places")

	f.vote(t, e.ID, f.bob, "p1")
	f.vote(t, e.ID, f.bob, "p2")

	votes, err := f.events.Votes(t.Context(), []string{e.ID})
	require.NoError(t, err)
	assert.Equal(t, []domain.Vote{{EventID: e.ID, UserID: f.bob, PlaceID: "p2"}}, votes, "changing a vote replaces it")
}

func TestStartVotingNeedsANomination(t *testing.T) {
	f := newEventFixture(t)
	e := f.event(t, time.Now().Add(48*time.Hour))

	started, err := f.events.StartVoting(t.Context(), e.ID)
	require.NoError(t, err)
	assert.False(t, started)
	assert.Equal(t, domain.StatusDraft, f.status(t, e.ID))
}

func TestConfirmVotingPicksTheWinner(t *testing.T) {
	tests := []struct {
		name  string
		votes map[string]string // voter -> place
		want  string
	}{
		{name: "MostVotes", votes: map[string]string{"alice": "p2", "bob": "p2", "carol": "p1"}, want: "p2"},
		{name: "TieGoesToEarliest", votes: map[string]string{"alice": "p3", "bob": "p2"}, want: "p2"},
		{name: "NoVotesGoesToEarliest", votes: nil, want: "p1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newEventFixture(t)
			e := f.voting(t, "p1", "p2", "p3")
			users := map[string]string{"alice": f.alice, "bob": f.bob, "carol": f.carol}

			for voter, place := range tt.votes {
				f.vote(t, e.ID, users[voter], place)
			}

			confirmed, err := f.events.ConfirmVoting(t.Context(), e.ID)
			require.NoError(t, err)
			require.True(t, confirmed)

			got, err := f.events.Get(t.Context(), e.ID)
			require.NoError(t, err)
			assert.Equal(t, domain.StatusConfirmed, got.Status)
			require.NotNil(t, got.Winner)
			assert.Equal(t, tt.want, got.Winner.PlaceID)
			assert.Equal(t, "Place "+tt.want, got.Winner.Name, "the winner's snapshot comes from its nomination")

			again, err := f.events.ConfirmVoting(t.Context(), e.ID)
			require.NoError(t, err)
			assert.False(t, again, "confirming twice does nothing")
		})
	}
}

func TestLifecycleStepsAreGuarded(t *testing.T) {
	f := newEventFixture(t)
	e := f.voting(t, "p1")

	started, err := f.events.StartNight(t.Context(), e.ID)
	require.NoError(t, err)
	assert.False(t, started, "a night can't start before the place is confirmed")

	closed, err := f.events.Close(t.Context(), e.ID)
	require.NoError(t, err)
	assert.Nil(t, closed, "only an active night closes by hand")

	_, err = f.events.ConfirmVoting(t.Context(), e.ID)
	require.NoError(t, err)

	started, err = f.events.StartNight(t.Context(), e.ID)
	require.NoError(t, err)
	assert.True(t, started)

	closed, err = f.events.Close(t.Context(), e.ID)
	require.NoError(t, err)
	require.NotNil(t, closed)
	assert.Equal(t, f.alice, closed.CreatedBy)
	require.NotNil(t, closed.Winner)
	assert.Equal(t, "p1", closed.Winner.PlaceID)

	got, err := f.events.Get(t.Context(), e.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.StatusClosed, got.Status)
	assert.NotNil(t, got.ClosedAt)
}

func TestUpdateOnlyBeforeConfirmed(t *testing.T) {
	f := newEventFixture(t)
	e := f.event(t, time.Now().Add(48*time.Hour))
	e.Name = "Saturday dinner"
	e.Type = domain.TypeDining

	updated, err := f.events.Update(t.Context(), e)
	require.NoError(t, err)
	assert.True(t, updated)

	got, err := f.events.Get(t.Context(), e.ID)
	require.NoError(t, err)
	assert.Equal(t, "Saturday dinner", got.Name)
	assert.Equal(t, domain.TypeDining, got.Type)

	f.setStatus(t, e.ID, domain.StatusConfirmed)

	updated, err = f.events.Update(t.Context(), e)
	require.NoError(t, err)
	assert.False(t, updated)
}

func TestListOrdersOpenThenClosedAndPages(t *testing.T) {
	f := newEventFixture(t)
	base := time.Now().Add(24 * time.Hour).Truncate(time.Second)

	// 15 open events, starting a day apart; then 10 closed ones in the past.
	var open, closed []string
	for i := range 15 {
		open = append(open, f.event(t, base.Add(time.Duration(14-i)*24*time.Hour)).ID)
	}

	for i := range 10 {
		e := f.event(t, base.Add(-time.Duration(30+i)*24*time.Hour))
		f.setStatus(t, e.ID, domain.StatusClosed)
		closed = append(closed, e.ID)
	}

	// Open soonest first (created latest-first above), then closed most recent first.
	var want []string
	for _, o := range slices.Backward(open) {
		want = append(want, o)
	}

	want = append(want, closed...)

	var (
		got   []string
		sizes []int
	)

	page := pagination.Page{Limit: 10}
	for {
		events, next, err := f.events.ListForSquad(t.Context(), f.squad, page)
		require.NoError(t, err)

		sizes = append(sizes, len(events))
		for _, e := range events {
			got = append(got, e.ID)
		}

		if next == "" {
			break
		}

		page.Cursor = next
	}

	assert.Equal(t, []int{10, 10, 5}, sizes)
	assert.Equal(t, want, got)

	openCount, err := f.events.CountOpen(t.Context(), f.squad)
	require.NoError(t, err)
	assert.Equal(t, 15, openCount)

	_, _, err = f.events.ListForSquad(t.Context(), f.squad, pagination.Page{Limit: 10, Cursor: "junk"})
	assert.ErrorIs(t, err, pagination.ErrInvalid)
}

func TestCountsAcrossEvents(t *testing.T) {
	f := newEventFixture(t)
	first := f.event(t, time.Now().Add(48*time.Hour))
	second := f.event(t, time.Now().Add(72*time.Hour))
	f.nominate(t, first.ID, "p1")
	f.nominate(t, first.ID, "p2")

	counts, err := f.events.NominationCounts(t.Context(), []string{first.ID, second.ID})
	require.NoError(t, err)
	assert.Equal(t, map[string]int{first.ID: 2}, counts)

	nominations, err := f.events.Nominations(t.Context(), first.ID)
	require.NoError(t, err)
	require.Len(t, nominations, 2)
	assert.Equal(t, "p1", nominations[0].PlaceID, "earliest first")
	assert.Equal(t, f.alice, nominations[0].NominatedBy)
}

// Overlapping sweeps (a retry, or two scheduler instances) must move each due
// event exactly once.
func TestSweepsMoveEachEventOnceUnderConcurrency(t *testing.T) {
	f := newEventFixture(t)

	var due []string

	for range 5 {
		e := f.voting(t, "p1", "p2")
		_, err := f.pool.Exec(t.Context(),
			"UPDATE events SET voting_deadline = NOW() - interval '14 hours', starts_at = NOW() - interval '13 hours' WHERE id = $1", e.ID)
		require.NoError(t, err)

		due = append(due, e.ID)
	}

	notDue := f.voting(t, "p1")

	sweep := func(run func() (int, error)) int {
		var (
			wg sync.WaitGroup
			mu sync.Mutex
		)

		total := 0

		var errs []error

		// require can't stop the test from another goroutine, so errors are
		// collected and checked after Wait.
		for range 4 {
			wg.Go(func() {
				n, err := run()

				mu.Lock()
				defer mu.Unlock()

				total += n

				errs = append(errs, err)
			})
		}

		wg.Wait()

		for _, err := range errs {
			require.NoError(t, err)
		}

		return total
	}

	confirmed := sweep(func() (int, error) { return f.events.ConfirmDue(t.Context(), time.Now()) })
	assert.Equal(t, len(due), confirmed)

	for _, id := range due {
		assert.Equal(t, domain.StatusConfirmed, f.status(t, id))
	}

	assert.Equal(t, domain.StatusVoting, f.status(t, notDue.ID), "not past its deadline")

	closedCount := sweep(func() (int, error) {
		events, err := f.events.CloseDue(t.Context(), time.Now())

		return len(events), err
	})
	assert.Equal(t, len(due), closedCount)

	for _, id := range due {
		assert.Equal(t, domain.StatusClosed, f.status(t, id))
	}
}
