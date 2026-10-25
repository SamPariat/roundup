package domain

import "time"

// Status is where an event is in its lifecycle. Events only move forward,
// one step at a time: DRAFT → VOTING → CONFIRMED → ACTIVE → CLOSED.
type Status string

// Event statuses.
const (
	// StatusDraft is a new event: members RSVP and suggest places.
	StatusDraft Status = "DRAFT"

	// StatusVoting means suggestions are closed and members vote.
	StatusVoting Status = "VOTING"

	// StatusConfirmed means voting closed and the winning place is set.
	StatusConfirmed Status = "CONFIRMED"

	// StatusActive is the night itself ("Start the night"). Billing opens
	// the bill here once billing-service exists.
	StatusActive Status = "ACTIVE"

	// StatusClosed is a finished night. Nothing changes after this.
	StatusClosed Status = "CLOSED"
)

// nextStatus is the only move allowed out of each status. CLOSED has none.
var nextStatus = map[Status]Status{
	StatusDraft:     StatusVoting,
	StatusVoting:    StatusConfirmed,
	StatusConfirmed: StatusActive,
	StatusActive:    StatusClosed,
}

// Type is what kind of night the event is.
type Type string

// Event types.
const (
	// TypeDrinks is a night out for drinks.
	TypeDrinks Type = "DRINKS"

	// TypeDining is a night out for a meal.
	TypeDining Type = "DINING"

	// TypeBoth is dinner and drinks.
	TypeBoth Type = "BOTH"
)

// AutoCloseAfter is how long after it starts an event closes on its own, if
// nobody closed it by hand.
const AutoCloseAfter = 12 * time.Hour

// Event is one night out planned by a squad.
type Event struct {
	// ID is the event's UUID.
	ID string

	// SquadID is the squad planning it.
	SquadID string

	// Name is the display name. Never empty.
	Name string

	// Type is drinks, dining or both.
	Type Type

	// Status is the lifecycle step.
	Status Status

	// CreatedBy is the member who created it. They may since have left.
	CreatedBy string

	// StartsAt is when the night starts.
	StartsAt time.Time

	// VotingDeadline is when voting closes on its own. Never after StartsAt.
	VotingDeadline time.Time

	// Winner is the confirmed place; nil until voting closes.
	Winner *Nomination

	// ClosedAt is when the event closed; nil until then.
	ClosedAt *time.Time

	// CreatedAt is when the event was created.
	CreatedAt time.Time
}

// NewEvent carries what a member sends to create an event.
type NewEvent struct {
	// SquadID is the squad planning it.
	SquadID string

	// CreatedBy is the member creating it.
	CreatedBy string

	// Name is the display name, already trimmed and checked.
	Name string

	// Type is drinks, dining or both.
	Type Type

	// Schedule is the start and voting deadline, already checked.
	Schedule Schedule
}

// EventChanges carries an edit. Nil fields are left unchanged.
type EventChanges struct {
	// Name is the new display name.
	Name *string

	// Type is the new type.
	Type *Type

	// StartsAt is the new start. Without VotingDeadline, the deadline moves to it.
	StartsAt *time.Time

	// VotingDeadline is the new deadline.
	VotingDeadline *time.Time
}

// IsValid reports whether t is a supported event type.
func (t Type) IsValid() bool {
	return t == TypeDrinks || t == TypeDining || t == TypeBoth
}

// CanMoveTo reports whether an event in s may move to next.
func (s Status) CanMoveTo(next Status) bool {
	return nextStatus[s] == next
}

// IsEditable reports whether the name, type and times may still change.
func (s Status) IsEditable() bool {
	return s == StatusDraft || s == StatusVoting
}

// TakesRSVPs reports whether members may still answer.
func (s Status) TakesRSVPs() bool {
	return s != StatusClosed
}

// TakesNominations reports whether places may be suggested or withdrawn.
func (s Status) TakesNominations() bool {
	return s == StatusDraft
}

// TakesVotes reports whether members may vote.
func (s Status) TakesVotes() bool {
	return s == StatusVoting
}
