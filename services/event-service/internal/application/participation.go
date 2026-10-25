package application

import (
	"context"
	"strings"

	"github.com/SamPariat/roundup/services/event-service/internal/domain"
	"github.com/SamPariat/roundup/shared/domainerr"
)

// SetRSVP stores the caller's answer until the event closes, and returns the
// updated event.
func (u *EventUseCase) SetRSVP(ctx context.Context, eventID, userID string, status domain.RSVPStatus) (*domain.EventDetail, error) {
	if !status.IsValid() {
		return nil, domain.ErrInvalidRSVP
	}

	event, members, err := u.access(ctx, eventID, userID)
	if err != nil {
		return nil, err
	}

	if err := u.require(memberOf(members, userID), event, domain.ActionRSVP); err != nil {
		return nil, err
	}

	if !event.Status.TakesRSVPs() {
		return nil, domainerr.New(domain.ErrInvalidTransition, "the night is over")
	}

	set, err := u.events.SetRSVP(ctx, eventID, userID, status)
	if err != nil {
		return nil, err
	}

	if !set {
		return nil, domainerr.New(domain.ErrInvalidTransition, "the night is over")
	}

	return u.Get(ctx, eventID, userID)
}

// Nominate suggests a place while the event is a draft, keeping a snapshot
// of it from venue-service, and returns the updated event. Returns
// ErrVenueNotFound for an unknown place and ErrAlreadyNominated for a repeat.
func (u *EventUseCase) Nominate(ctx context.Context, eventID, userID, placeID string) (*domain.EventDetail, error) {
	event, members, err := u.access(ctx, eventID, userID)
	if err != nil {
		return nil, err
	}

	if err := u.require(memberOf(members, userID), event, domain.ActionNominate); err != nil {
		return nil, err
	}

	if !event.Status.TakesNominations() {
		return nil, domainerr.New(domain.ErrInvalidTransition, "suggestions are closed")
	}

	if strings.TrimSpace(placeID) == "" {
		return nil, domainerr.New(domain.ErrInvalidEvent, "placeID is required")
	}

	place, err := u.venues.Place(ctx, placeID, userID)
	if err != nil {
		return nil, err
	}

	place.PlaceID = placeID

	added, err := u.events.AddNomination(ctx, eventID, domain.Nomination{Place: place, NominatedBy: userID})
	if err != nil {
		return nil, err
	}

	if !added {
		return nil, domainerr.New(domain.ErrInvalidTransition, "suggestions are closed")
	}

	return u.Get(ctx, eventID, userID)
}

// RemoveNomination withdraws a suggestion while the event is a draft: your
// own, or anyone's for the creator and admins. Returns the updated event.
func (u *EventUseCase) RemoveNomination(ctx context.Context, eventID, userID, placeID string) (*domain.EventDetail, error) {
	event, members, err := u.access(ctx, eventID, userID)
	if err != nil {
		return nil, err
	}

	nominations, err := u.events.Nominations(ctx, eventID)
	if err != nil {
		return nil, err
	}

	nomination, ok := findNomination(nominations, placeID)
	if !ok {
		return nil, domainerr.New(domain.ErrNominationNotFound, placeID)
	}

	roles := domain.RolesOf(memberOf(members, userID), event)
	if !canRemove(u.authorizer, roles, nomination, userID) {
		return nil, domainerr.New(domain.ErrForbidden, string(domain.ActionRemoveNomination))
	}

	if !event.Status.TakesNominations() {
		return nil, domainerr.New(domain.ErrInvalidTransition, "suggestions are closed")
	}

	removed, err := u.events.RemoveNomination(ctx, eventID, placeID)
	if err != nil {
		return nil, err
	}

	if !removed {
		return nil, domainerr.New(domain.ErrInvalidTransition, "suggestions are closed")
	}

	return u.Get(ctx, eventID, userID)
}

// Vote stores or changes the caller's vote while voting is open, and returns
// the updated event. Returns ErrNotNominated for a place nobody suggested.
func (u *EventUseCase) Vote(ctx context.Context, eventID, userID, placeID string) (*domain.EventDetail, error) {
	event, members, err := u.access(ctx, eventID, userID)
	if err != nil {
		return nil, err
	}

	if err := u.require(memberOf(members, userID), event, domain.ActionVote); err != nil {
		return nil, err
	}

	if !event.Status.TakesVotes() {
		return nil, domainerr.New(domain.ErrInvalidTransition, "voting isn't open")
	}

	cast, err := u.events.CastVote(ctx, eventID, userID, placeID)
	if err != nil {
		return nil, err
	}

	// The status was VOTING a moment ago, so a refused vote almost always
	// means the place wasn't suggested.
	if !cast {
		return nil, domainerr.New(domain.ErrNotNominated, placeID)
	}

	return u.Get(ctx, eventID, userID)
}

// canRemove reports whether a caller holding roles may withdraw n: anyone's
// with remove_nomination, their own with remove_own_nomination.
func canRemove(authorizer domain.Authorizer, roles []domain.Role, n domain.Nomination, userID string) bool {
	if domain.Allowed(authorizer, roles, domain.ActionRemoveNomination) {
		return true
	}

	return n.NominatedBy == userID && domain.Allowed(authorizer, roles, domain.ActionRemoveOwnNomination)
}

func findNomination(nominations []domain.Nomination, placeID string) (domain.Nomination, bool) {
	for _, n := range nominations {
		if n.PlaceID == placeID {
			return n, true
		}
	}

	return domain.Nomination{}, false
}
