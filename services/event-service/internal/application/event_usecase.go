package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/SamPariat/roundup/services/event-service/internal/domain"
	"github.com/SamPariat/roundup/shared/domainerr"
	"github.com/SamPariat/roundup/shared/pagination"
	"go.uber.org/zap"
)

// EventUseCase plans nights out: creating and listing events, RSVPs,
// suggestions and votes (participation.go), and moving events through their
// lifecycle (lifecycle.go). Every method acting on an event first checks that
// the caller is in its squad; a non-member gets ErrEventNotFound, the same as
// for an event that doesn't exist.
type EventUseCase struct {
	events     domain.EventRepository
	squads     domain.SquadDirectory
	venues     domain.VenueDirectory
	authorizer domain.Authorizer
	log        *zap.Logger
	now        func() time.Time
}

// CreateEventCommand carries what a member sends to create an event.
type CreateEventCommand struct {
	SquadID  string
	UserID   string
	Name     string
	Type     domain.Type
	StartsAt time.Time

	// VotingDeadline is nil when the member didn't pick one.
	VotingDeadline *time.Time
}

// SquadEvents is one page of a squad's events.
type SquadEvents struct {
	Events     []domain.EventSummary
	NextCursor string

	// OpenCount counts all the squad's events that aren't closed, for the
	// "Events (2)" tab label.
	OpenCount int
}

// NewEventUseCase constructs an EventUseCase. log records failures of
// best-effort calls, such as recording a visit when a night closes.
func NewEventUseCase(events domain.EventRepository, squads domain.SquadDirectory, venues domain.VenueDirectory,
	authorizer domain.Authorizer, log *zap.Logger,
) *EventUseCase {
	return &EventUseCase{events: events, squads: squads, venues: venues, authorizer: authorizer, log: log, now: time.Now}
}

// Create creates a DRAFT event in the squad. Returns ErrSquadNotFound for a
// non-member and ErrInvalidEvent for a blank name, unknown type or bad times.
func (u *EventUseCase) Create(ctx context.Context, cmd CreateEventCommand) (*domain.Event, error) {
	members, err := u.squads.Members(ctx, cmd.SquadID, cmd.UserID)
	if err != nil {
		return nil, err
	}

	if err := u.require(memberOf(members, cmd.UserID), nil, domain.ActionCreate); err != nil {
		return nil, err
	}

	name, err := validName(cmd.Name)
	if err != nil {
		return nil, err
	}

	if !cmd.Type.IsValid() {
		return nil, domainerr.New(domain.ErrInvalidEvent, "type must be DRINKS, DINING or BOTH")
	}

	schedule, err := domain.NewSchedule(u.now(), cmd.StartsAt, cmd.VotingDeadline)
	if err != nil {
		return nil, err
	}

	return u.events.Create(ctx, domain.NewEvent{
		SquadID:   cmd.SquadID,
		CreatedBy: cmd.UserID,
		Name:      name,
		Type:      cmd.Type,
		Schedule:  schedule,
	})
}

// List returns one page of the squad's events with their RSVP counts, going
// preview and the caller's answer and vote. Returns ErrSquadNotFound for a
// non-member.
func (u *EventUseCase) List(ctx context.Context, squadID, userID string, page pagination.Page) (*SquadEvents, error) {
	members, err := u.squads.Members(ctx, squadID, userID)
	if err != nil {
		return nil, err
	}

	events, next, err := u.events.ListForSquad(ctx, squadID, page)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(events))
	for _, e := range events {
		ids = append(ids, e.ID)
	}

	rsvps, err := u.events.RSVPs(ctx, ids)
	if err != nil {
		return nil, err
	}

	votes, err := u.events.Votes(ctx, ids)
	if err != nil {
		return nil, err
	}

	nominationCounts, err := u.events.NominationCounts(ctx, ids)
	if err != nil {
		return nil, err
	}

	openCount, err := u.events.CountOpen(ctx, squadID)
	if err != nil {
		return nil, err
	}

	return &SquadEvents{
		Events:     summarise(events, members, userID, rsvps, votes, nominationCounts),
		NextCursor: next,
		OpenCount:  openCount,
	}, nil
}

