package application

import (
	"context"

	"github.com/SamPariat/roundup/services/event-service/internal/domain"
	"github.com/SamPariat/roundup/shared/domainerr"
	"go.uber.org/zap"
)

// StartVoting closes suggestions and opens voting. Needs at least one
// suggestion (ErrNoNominations otherwise).
func (u *EventUseCase) StartVoting(ctx context.Context, eventID, userID string) (*domain.EventDetail, error) {
	event, err := u.manage(ctx, eventID, userID, domain.StatusVoting)
	if err != nil {
		return nil, err
	}

	started, err := u.events.StartVoting(ctx, event.ID)
	if err != nil {
		return nil, err
	}

	if !started {
		// Still a draft after the guarded update means there was nothing to vote on.
		if current, err := u.events.Get(ctx, eventID); err == nil && current.Status == domain.StatusDraft {
			return nil, domain.ErrNoNominations
		}

		return nil, domainerr.New(domain.ErrInvalidTransition, "voting already started")
	}

	return u.Get(ctx, eventID, userID)
}

// CloseVoting closes voting now and confirms the winning place.
func (u *EventUseCase) CloseVoting(ctx context.Context, eventID, userID string) (*domain.EventDetail, error) {
	return u.step(ctx, eventID, userID, domain.StatusConfirmed, u.events.ConfirmVoting)
}

// StartNight marks the night as started ("Start the night").
func (u *EventUseCase) StartNight(ctx context.Context, eventID, userID string) (*domain.EventDetail, error) {
	return u.step(ctx, eventID, userID, domain.StatusActive, u.events.StartNight)
}

// Close ends the night ("Close the bill") and records the visit in the
// squad's history. A failure to record the visit is logged, not returned:
// the night is closed either way.
func (u *EventUseCase) Close(ctx context.Context, eventID, userID string) (*domain.EventDetail, error) {
	if _, err := u.manage(ctx, eventID, userID, domain.StatusClosed); err != nil {
		return nil, err
	}

	closed, err := u.events.Close(ctx, eventID)
	if err != nil {
		return nil, err
	}

	if closed == nil {
		return nil, domainerr.New(domain.ErrInvalidTransition, "the night isn't active")
	}

	u.recordVisit(ctx, *closed, userID)

	return u.Get(ctx, eventID, userID)
}

// ConfirmDue confirms every event whose voting deadline has passed, and
// returns how many. Safe to call repeatedly and concurrently: called by
// scheduler-service every minute.
func (u *EventUseCase) ConfirmDue(ctx context.Context) (int, error) {
	return u.events.ConfirmDue(ctx, u.now())
}

// CloseDue closes every night that started domain.AutoCloseAfter ago and
// nobody closed, records their visits as the event's creator, and returns
// how many it closed. Safe to call repeatedly and concurrently.
func (u *EventUseCase) CloseDue(ctx context.Context) (int, error) {
	closed, err := u.events.CloseDue(ctx, u.now())
	if err != nil {
		return 0, err
	}

	for _, e := range closed {
		u.recordVisit(ctx, e, e.CreatedBy)
	}

	return len(closed), nil
}

// step checks the caller may move the event to next, then runs the guarded
// update, which reports false if the status changed in between.
func (u *EventUseCase) step(ctx context.Context, eventID, userID string, next domain.Status,
	update func(ctx context.Context, eventID string) (bool, error),
) (*domain.EventDetail, error) {
	if _, err := u.manage(ctx, eventID, userID, next); err != nil {
		return nil, err
	}

	moved, err := update(ctx, eventID)
	if err != nil {
		return nil, err
	}

	if !moved {
		return nil, domainerr.New(domain.ErrInvalidTransition, "the event moved on")
	}

	return u.Get(ctx, eventID, userID)
}

// manage loads the event and checks that userID may manage it and that it
// may move to next.
func (u *EventUseCase) manage(ctx context.Context, eventID, userID string, next domain.Status) (*domain.Event, error) {
	event, members, err := u.access(ctx, eventID, userID)
	if err != nil {
		return nil, err
	}

	if err := u.require(memberOf(members, userID), event, domain.ActionManage); err != nil {
		return nil, err
	}

	if !event.Status.CanMoveTo(next) {
		return nil, domainerr.New(domain.ErrInvalidTransition, string(event.Status)+" → "+string(next))
	}

	return event, nil
}

// recordVisit adds a closed event's winning place to the squad's history,
// best-effort. An event closed without a winner has nothing to record.
func (u *EventUseCase) recordVisit(ctx context.Context, e domain.Event, userID string) {
	if e.Winner == nil {
		return
	}

	if err := u.venues.RecordVisit(ctx, e, userID); err != nil {
		u.log.Warn("failed to record the visit", zap.String("event_id", e.ID), zap.Error(err))
	}
}
