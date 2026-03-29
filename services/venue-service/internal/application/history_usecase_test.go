package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SamPariatIL/roundup/services/venue-service/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// history_usecase_test.go tests HistoryUseCase using the fakeRepo defined in
// favorite_usecase_test.go. No real database calls are made.

func TestRecordVisitSuccess(t *testing.T) {
	uc := NewHistoryUseCase(&fakeRepo{})

	err := uc.RecordVisit(context.Background(), domain.RecordVisitCommand{
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

	uc := NewHistoryUseCase(&fakeRepo{
		recordVisitFn: func(_ context.Context, _ domain.RecordVisitCommand) error {
			return repoErr
		},
	})

	err := uc.RecordVisit(context.Background(), domain.RecordVisitCommand{})
	assert.ErrorIs(t, err, repoErr)
}

func TestGetVisitHistoryReturnsSummaries(t *testing.T) {
	summaries := []domain.VisitSummary{
		{PlaceID: "abc", Name: "The Rooftop", VisitCount: 3, LastVisitedAt: time.Now()},
	}

	uc := NewHistoryUseCase(&fakeRepo{
		getVisitHistoryFn: func(_ context.Context, _ string) ([]domain.VisitSummary, error) {
			return summaries, nil
		},
	})

	got, err := uc.GetVisitHistory(context.Background(), "squad-1")
	require.NoError(t, err)
	assert.Equal(t, summaries, got)
}

func TestGetVisitHistoryPropagatesError(t *testing.T) {
	repoErr := errors.New("db read failed")

	uc := NewHistoryUseCase(&fakeRepo{
		getVisitHistoryFn: func(_ context.Context, _ string) ([]domain.VisitSummary, error) {
			return nil, repoErr
		},
	})

	_, err := uc.GetVisitHistory(context.Background(), "squad-1")
	assert.ErrorIs(t, err, repoErr)
}
