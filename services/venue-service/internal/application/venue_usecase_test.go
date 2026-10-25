package application

import (
	"context"
	"errors"
	"testing"

	"github.com/SamPariat/roundup/services/venue-service/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// venue_usecase_test.go tests VenueUseCase using hand-rolled fakes for the
// PlaceProvider and VenueCache ports. No real network or cache calls are made.

// --- Search tests ---

// validSearch passes SearchParams.Validate so tests exercise the cache and provider paths.
var validSearch = domain.SearchParams{Latitude: 12.97, Longitude: 77.59, RadiusInMeters: 2000}

type fakeVenueCache struct {
	getNearbyFn func(ctx context.Context, key string) (*domain.SearchResult, error)
	getDetailFn func(ctx context.Context, placeID string) (*domain.VenueDetail, error)
	setKeys     []string
}

type fakePlaceProvider struct {
	searchNearbyFn func(ctx context.Context, params domain.SearchParams) (*domain.SearchResult, error)
	getDetailFn    func(ctx context.Context, placeID string) (*domain.VenueDetail, error)
	photoFn        func(ctx context.Context, reference string, maxWidth int) (*domain.Photo, error)
}

func (f *fakeVenueCache) GetNearby(ctx context.Context, key string) (*domain.SearchResult, error) {
	if f.getNearbyFn != nil {
		return f.getNearbyFn(ctx, key)
	}

	return nil, nil
}

func (f *fakeVenueCache) SetNearby(_ context.Context, key string, _ *domain.SearchResult) error {
	f.setKeys = append(f.setKeys, key)

	return nil
}

func (f *fakeVenueCache) GetDetail(ctx context.Context, placeID string) (*domain.VenueDetail, error) {
	if f.getDetailFn != nil {
		return f.getDetailFn(ctx, placeID)
	}

	return nil, nil
}

func (f *fakeVenueCache) SetDetail(_ context.Context, _ string, _ *domain.VenueDetail) error {
	return nil
}

func (f *fakePlaceProvider) SearchNearby(ctx context.Context, params domain.SearchParams) (*domain.SearchResult, error) {
	if f.searchNearbyFn != nil {
		return f.searchNearbyFn(ctx, params)
	}

	return nil, nil
}

func (f *fakePlaceProvider) GetDetail(ctx context.Context, placeID string) (*domain.VenueDetail, error) {
	if f.getDetailFn != nil {
		return f.getDetailFn(ctx, placeID)
	}

	return nil, nil
}

func (f *fakePlaceProvider) Photo(ctx context.Context, reference string, maxWidth int) (*domain.Photo, error) {
	if f.photoFn != nil {
		return f.photoFn(ctx, reference, maxWidth)
	}

	return nil, nil
}

func TestSearchCacheHit(t *testing.T) {
	cached := &domain.SearchResult{Venues: []domain.Venue{{PlaceID: "abc", Name: "The Rooftop"}}, NextPageToken: "tok"}

	uc := NewVenueUseCase(
		// provider should never be called on a cache hit
		&fakePlaceProvider{},
		&fakeVenueCache{
			getNearbyFn: func(_ context.Context, _ string) (*domain.SearchResult, error) {
				return cached, nil
			},
		},
	)

	got, err := uc.Search(t.Context(), validSearch)
	require.NoError(t, err)
	assert.Equal(t, cached, got)
}

func TestSearchCacheMissCallsProvider(t *testing.T) {
	providerVenues := &domain.SearchResult{Venues: []domain.Venue{{PlaceID: "xyz", Name: "Bar Boulud"}}}
	providerCalled := false

	uc := NewVenueUseCase(
		&fakePlaceProvider{
			searchNearbyFn: func(_ context.Context, _ domain.SearchParams) (*domain.SearchResult, error) {
				providerCalled = true

				return providerVenues, nil
			},
		},
		&fakeVenueCache{
			getNearbyFn: func(_ context.Context, _ string) (*domain.SearchResult, error) {
				return nil, nil // cache miss
			},
		},
	)

	got, err := uc.Search(t.Context(), validSearch)
	require.NoError(t, err)
	assert.True(t, providerCalled)
	assert.Equal(t, providerVenues, got)
}

func TestSearchProviderErrorReturnsError(t *testing.T) {
	providerErr := errors.New("provider down")

	uc := NewVenueUseCase(
		&fakePlaceProvider{
			searchNearbyFn: func(_ context.Context, _ domain.SearchParams) (*domain.SearchResult, error) {
				return nil, providerErr
			},
		},
		&fakeVenueCache{
			getNearbyFn: func(_ context.Context, _ string) (*domain.SearchResult, error) {
				return nil, nil
			},
		},
	)

	_, err := uc.Search(t.Context(), validSearch)
	assert.ErrorIs(t, err, providerErr)
}

func TestSearchInvalidParamsSkipsCacheAndProvider(t *testing.T) {
	called := false

	uc := NewVenueUseCase(
		&fakePlaceProvider{
			searchNearbyFn: func(_ context.Context, _ domain.SearchParams) (*domain.SearchResult, error) {
				called = true

				return nil, nil
			},
		},
		&fakeVenueCache{
			getNearbyFn: func(_ context.Context, _ string) (*domain.SearchResult, error) {
				called = true

				return nil, nil
			},
		},
	)

	_, err := uc.Search(t.Context(), domain.SearchParams{})
	require.ErrorIs(t, err, domain.ErrInvalidSearch)
	assert.False(t, called, "cache and provider must not be called for invalid params")
}

func TestSearchCachesEachPageSeparately(t *testing.T) {
	cache := &fakeVenueCache{}
	uc := NewVenueUseCase(&fakePlaceProvider{
		searchNearbyFn: func(_ context.Context, _ domain.SearchParams) (*domain.SearchResult, error) {
			return &domain.SearchResult{}, nil
		},
	}, cache)

	second := validSearch
	second.PageToken = "page-2"

	_, err := uc.Search(t.Context(), validSearch)
	require.NoError(t, err)
	_, err = uc.Search(t.Context(), second)
	require.NoError(t, err)

	require.Len(t, cache.setKeys, 2)
	assert.NotEqual(t, cache.setKeys[0], cache.setKeys[1], "page 2 must not be served from page 1's cache entry")
}

// --- GetDetail tests ---

func TestGetDetailCacheHit(t *testing.T) {
	cached := &domain.VenueDetail{PlaceID: "abc"}

	uc := NewVenueUseCase(
		&fakePlaceProvider{},
		&fakeVenueCache{
			getDetailFn: func(_ context.Context, _ string) (*domain.VenueDetail, error) {
				return cached, nil
			},
		},
	)

	got, err := uc.GetDetail(t.Context(), "abc")
	require.NoError(t, err)
	assert.Equal(t, cached, got)
}

func TestGetDetailCacheMissCallsProvider(t *testing.T) {
	detail := &domain.VenueDetail{PlaceID: "abc"}
	providerCalled := false

	uc := NewVenueUseCase(
		&fakePlaceProvider{
			getDetailFn: func(_ context.Context, _ string) (*domain.VenueDetail, error) {
				providerCalled = true

				return detail, nil
			},
		},
		&fakeVenueCache{
			getDetailFn: func(_ context.Context, _ string) (*domain.VenueDetail, error) {
				return nil, nil
			},
		},
	)

	got, err := uc.GetDetail(t.Context(), "abc")
	require.NoError(t, err)
	assert.True(t, providerCalled)
	assert.Equal(t, detail, got)
}

func TestGetDetailProviderErrorReturnsError(t *testing.T) {
	uc := NewVenueUseCase(
		&fakePlaceProvider{
			getDetailFn: func(_ context.Context, _ string) (*domain.VenueDetail, error) {
				return nil, domain.ErrVenueNotFound
			},
		},
		&fakeVenueCache{
			getDetailFn: func(_ context.Context, _ string) (*domain.VenueDetail, error) {
				return nil, nil
			},
		},
	)

	_, err := uc.GetDetail(t.Context(), "abc")
	assert.ErrorIs(t, err, domain.ErrVenueNotFound)
}

func TestPhotoClampsWidth(t *testing.T) {
	tests := []struct {
		name      string
		requested int
		want      int
	}{
		{"OmittedUsesDefault", 0, domain.DefaultPhotoWidth},
		{"WithinRangeKept", 400, 400},
		{"TooWideCapped", 4000, domain.MaxPhotoWidth},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got int

			uc := NewVenueUseCase(&fakePlaceProvider{
				photoFn: func(_ context.Context, _ string, maxWidth int) (*domain.Photo, error) {
					got = maxWidth

					return &domain.Photo{}, nil
				},
			}, &fakeVenueCache{})

			_, err := uc.Photo(t.Context(), "ref", tt.requested)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
