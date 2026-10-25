package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SamPariat/roundup/services/event-service/internal/domain"
	"github.com/SamPariat/roundup/services/event-service/internal/policy"
	"github.com/SamPariat/roundup/shared/authz"
	"github.com/SamPariat/roundup/shared/domainerr"
	"github.com/SamPariat/roundup/shared/pagination"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// event_usecase_test.go checks the use cases' own rules (membership,
// permissions from the real policy, status checks, best-effort visits) with
// in-memory fakes. The SQL behind each guarded write is tested in persistence.

const (
	squadID  = "squad-1"
	eventID  = "event-1"
	creator  = "u-creator"
	admin    = "u-admin"
	member   = "u-member"
	outsider = "u-outsider"
)

var squadMembers = []domain.SquadMember{
	{UserID: creator, Name: "Asha", Role: string(domain.RoleMember)},
	{UserID: admin, Name: "Ben", Role: string(domain.RoleAdmin)},
	{UserID: member, Name: "Chen", Role: string(domain.RoleMember)},
}

// fakeSquads answers membership from squadMembers.
type fakeSquads struct{}

func (fakeSquads) Members(_ context.Context, id, userID string) ([]domain.SquadMember, error) {
	for _, m := range squadMembers {
		if m.UserID == userID && id == squadID {
			return squadMembers, nil
		}
	}

	return nil, domainerr.New(domain.ErrSquadNotFound, id)
}

// fakeVenues returns a fixed place and records the visits it was asked for.
type fakeVenues struct {
	visits   []string // eventID/userID
	visitErr error
}

func (v *fakeVenues) Place(_ context.Context, placeID, _ string) (domain.Place, error) {
	if placeID == "unknown" {
		return domain.Place{}, domainerr.New(domain.ErrVenueNotFound, placeID)
	}

	return domain.Place{PlaceID: placeID, Name: "Place " + placeID}, nil
}

func (v *fakeVenues) RecordVisit(_ context.Context, e domain.Event, userID string) error {
	v.visits = append(v.visits, e.ID+"/"+userID)

	return v.visitErr
}

// fakeRepo holds one event. Guarded writes follow the same status rules as
// the SQL; calls counts every method call, to prove some never happen.
type fakeRepo struct {
	domain.EventRepository

	event       *domain.Event
	rsvps       []domain.RSVP
	votes       []domain.Vote
	nominations []domain.Nomination
	calls       int
	closeDue    []domain.Event
}

func (r *fakeRepo) Create(_ context.Context, e domain.NewEvent) (*domain.Event, error) {
	r.calls++
	r.event = &domain.Event{
		ID: eventID, SquadID: e.SquadID, Name: e.Name, Type: e.Type, Status: domain.StatusDraft,
		CreatedBy: e.CreatedBy, StartsAt: e.Schedule.StartsAt, VotingDeadline: e.Schedule.VotingDeadline,
	}

	return r.event, nil
}

func (r *fakeRepo) Get(_ context.Context, id string) (*domain.Event, error) {
	r.calls++
	if r.event == nil || id != r.event.ID {
		return nil, domainerr.New(domain.ErrEventNotFound, id)
	}

	e := *r.event

	return &e, nil
}

func (r *fakeRepo) RSVPs(context.Context, []string) ([]domain.RSVP, error) { return r.rsvps, nil }
func (r *fakeRepo) Votes(context.Context, []string) ([]domain.Vote, error) { return r.votes, nil }

func (r *fakeRepo) Nominations(context.Context, string) ([]domain.Nomination, error) {
	return r.nominations, nil
}

func (r *fakeRepo) AddNomination(_ context.Context, _ string, n domain.Nomination) (bool, error) {
	r.calls++
	if r.event.Status != domain.StatusDraft {
		return false, nil
	}

	r.nominations = append(r.nominations, n)

	return true, nil
}

func (r *fakeRepo) CastVote(_ context.Context, id, userID, placeID string) (bool, error) {
	r.calls++
	if _, ok := findNomination(r.nominations, placeID); !ok || r.event.Status != domain.StatusVoting {
		return false, nil
	}

	r.votes = append(r.votes, domain.Vote{EventID: id, UserID: userID, PlaceID: placeID})

	return true, nil
}

func (r *fakeRepo) Update(_ context.Context, e *domain.Event) (bool, error) {
	r.calls++
	if !r.event.Status.IsEditable() {
		return false, nil
	}

	*r.event = *e

	return true, nil
}

func (r *fakeRepo) StartNight(context.Context, string) (bool, error) {
	r.calls++
	if r.event.Status != domain.StatusConfirmed {
		return false, nil
	}

	r.event.Status = domain.StatusActive

	return true, nil
}

func (r *fakeRepo) Close(context.Context, string) (*domain.Event, error) {
	r.calls++
	if r.event.Status != domain.StatusActive {
		return nil, nil
	}

	r.event.Status = domain.StatusClosed
	e := *r.event

	return &e, nil
}

