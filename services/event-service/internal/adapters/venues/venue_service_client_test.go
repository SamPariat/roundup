package venues

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/SamPariat/roundup/services/event-service/internal/domain"
	sharedmw "github.com/SamPariat/roundup/shared/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const caller = "7dc92d84-1950-43d6-b676-debacf84d830"

func TestPlaceKeepsTheSnapshotFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/venues/ChIJ%2Fabc", r.URL.EscapedPath())
		assert.Equal(t, caller, r.Header.Get(sharedmw.UserIDHeader))

		_, _ = w.Write([]byte(`{"success":true,"data":{"placeID":"ChIJ/abc","name":"The Lantern","address":"1 Main St",
			"latitude":19.07,"longitude":72.87,"averageRating":4.5,"priceLevel":2,
			"photos":[{"url":"/api/v1/venues/photos/r1","width":400,"height":300},{"url":"/api/v1/venues/photos/r2"}]}}`))
	}))
	defer server.Close()

	place, err := NewVenueServiceClient(server.URL).Place(t.Context(), "ChIJ/abc", caller)
	require.NoError(t, err)
	assert.Equal(t, domain.Place{
		PlaceID: "ChIJ/abc", Name: "The Lantern", Address: "1 Main St", Lat: 19.07, Lng: 72.87, Rating: 4.5,
		PriceLevel: 2, PhotoURL: "/api/v1/venues/photos/r1", PhotoWidth: 400, PhotoHeight: 300,
	}, place)
}

func TestPlaceErrors(t *testing.T) {
	tests := []struct {
		status int
		body   string
		want   error
	}{
		{http.StatusNotFound, `{"success":false,"error":"venue not found"}`, domain.ErrVenueNotFound},
		{http.StatusServiceUnavailable, `{"success":false,"error":"place provider unavailable"}`, domain.ErrVenueServiceUnavailable},
		{http.StatusOK, `not json`, domain.ErrVenueServiceUnavailable},
	}

	for _, tt := range tests {
		t.Run(http.StatusText(tt.status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			_, err := NewVenueServiceClient(server.URL).Place(t.Context(), "p1", caller)
			assert.ErrorIs(t, err, tt.want)
		})
	}

	_, err := NewVenueServiceClient("http://127.0.0.1:1").Place(t.Context(), "p1", caller)
	assert.ErrorIs(t, err, domain.ErrVenueServiceUnavailable, "unreachable")
}

func TestRecordVisit(t *testing.T) {
	startsAt := time.Date(2026, 10, 30, 20, 0, 0, 0, time.UTC)
	event := domain.Event{ID: "e1", SquadID: "s1", StartsAt: startsAt, Winner: &domain.Nomination{PlaceID: "p1"}}

	for _, status := range []int{http.StatusCreated, http.StatusConflict} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/api/v1/squads/s1/history", r.URL.Path)
			assert.Equal(t, caller, r.Header.Get(sharedmw.UserIDHeader))

			var body recordVisitRequest
			assert.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			assert.Equal(t, recordVisitRequest{EventID: "e1", PlaceID: "p1", VisitedAt: startsAt}, body)

			w.WriteHeader(status)
		}))

		require.NoError(t, NewVenueServiceClient(server.URL).RecordVisit(t.Context(), event, caller), "%d counts as recorded", status)
		server.Close()
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	assert.ErrorIs(t, NewVenueServiceClient(server.URL).RecordVisit(t.Context(), event, caller), domain.ErrVenueServiceUnavailable)
}
