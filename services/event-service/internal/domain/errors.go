package domain

import (
	"errors"

	"github.com/SamPariat/roundup/shared/userclient"
)

// Sentinel errors are used as stable identities for errors.Is comparisons.
// Adapters and use cases add context with domainerr.New.
var (
	// ErrEventNotFound is returned when the event does not exist, or the
	// caller is not in its squad. The two look the same from outside.
	ErrEventNotFound = errors.New("event not found")

	// ErrInvalidEvent is returned for a blank name, an unknown type or
	// times that can't work.
	ErrInvalidEvent = errors.New("invalid event")

	// ErrInvalidRSVP is returned for an answer other than GOING, MAYBE or
	// NOT_GOING.
	ErrInvalidRSVP = errors.New("rsvp must be GOING, MAYBE or NOT_GOING")

	// ErrInvalidTransition is returned when the event's status doesn't
	// allow the step, e.g. voting on a draft or editing a closed event.
	ErrInvalidTransition = errors.New("not allowed while the event is in this state")

	// ErrForbidden is returned when the caller's roles don't allow the action.
	ErrForbidden = errors.New("you are not allowed to do this for this event")

	// ErrNoNominations is returned when voting starts with nothing to vote on.
	ErrNoNominations = errors.New("suggest at least one place before voting starts")

	// ErrAlreadyNominated is returned when the place was already suggested.
	ErrAlreadyNominated = errors.New("place already suggested")

	// ErrNominationNotFound is returned when withdrawing a place that wasn't suggested.
	ErrNominationNotFound = errors.New("place not suggested")

	// ErrNotNominated is returned for a vote on a place that wasn't suggested.
	ErrNotNominated = errors.New("vote for one of the suggested places")

	// ErrVenueNotFound is returned when venue-service doesn't know the place.
	ErrVenueNotFound = errors.New("venue not found")

	// ErrVenueServiceUnavailable is returned when venue-service can't answer.
	ErrVenueServiceUnavailable = errors.New("venue service unavailable")

	// ErrSquadNotFound is returned when the squad does not exist or the
	// caller is not a member.
	ErrSquadNotFound = userclient.ErrSquadNotFound

	// ErrUserServiceUnavailable is returned when user-service can't answer a
	// membership check.
	ErrUserServiceUnavailable = userclient.ErrUserServiceUnavailable
)
