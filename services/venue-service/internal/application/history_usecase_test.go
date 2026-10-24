package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SamPariat/roundup/services/venue-service/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// history_usecase_test.go tests HistoryUseCase using the fakeVenueRepo defined in
// favorite_usecase_test.go. No real database calls are made.

func TestRecordVisitSuccess(t *testing.T) {
	uc := NewHistoryUseCase(&fakeVenueRepo{}, fakeSquadDirectory{})

	err := uc.Record(t.Context(), domain.RecordVisitCommand{
		SquadID:         "squad-id",
		EventID:         "event-id",
		PlaceID:         "place-id",
		VisitedAt:       time.Now(),
		AvgSpendInPaise: 10,
	})

	require.NoError(t, err)
}

func TestRecordVisitPropagatesError(t *testing.T) {
	repoErr := errors.New("db write failed")

	uc := NewHistoryUseCase(&fakeVenueRepo{
		recordVisitFn: func(_ context.Context, _ domain.RecordVisitCommand) error {
			return repoErr
		},
	}, fakeSquadDirectory{})

	err := uc.Record(t.Context(), domain.RecordVisitCommand{})
	assert.ErrorIs(t, err, repoErr)
}

func TestGetVisitHistoryReturnsSummaries(t *testing.T) {
	summaries := []domain.VisitSummary{
		{PlaceID: "abc", Name: "The Rooftop", VisitCount: 3, LastVisitedAt: time.Now()},
	}

	uc := NewHistoryUseCase(&fakeVenueRepo{
		getVisitHistoryFn: func(_ context.Context, _ string) ([]domain.VisitSummary, error) {
			return summaries, nil
		},
	}, fakeSquadDirectory{})

	got, err := uc.List(t.Context(), "squad-1", "user-1")
	require.NoError(t, err)
	assert.Equal(t, summaries, got)
}

func TestGetVisitHistoryPropagatesError(t *testing.T) {
	repoErr := errors.New("db read failed")

	uc := NewHistoryUseCase(&fakeVenueRepo{
		getVisitHistoryFn: func(_ context.Context, _ string) ([]domain.VisitSummary, error) {
			return nil, repoErr
		},
	}, fakeSquadDirectory{})

	_, err := uc.List(t.Context(), "squad-1", "user-1")
	assert.ErrorIs(t, err, repoErr)
}
