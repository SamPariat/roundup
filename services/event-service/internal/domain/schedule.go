package domain

import (
	"time"

	"github.com/SamPariat/roundup/shared/domainerr"
)

// Schedule is when an event starts and when its voting closes.
type Schedule struct {
	// StartsAt is when the night starts. Always in the future when checked.
	StartsAt time.Time

	// VotingDeadline is when voting closes. Never after StartsAt.
	VotingDeadline time.Time
}

// NewSchedule checks and completes the times a member sent, as of now. A nil
// deadline means the member didn't pick one, and voting stays open until the
// event starts. Returns ErrInvalidEvent, with the reason, for times that can't
// work. Used both when creating an event and when editing its times.
func NewSchedule(now, startsAt time.Time, deadline *time.Time) (Schedule, error) {
	if !startsAt.After(now) {
		return Schedule{}, domainerr.New(ErrInvalidEvent, "startsAt must be in the future")
	}

	votingDeadline := startsAt
	if deadline != nil {
		votingDeadline = *deadline
	}

	if !votingDeadline.After(now) {
		return Schedule{}, domainerr.New(ErrInvalidEvent, "votingDeadline must be in the future")
	}

	if votingDeadline.After(startsAt) {
		return Schedule{}, domainerr.New(ErrInvalidEvent, "votingDeadline must not be after startsAt")
	}

	return Schedule{StartsAt: startsAt, VotingDeadline: votingDeadline}, nil
}
