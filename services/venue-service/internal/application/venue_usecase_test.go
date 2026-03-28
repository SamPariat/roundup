// venue_usecase_test.go tests VenueUseCase using hand-rolled fakes for the
// PlaceProvider and VenueCache ports. No real network or cache calls are made.
package application

import (
	"context"
	"errors"
	"testing"

	"github.com/SamPariatIL/roundup/services/venue-service/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Search tests ---

type fakeCache struct {
	getNearbyFn func(ctx context.Context, key string) ([]domain.Venue, error)
	getDetailFn func(ctx context.Context, placeID string) (*domain.VenueDetail, error)
}

type fakeProvider struct {
	searchNearbyFn func(ctx context.Context, params domain.SearchParams) ([]domain.Venue, error)
	getDetailFn    func(ctx context.Context, placeID string) (*domain.VenueDetail, error)
}

func (f *fakeCache) GetNearby(ctx context.Context, key string) ([]domain.Venue, error) {
	if f.getNearbyFn != nil {
		return f.getNearbyFn(ctx, key)
	}

	return nil, nil
}

func (f *fakeCache) SetNearby(_ context.Context, _ string, _ []domain.Venue) error {
	return nil
}

func (f *fakeCache) GetDetail(ctx context.Context, placeID string) (*domain.VenueDetail, error) {
	if f.getDetailFn != nil {
		return f.getDetailFn(ctx, placeID)
	}

	return nil, nil
}

func (f *fakeCache) SetDetail(_ context.Context, _ string, _ *domain.VenueDetail) error {
	return nil
}

func (f *fakeCache) InvalidateDetail(_ context.Context, _ string) error {
	return nil
}

func (f *fakeProvider) SearchNearby(ctx context.Context, params domain.SearchParams) ([]domain.Venue, error) {
	if f.searchNearbyFn != nil {
		return f.searchNearbyFn(ctx, params)
	}
	return nil, nil
}

func (f *fakeProvider) GetDetail(ctx context.Context, placeID string) (*domain.VenueDetail, error) {
	if f.getDetailFn != nil {
		return f.getDetailFn(ctx, placeID)
	}
	return nil, nil
}

func TestSearchCacheHit(t *testing.T) {
	cached := []domain.Venue{{PlaceID: "abc", Name: "The Rooftop"}}

	uc := NewVenueUseCase(
		// provider should never be called on a cache hit
		&fakeProvider{},
		&fakeCache{
			getNearbyFn: func(_ context.Context, _ string) ([]domain.Venue, error) {
				return cached, nil
			},
		},
	)

	got, err := uc.Search(context.Background(), domain.SearchParams{})
	require.NoError(t, err)
	assert.Equal(t, cached, got)
}

func TestSearchCacheMissCallsProvider(t *testing.T) {
	providerVenues := []domain.Venue{{PlaceID: "xyz", Name: "Bar Boulud"}}
	providerCalled := false

	uc := NewVenueUseCase(
		&fakeProvider{
			searchNearbyFn: func(_ context.Context, _ domain.SearchParams) ([]domain.Venue, error) {
				providerCalled = true
				return providerVenues, nil
			},
		},
		&fakeCache{
			getNearbyFn: func(_ context.Context, _ string) ([]domain.Venue, error) {
				return nil, nil // cache miss
			},
		},
	)

	got, err := uc.Search(context.Background(), domain.SearchParams{})
	require.NoError(t, err)
	assert.True(t, providerCalled)
	assert.Equal(t, providerVenues, got)
}

func TestSearchProviderErrorReturnsError(t *testing.T) {
	providerErr := errors.New("provider down")

	uc := NewVenueUseCase(
		&fakeProvider{
			searchNearbyFn: func(_ context.Context, _ domain.SearchParams) ([]domain.Venue, error) {
				return nil, providerErr
			},
		},
		&fakeCache{
			getNearbyFn: func(_ context.Context, _ string) ([]domain.Venue, error) {
				return nil, nil
			},
		},
	)

	_, err := uc.Search(context.Background(), domain.SearchParams{})
	assert.ErrorIs(t, err, providerErr)
}

// --- GetDetail tests ---

func TestGetDetailCacheHit(t *testing.T) {
	cached := &domain.VenueDetail{Venue: domain.Venue{PlaceID: "abc"}}

	uc := NewVenueUseCase(
		&fakeProvider{},
		&fakeCache{
			getDetailFn: func(_ context.Context, _ string) (*domain.VenueDetail, error) {
				return cached, nil
			},
		},
	)

	got, err := uc.GetDetail(context.Background(), "abc")
	require.NoError(t, err)
	assert.Equal(t, cached, got)
}

func TestGetDetailCacheMissCallsProvider(t *testing.T) {
	detail := &domain.VenueDetail{Venue: domain.Venue{PlaceID: "abc"}}
	providerCalled := false

	uc := NewVenueUseCase(
		&fakeProvider{
			getDetailFn: func(_ context.Context, _ string) (*domain.VenueDetail, error) {
				providerCalled = true
				return detail, nil
			},
		},
		&fakeCache{
			getDetailFn: func(_ context.Context, _ string) (*domain.VenueDetail, error) {
				return nil, nil
			},
		},
	)

	got, err := uc.GetDetail(context.Background(), "abc")
	require.NoError(t, err)
	assert.True(t, providerCalled)
	assert.Equal(t, detail, got)
}

func TestGetDetailProviderErrorReturnsError(t *testing.T) {
	uc := NewVenueUseCase(
		&fakeProvider{
			getDetailFn: func(_ context.Context, _ string) (*domain.VenueDetail, error) {
				return nil, domain.ErrVenueNotFound
			},
		},
		&fakeCache{
			getDetailFn: func(_ context.Context, _ string) (*domain.VenueDetail, error) {
				return nil, nil
			},
		},
	)

	_, err := uc.GetDetail(context.Background(), "abc")
	assert.ErrorIs(t, err, domain.ErrVenueNotFound)
}
