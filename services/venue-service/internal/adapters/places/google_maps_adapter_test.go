package places

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/SamPariat/roundup/services/venue-service/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"googlemaps.github.io/maps"
)

func TestMapProviderError(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantSentinel error
	}{
		{name: "NotFound", input: "maps: NOT_FOUND - ", wantSentinel: domain.ErrVenueNotFound},
		{name: "OverQueryLimit", input: "OVER_QUERY_LIMIT", wantSentinel: domain.ErrProviderUnavailable},
		{name: "RequestDenied", input: "REQUEST_DENIED", wantSentinel: domain.ErrProviderUnavailable},
		{name: "UnknownAPIError", input: "UNKNOWN_ERROR", wantSentinel: domain.ErrProviderUnavailable},
		{name: "UnrecognisedError", input: "connection refused", wantSentinel: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mapProviderError(errors.New(tt.input))

			if tt.wantSentinel != nil {
				assert.ErrorIs(t, err, tt.wantSentinel)
			} else {
				require.NotErrorIs(t, err, domain.ErrVenueNotFound)
				assert.NotErrorIs(t, err, domain.ErrProviderUnavailable)
			}
		})
	}
}

func TestMapDetailError(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantSentinel error
	}{
		{name: "MalformedPlaceID", input: "maps: INVALID_REQUEST - ", wantSentinel: domain.ErrVenueNotFound},
		{name: "UnknownPlaceID", input: "maps: NOT_FOUND - ", wantSentinel: domain.ErrVenueNotFound},
		{name: "RequestDenied", input: "maps: REQUEST_DENIED - billing", wantSentinel: domain.ErrProviderUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.ErrorIs(t, mapDetailError(errors.New(tt.input), "bad-id"), tt.wantSentinel)
		})
	}
}

func TestMapPlaceResultToVenueNilOpeningHours(t *testing.T) {
	res := maps.PlacesSearchResult{
		PlaceID: "abc123",
		Name:    "The Rooftop",
		// OpeningHours intentionally nil — provider did not return it
	}

	venue := mapPlaceResultToVenue(res)

	assert.Nil(t, venue.IsOpen, "IsOpen should be nil when provider omits opening hours")
}

func TestMapPlaceResultToVenuePhotoRefDimensions(t *testing.T) {
	res := maps.PlacesSearchResult{
		Photos: []maps.Photo{
			{PhotoReference: "ref1", Width: 1600, Height: 900},
			{PhotoReference: "ref2", Width: 800, Height: 600},
		},
	}

	venue := mapPlaceResultToVenue(res)

	assert.Len(t, res.Photos, len(venue.PhotoRefs))
	assert.Equal(t, domain.PhotoRef{Reference: "ref1", Width: 1600, Height: 900}, venue.PhotoRefs[0])
	assert.Equal(t, domain.PhotoRef{Reference: "ref2", Width: 800, Height: 600}, venue.PhotoRefs[1])
}

func TestMapPlaceResultToVenueIsOpenSet(t *testing.T) {
	res := maps.PlacesSearchResult{
		OpeningHours: &maps.OpeningHours{OpenNow: new(true)},
	}

	venue := mapPlaceResultToVenue(res)

	assert.NotNil(t, venue.IsOpen)
	assert.True(t, *venue.IsOpen)
}

func TestMapPlaceDetailsToVenueDetailNilOpeningHours(t *testing.T) {
	res := maps.PlaceDetailsResult{PlaceID: "abc123"}

	detail := mapPlaceDetailsToVenueDetail(res)

	assert.Nil(t, detail.OpeningHours)
}

func TestMapPlaceResultToVenueRatingCount(t *testing.T) {
	venue := mapPlaceResultToVenue(maps.PlacesSearchResult{Rating: 4.5, UserRatingsTotal: 1240})

	assert.Equal(t, 1240, venue.RatingCount)
}

func TestMapPlaceDetailsCarriesRatingCount(t *testing.T) {
	detail := mapPlaceDetailsToVenueDetail(maps.PlaceDetailsResult{UserRatingsTotal: 87})

	assert.Equal(t, 87, detail.RatingCount)
}

func TestMapPhotoError(t *testing.T) {
	require.ErrorIs(t, mapPhotoError(errors.New("maps: status code 400"), "ref"), domain.ErrPhotoNotFound)
	assert.ErrorIs(t, mapPhotoError(errors.New("maps: OVER_QUERY_LIMIT"), "ref"), domain.ErrProviderUnavailable)
}

// fakeGoogle serves Nearby Search with one canned body per call, in order.
func fakeGoogle(t *testing.T, bodies ...string) (*GoogleMapsAdapter, *int) {
	t.Helper()

	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/maps/api/place/nearbysearch/json", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(bodies[min(calls, len(bodies)-1)]))
		calls++
	}))
	t.Cleanup(srv.Close)

	client, err := maps.NewClient(maps.WithAPIKey("test-key"), maps.WithBaseURL(srv.URL))
	require.NoError(t, err)

	return &GoogleMapsAdapter{client: client, pageTokenWait: time.Millisecond}, &calls
}

const (
	invalidRequest = `{"status":"INVALID_REQUEST","results":[]}`
	onePage        = `{"status":"OK","next_page_token":"tok-3","results":[
		{"place_id":"p1","name":"Toit","geometry":{"location":{"lat":12.97,"lng":77.59}}}]}`
)

func TestSearchNearbyReturnsNextPageToken(t *testing.T) {
	adapter, _ := fakeGoogle(t, onePage)

	got, err := adapter.SearchNearby(t.Context(), domain.SearchParams{Latitude: 12.97, Longitude: 77.59, RadiusInMeters: 500})

	require.NoError(t, err)
	require.Len(t, got.Venues, 1)
	assert.Equal(t, "p1", got.Venues[0].PlaceID)
	assert.Equal(t, "tok-3", got.NextPageToken)
}

func TestSearchNearbyRetriesATokenNotYetActive(t *testing.T) {
	adapter, calls := fakeGoogle(t, invalidRequest, onePage)

	got, err := adapter.SearchNearby(t.Context(), domain.SearchParams{PageToken: "tok-2"})

	require.NoError(t, err)
	assert.Equal(t, 2, *calls, "one retry after the token wait")
	assert.Equal(t, "tok-3", got.NextPageToken)
}

func TestSearchNearbyRejectsABadToken(t *testing.T) {
	adapter, calls := fakeGoogle(t, invalidRequest)

	_, err := adapter.SearchNearby(t.Context(), domain.SearchParams{PageToken: "made-up"})

	require.ErrorIs(t, err, domain.ErrInvalidSearch, "a cursor Google never issued is the client's mistake: 400")
	assert.Equal(t, 2, *calls)
}

func TestSearchNearbyDoesNotRetryWithoutToken(t *testing.T) {
	adapter, calls := fakeGoogle(t, invalidRequest)

	_, err := adapter.SearchNearby(t.Context(), domain.SearchParams{Latitude: 12.97, Longitude: 77.59, RadiusInMeters: 500})

	require.Error(t, err)
	require.NotErrorIs(t, err, domain.ErrInvalidSearch)
	assert.Equal(t, 1, *calls, "INVALID_REQUEST on a first page is not a token problem")
}
