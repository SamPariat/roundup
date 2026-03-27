package places

import (
	"errors"
	"testing"

	"github.com/SamPariatIL/roundup/services/venue-service/internal/domain"
	"github.com/stretchr/testify/assert"
	"googlemaps.github.io/maps"
)

func TestMapProviderError(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantSentinel error
	}{
		{name: "zero results", input: "ZERO_RESULTS", wantSentinel: domain.ErrVenueNotFound},
		{name: "over query limit", input: "OVER_QUERY_LIMIT", wantSentinel: domain.ErrProviderUnavailable},
		{name: "request denied", input: "REQUEST_DENIED", wantSentinel: domain.ErrProviderUnavailable},
		{name: "unknown api error", input: "UNKNOWN_ERROR", wantSentinel: domain.ErrProviderUnavailable},
		{name: "unrecognised error", input: "connection refused", wantSentinel: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mapProviderError(errors.New(tt.input))

			if tt.wantSentinel != nil {
				assert.ErrorIs(t, err, tt.wantSentinel)
			} else {
				assert.False(t, errors.Is(err, domain.ErrVenueNotFound))
				assert.False(t, errors.Is(err, domain.ErrProviderUnavailable))
			}
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

	assert.True(t, len(venue.PhotoRefs) == len(res.Photos))
	assert.Equal(t, domain.PhotoRef{Reference: "ref1", Width: 1600, Height: 900}, venue.PhotoRefs[0])
	assert.Equal(t, domain.PhotoRef{Reference: "ref2", Width: 800, Height: 600}, venue.PhotoRefs[1])
}

func TestMapPlaceResultToVenueIsOpenSet(t *testing.T) {
	open := true

	res := maps.PlacesSearchResult{
		OpeningHours: &maps.OpeningHours{OpenNow: &open},
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