func (r *fakeRepo) CloseDue(context.Context, time.Time) ([]domain.Event, error) {
	return r.closeDue, nil
}

func newUseCase(t *testing.T, repo *fakeRepo, venues *fakeVenues) *EventUseCase {
	t.Helper()

	authorizer, err := authz.New[domain.Role, domain.Action](policy.CSV)
	require.NoError(t, err)

	return NewEventUseCase(repo, fakeSquads{}, venues, authorizer, zap.NewNop())
}

// withEvent returns a repo holding an event created by creator, in status.
func withEvent(status domain.Status) *fakeRepo {
	return &fakeRepo{event: &domain.Event{
		ID: eventID, SquadID: squadID, Name: "Friday drinks", Type: domain.TypeDrinks, Status: status, CreatedBy: creator,
		StartsAt: time.Now().Add(48 * time.Hour), VotingDeadline: time.Now().Add(24 * time.Hour),
		Winner: &domain.Nomination{PlaceID: "p1"},
	}}
}

func TestNonMemberNeverReachesTheRepository(t *testing.T) {
	repo := &fakeRepo{}
	uc := newUseCase(t, repo, &fakeVenues{})
	deadline := time.Now().Add(time.Hour)

	_, err := uc.Create(t.Context(), CreateEventCommand{
		SquadID: squadID, UserID: outsider, Name: "x", Type: domain.TypeBoth,
		StartsAt: time.Now().Add(48 * time.Hour), VotingDeadline: &deadline,
	})
	require.ErrorIs(t, err, domain.ErrSquadNotFound)

	_, err = uc.List(t.Context(), squadID, outsider, pagination.Page{Limit: 10})
	require.ErrorIs(t, err, domain.ErrSquadNotFound)
	assert.Zero(t, repo.calls)
}

func TestMovingTheStartMovesTheDeadline(t *testing.T) {
	repo := withEvent(domain.StatusVoting)
	uc := newUseCase(t, repo, &fakeVenues{})
	later := time.Now().Add(96 * time.Hour)

	d, err := uc.Update(t.Context(), eventID, creator, domain.EventChanges{StartsAt: &later})
	require.NoError(t, err)
	assert.Equal(t, later, d.VotingDeadline, "voting stays open until the new start")

	picked := time.Now().Add(72 * time.Hour)
	evenLater := time.Now().Add(120 * time.Hour)

	d, err = uc.Update(t.Context(), eventID, creator, domain.EventChanges{StartsAt: &evenLater, VotingDeadline: &picked})
	require.NoError(t, err)
	assert.Equal(t, picked, d.VotingDeadline, "a deadline sent with the start is kept")
}

func TestOutsiderSeesNoEvent(t *testing.T) {
	repo := withEvent(domain.StatusDraft)
	uc := newUseCase(t, repo, &fakeVenues{})

	_, err := uc.Get(t.Context(), eventID, outsider)
	require.ErrorIs(t, err, domain.ErrEventNotFound, "an outsider can't tell the event exists")
}

func TestOnlyCreatorAndAdminsManage(t *testing.T) {
	tests := []struct {
		user    string
		wantErr error
	}{
		{member, domain.ErrForbidden},
		{creator, nil},
		{admin, nil},
	}

	for _, tt := range tests {
		t.Run(tt.user, func(t *testing.T) {
			repo := withEvent(domain.StatusConfirmed)
			uc := newUseCase(t, repo, &fakeVenues{})

			_, err := uc.StartNight(t.Context(), eventID, tt.user)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, domain.StatusConfirmed, repo.event.Status)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, domain.StatusActive, repo.event.Status)
		})
	}
}

func TestStepsRefuseTheWrongStatus(t *testing.T) {
	repo := withEvent(domain.StatusDraft)
	uc := newUseCase(t, repo, &fakeVenues{})

	_, err := uc.StartNight(t.Context(), eventID, creator)
	require.ErrorIs(t, err, domain.ErrInvalidTransition)

	_, err = uc.Close(t.Context(), eventID, creator)
	require.ErrorIs(t, err, domain.ErrInvalidTransition)
}

func TestNominateOnlyWhileDraft(t *testing.T) {
	repo := withEvent(domain.StatusVoting)
	uc := newUseCase(t, repo, &fakeVenues{})

	_, err := uc.Nominate(t.Context(), eventID, member, "p1")
	require.ErrorIs(t, err, domain.ErrInvalidTransition)
	assert.Empty(t, repo.nominations)

	repo.event.Status = domain.StatusDraft

	_, err = uc.Nominate(t.Context(), eventID, member, "unknown")
	require.ErrorIs(t, err, domain.ErrVenueNotFound)

	_, err = uc.Nominate(t.Context(), eventID, member, " ")
	require.ErrorIs(t, err, domain.ErrInvalidEvent, "a blank place never reaches venue-service")

	d, err := uc.Nominate(t.Context(), eventID, member, "p1")
	require.NoError(t, err)
	require.Len(t, d.Nominations, 1)
	assert.Equal(t, "Place p1", d.Nominations[0].Name, "the snapshot comes from venue-service")
	assert.Equal(t, "Chen", d.Nominations[0].NominatedByName)
	assert.True(t, d.Nominations[0].CanRemove, "you may withdraw your own suggestion")
}

