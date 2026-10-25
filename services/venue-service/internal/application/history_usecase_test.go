package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SamPariat/roundup/services/venue-service/internal/domain"
	"github.com/SamPariat/roundup/shared/pagination"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// history_usecase_test.go tests HistoryUseCase using the fakeVenueRepo defined in
// favorite_usecase_test.go. No real database calls are made.

func TestRecordVisitSuccess(t *testing.T) {
	uc := NewHistoryUseCase(&fakeVenueRepo{}, fakeSquadDirectory{}, fakeVenueLookup{})

	err := uc.Record(t.Context(), domain.RecordVisitCommand{
		SquadID:         "squad-id",
		EventID:         "event-id",
		PlaceID:         "place-id",
		VisitedAt:       time.Now(),
		AvgSpendInPaise: 10,
	})

	require.NoError(t, err)
}

func TestRecordVisitUnknownVenue(t *testing.T) {
	uc := NewHistoryUseCase(mustNotTouchRepo(t), fakeSquadDirectory{}, fakeVenueLookup{err: domain.ErrVenueNotFound})

	err := uc.Record(t.Context(), domain.RecordVisitCommand{PlaceID: "nope"})

	assert.ErrorIs(t, err, domain.ErrVenueNotFound)
}

func TestRecordVisitPropagatesError(t *testing.T) {
	repoErr := errors.New("db write failed")

	uc := NewHistoryUseCase(&fakeVenueRepo{
		recordVisitFn: func(_ context.Context, _ domain.RecordVisitCommand, _ domain.VenueSnapshot) error {
			return repoErr
		},
	}, fakeSquadDirectory{}, fakeVenueLookup{})

	err := uc.Record(t.Context(), domain.RecordVisitCommand{})
	assert.ErrorIs(t, err, repoErr)
}

func TestGetVisitHistoryReturnsSummaries(t *testing.T) {
	summaries := []domain.VisitSummary{
		{PlaceID: "abc", Name: "The Rooftop", VisitCount: 3, LastVisitedAt: time.Now()},
	}

	uc := NewHistoryUseCase(&fakeVenueRepo{
		getVisitHistoryFn: func(_ context.Context, _ string, _ pagination.Page) ([]domain.VisitSummary, string, error) {
			return summaries, "", nil
		},
	}, fakeSquadDirectory{}, fakeVenueLookup{})

	got, _, err := uc.List(t.Context(), "squad-1", "user-1", firstPage)
	require.NoError(t, err)
	assert.Equal(t, summaries, got)
}

func TestGetVisitHistoryPropagatesError(t *testing.T) {
	repoErr := errors.New("db read failed")

	uc := NewHistoryUseCase(&fakeVenueRepo{
		getVisitHistoryFn: func(_ context.Context, _ string, _ pagination.Page) ([]domain.VisitSummary, string, error) {
			return nil, "", repoErr
		},
	}, fakeSquadDirectory{}, fakeVenueLookup{})

	_, _, err := uc.List(t.Context(), "squad-1", "user-1", firstPage)
	assert.ErrorIs(t, err, repoErr)
}
