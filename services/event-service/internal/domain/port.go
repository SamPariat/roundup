package domain

import (
	"context"
	"time"

	"github.com/SamPariat/roundup/shared/pagination"
	"github.com/SamPariat/roundup/shared/userclient"
)

// EventRepository stores events, RSVPs, nominations and votes. Every write
// that depends on the event's status checks it in the same statement, so a
// concurrent change can't slip between a read and a write; those methods
// return false when the status no longer allowed the write.
type EventRepository interface {
	// Create stores a new DRAFT event.
	Create(ctx context.Context, e NewEvent) (*Event, error)

	// Get returns the event, or ErrEventNotFound (also for a malformed ID).
	Get(ctx context.Context, eventID string) (*Event, error)

	// Update saves e's name, type and times if it is still DRAFT or VOTING.
	Update(ctx context.Context, e *Event) (bool, error)

	// ListForSquad returns one page of the squad's events, open ones first
	// by start time, then closed ones most recent first, and the next
	// page's cursor ("" on the last). Returns pagination.ErrInvalid for a
	// cursor it did not issue.
	ListForSquad(ctx context.Context, squadID string, page pagination.Page) ([]Event, string, error)

	// CountOpen counts the squad's events that aren't CLOSED.
	CountOpen(ctx context.Context, squadID string) (int, error)

	// RSVPs returns every answer to the given events.
	RSVPs(ctx context.Context, eventIDs []string) ([]RSVP, error)

	// Votes returns every vote cast in the given events.
	Votes(ctx context.Context, eventIDs []string) ([]Vote, error)

	// Nominations returns the event's suggested places, earliest first.
	Nominations(ctx context.Context, eventID string) ([]Nomination, error)

	// NominationCounts returns how many places each event has, by event ID.
	NominationCounts(ctx context.Context, eventIDs []string) (map[string]int, error)

	// SetRSVP stores or changes userID's answer unless the event is CLOSED.
	SetRSVP(ctx context.Context, eventID, userID string, status RSVPStatus) (bool, error)

	// AddNomination stores n if the event is DRAFT. Returns
	// ErrAlreadyNominated if the place was already suggested.
	AddNomination(ctx context.Context, eventID string, n Nomination) (bool, error)

	// RemoveNomination deletes the place if the event is DRAFT.
	RemoveNomination(ctx context.Context, eventID, placeID string) (bool, error)

	// CastVote stores or changes userID's vote if the event is VOTING and
	// placeID is one of its nominations.
	CastVote(ctx context.Context, eventID, userID, placeID string) (bool, error)

	// StartVoting moves a DRAFT event with at least one nomination to VOTING.
	StartVoting(ctx context.Context, eventID string) (bool, error)

	// ConfirmVoting moves a VOTING event to CONFIRMED and sets its winner:
	// the most votes, a tie going to the earliest nomination.
	ConfirmVoting(ctx context.Context, eventID string) (bool, error)

	// ConfirmDue does ConfirmVoting for every VOTING event whose deadline
	// is at or before now, and returns how many it confirmed.
	ConfirmDue(ctx context.Context, now time.Time) (int, error)

	// StartNight moves a CONFIRMED event to ACTIVE.
	StartNight(ctx context.Context, eventID string) (bool, error)

	// Close moves an ACTIVE event to CLOSED and returns it, or nil if it
	// wasn't ACTIVE.
	Close(ctx context.Context, eventID string) (*Event, error)

	// CloseDue closes every CONFIRMED or ACTIVE event that started
	// AutoCloseAfter or more before now, and returns them.
	CloseDue(ctx context.Context, now time.Time) ([]Event, error)
}

// SquadMember is a squad member as user-service describes them.
type SquadMember = userclient.SquadMember

// SquadDirectory answers who belongs to a squad. Squads live in
// user-service; implemented by userclient.UserServiceClient.
type SquadDirectory interface {
	// Members returns squadID's members if userID is one of them,
	// ErrSquadNotFound if not, and ErrUserServiceUnavailable if user-service
	// cannot answer.
	Members(ctx context.Context, squadID, userID string) ([]SquadMember, error)
}

// VenueDirectory is venue-service, called as the member acting.
type VenueDirectory interface {
	// Place returns placeID's display fields. Returns ErrVenueNotFound for
	// an unknown place and ErrVenueServiceUnavailable otherwise.
	Place(ctx context.Context, placeID, userID string) (Place, error)

	// RecordVisit adds the event's place to the squad's history.
	RecordVisit(ctx context.Context, e Event, userID string) error
}