// Get returns everything the event screen shows. Returns ErrEventNotFound for
// an unknown event or a non-member.
func (u *EventUseCase) Get(ctx context.Context, eventID, userID string) (*domain.EventDetail, error) {
	event, members, err := u.access(ctx, eventID, userID)
	if err != nil {
		return nil, err
	}

	ids := []string{event.ID}

	rsvps, err := u.events.RSVPs(ctx, ids)
	if err != nil {
		return nil, err
	}

	votes, err := u.events.Votes(ctx, ids)
	if err != nil {
		return nil, err
	}

	nominations, err := u.events.Nominations(ctx, event.ID)
	if err != nil {
		return nil, err
	}

	return detail(event, members, userID, rsvps, votes, nominations, u.authorizer), nil
}

// Update changes the name, type or times while the event is DRAFT or VOTING.
// Moving StartsAt without a new deadline moves the deadline to the new start.
func (u *EventUseCase) Update(ctx context.Context, eventID, userID string, changes domain.EventChanges) (*domain.EventDetail, error) {
	event, members, err := u.access(ctx, eventID, userID)
	if err != nil {
		return nil, err
	}

	if err := u.require(memberOf(members, userID), event, domain.ActionEdit); err != nil {
		return nil, err
	}

	if !event.Status.IsEditable() {
		return nil, domainerr.New(domain.ErrInvalidTransition, "the place is already confirmed")
	}

	if changes.Name != nil {
		if event.Name, err = validName(*changes.Name); err != nil {
			return nil, err
		}
	}

	if changes.Type != nil {
		if !changes.Type.IsValid() {
			return nil, domainerr.New(domain.ErrInvalidEvent, "type must be DRINKS, DINING or BOTH")
		}

		event.Type = *changes.Type
	}

	if changes.StartsAt != nil || changes.VotingDeadline != nil {
		startsAt := event.StartsAt
		if changes.StartsAt != nil {
			startsAt = *changes.StartsAt
		}

		schedule, err := domain.NewSchedule(u.now(), startsAt, changes.VotingDeadline)
		if err != nil {
			return nil, err
		}

		event.StartsAt, event.VotingDeadline = schedule.StartsAt, schedule.VotingDeadline
	}

	updated, err := u.events.Update(ctx, event)
	if err != nil {
		return nil, err
	}

	if !updated {
		return nil, domainerr.New(domain.ErrInvalidTransition, "the place is already confirmed")
	}

	return u.Get(ctx, eventID, userID)
}

// access loads the event and its squad's members, as seen by userID. A
// non-member gets ErrEventNotFound, so they learn nothing about the event.
func (u *EventUseCase) access(ctx context.Context, eventID, userID string) (*domain.Event, []domain.SquadMember, error) {
	event, err := u.events.Get(ctx, eventID)
	if err != nil {
		return nil, nil, err
	}

	members, err := u.squads.Members(ctx, event.SquadID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrSquadNotFound) {
			return nil, nil, domainerr.New(domain.ErrEventNotFound, eventID)
		}

		return nil, nil, err
	}

	return event, members, nil
}

// require returns ErrForbidden unless member may take action on event (nil
// for squad-level actions).
func (u *EventUseCase) require(member domain.SquadMember, event *domain.Event, action domain.Action) error {
	if !domain.Allowed(u.authorizer, domain.RolesOf(member, event), action) {
		return domainerr.New(domain.ErrForbidden, string(action))
	}

	return nil
}

// memberOf returns userID's entry in members. Members always includes the
// caller, since user-service only answers members.
func memberOf(members []domain.SquadMember, userID string) domain.SquadMember {
	for _, m := range members {
		if m.UserID == userID {
			return m
		}
	}

	return domain.SquadMember{UserID: userID}
}

// validName trims name and rejects a blank one.
func validName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", domainerr.New(domain.ErrInvalidEvent, "name must not be blank")
	}

	return name, nil
}