func TestRemovingSomeoneElsesNomination(t *testing.T) {
	repo := withEvent(domain.StatusDraft)
	repo.nominations = []domain.Nomination{{PlaceID: "p1", NominatedBy: admin}}
	uc := newUseCase(t, repo, &fakeVenues{})

	_, err := uc.RemoveNomination(t.Context(), eventID, member, "p1")
	require.ErrorIs(t, err, domain.ErrForbidden)

	_, err = uc.RemoveNomination(t.Context(), eventID, member, "p9")
	require.ErrorIs(t, err, domain.ErrNominationNotFound)
}

func TestVoteForAPlaceNobodySuggested(t *testing.T) {
	repo := withEvent(domain.StatusVoting)
	repo.nominations = []domain.Nomination{{PlaceID: "p1", NominatedBy: creator}}
	uc := newUseCase(t, repo, &fakeVenues{})

	_, err := uc.Vote(t.Context(), eventID, member, "p9")
	require.ErrorIs(t, err, domain.ErrNotNominated)

	d, err := uc.Vote(t.Context(), eventID, member, "p1")
	require.NoError(t, err)
	assert.Equal(t, "p1", d.MyVote)
	assert.Equal(t, 1, d.VotedCount)
	assert.Equal(t, []domain.SquadMember{squadMembers[2]}, d.Nominations[0].Voters)
}

func TestCloseRecordsTheVisitBestEffort(t *testing.T) {
	for _, visitErr := range []error{nil, errors.New("venue-service down")} {
		repo := withEvent(domain.StatusActive)
		venues := &fakeVenues{visitErr: visitErr}
		uc := newUseCase(t, repo, venues)

		_, err := uc.Close(t.Context(), eventID, admin)
		require.NoError(t, err, "a failed visit doesn't fail the close")
		assert.Equal(t, domain.StatusClosed, repo.event.Status)
		assert.Equal(t, []string{eventID + "/" + admin}, venues.visits)
	}
}

func TestCloseDueRecordsVisitsAsTheCreator(t *testing.T) {
	repo := &fakeRepo{closeDue: []domain.Event{
		{ID: "e1", CreatedBy: creator, Winner: &domain.Nomination{PlaceID: "p1"}},
		{ID: "e2", CreatedBy: admin}, // closed without a winner: nothing to record
	}}
	venues := &fakeVenues{}
	uc := newUseCase(t, repo, venues)

	n, err := uc.CloseDue(t.Context())
	require.NoError(t, err)
	assert.Equal(t, 2, n)
	assert.Equal(t, []string{"e1/" + creator}, venues.visits)
}

func TestDetailPermissionsFollowRoleAndStatus(t *testing.T) {
	repo := withEvent(domain.StatusDraft)
	repo.nominations = []domain.Nomination{{PlaceID: "p1", NominatedBy: member}}
	uc := newUseCase(t, repo, &fakeVenues{})

	asMember, err := uc.Get(t.Context(), eventID, member)
	require.NoError(t, err)
	assert.Equal(t, domain.Permissions{CanRSVP: true, CanNominate: true}, asMember.Permissions)

	asCreator, err := uc.Get(t.Context(), eventID, creator)
	require.NoError(t, err)
	assert.Equal(t, domain.Permissions{CanEdit: true, CanRSVP: true, CanNominate: true, CanStartVoting: true}, asCreator.Permissions)
	assert.True(t, asCreator.Nominations[0].CanRemove, "the creator may withdraw anyone's suggestion")
	assert.Equal(t, "Asha", asCreator.CreatedByName)
}

func TestSummariesCountOnlyCurrentMembers(t *testing.T) {
	events := []domain.Event{{ID: "e1"}, {ID: "e2"}}
	rsvps := []domain.RSVP{
		{EventID: "e1", UserID: creator, Status: domain.RSVPGoing},
		{EventID: "e1", UserID: admin, Status: domain.RSVPMaybe},
		{EventID: "e1", UserID: "u-left", Status: domain.RSVPGoing}, // left the squad
	}
	votes := []domain.Vote{{EventID: "e2", UserID: member, PlaceID: "p1"}}

	got := summarise(events, squadMembers, member, rsvps, votes, map[string]int{"e1": 3})

	assert.Equal(t, domain.RSVPCounts{Going: 1, Maybe: 1, Pending: 1}, got[0].RSVPCounts)
	assert.Equal(t, []domain.SquadMember{squadMembers[0]}, got[0].GoingPreview)
	assert.Equal(t, 3, got[0].NominationCount)
	assert.False(t, got[0].HasVoted)

	assert.Equal(t, domain.RSVPCounts{Pending: 3}, got[1].RSVPCounts)
	assert.Empty(t, got[1].GoingPreview)
	assert.True(t, got[1].HasVoted)
}
