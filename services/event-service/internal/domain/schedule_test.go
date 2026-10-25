package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SamPariat/roundup/services/event-service/internal/domain"
)

func TestNewSchedule(t *testing.T) {
	now := time.Date(2026, 10, 25, 18, 0, 0, 0, time.UTC)
	at := func(d time.Duration) time.Time { return now.Add(d) }
	ptr := func(d time.Duration) *time.Time {
		v := at(d)

		return &v
	}

	tests := []struct {
		name         string
		startsAt     time.Time
		deadline     *time.Time
		wantDeadline time.Time
		wantErr      bool
	}{
		{name: "default is the start", startsAt: at(72 * time.Hour), wantDeadline: at(72 * time.Hour)},
		{name: "picked deadline is kept", startsAt: at(72 * time.Hour), deadline: ptr(time.Hour), wantDeadline: at(time.Hour)},
		{name: "deadline at the start", startsAt: at(72 * time.Hour), deadline: ptr(72 * time.Hour), wantDeadline: at(72 * time.Hour)},
		{name: "start in the past", startsAt: at(-time.Hour), wantErr: true},
		{name: "start now", startsAt: now, wantErr: true},
		{name: "deadline in the past", startsAt: at(72 * time.Hour), deadline: ptr(-time.Minute), wantErr: true},
		{name: "deadline after the start", startsAt: at(72 * time.Hour), deadline: ptr(73 * time.Hour), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.NewSchedule(now, tt.startsAt, tt.deadline)
			if tt.wantErr {
				require.ErrorIs(t, err, domain.ErrInvalidEvent)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.startsAt, got.StartsAt)
			assert.Equal(t, tt.wantDeadline, got.VotingDeadline)
		})
	}
}
