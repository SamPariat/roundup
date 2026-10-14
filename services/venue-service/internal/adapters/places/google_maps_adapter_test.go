package places

import (
	"errors"
	"testing"

	"github.com/SamPariat/roundup/services/venue-service/internal/domain"
	"github.com/stretchr/testify/assert"
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
				assert.False(t, errors.Is(err, domain.ErrVenueNotFound))
				assert.False(t, errors.Is(err, domain.ErrProviderUnavailable))
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

	assert.True(t, len(venue.PhotoRefs) == len(res.Photos))
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
