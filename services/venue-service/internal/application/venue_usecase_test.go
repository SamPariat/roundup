package application

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/SamPariat/roundup/services/venue-service/internal/domain"
	"github.com/SamPariat/roundup/shared/domainerr"
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
	photoURLFn     func(ctx context.Context, reference string, maxWidth int) (string, error)
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

func (f *fakePlaceProvider) PhotoURL(ctx context.Context, reference string, maxWidth int) (string, error) {
	if f.photoURLFn != nil {
		return f.photoURLFn(ctx, reference, maxWidth)
	}

	return "", nil
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

func TestPhotoURLsLooksUpEveryPhotoAtTheWidth(t *testing.T) {
	var (
		mu    sync.Mutex
		asked []string
	)

	uc := NewVenueUseCase(&fakePlaceProvider{
		photoURLFn: func(_ context.Context, reference string, maxWidth int) (string, error) {
			mu.Lock()
			defer mu.Unlock()

			asked = append(asked, reference)

			return fmt.Sprintf("https://img.example/%s?w=%d", reference, maxWidth), nil
		},
	}, &fakeVenueCache{})

	urls, err := uc.PhotoURLs(t.Context(), []domain.PhotoRef{{Reference: "a"}, {}, {Reference: "b"}}, 400)
	require.NoError(t, err)

	assert.Equal(t, map[string]string{
		"a": "https://img.example/a?w=400",
		"b": "https://img.example/b?w=400",
	}, urls)
	assert.ElementsMatch(t, []string{"a", "b"}, asked, "a ref without a reference is never looked up")
}

func TestPhotoURLsLeavesOutAPhotoGoogleNoLongerHas(t *testing.T) {
	uc := NewVenueUseCase(&fakePlaceProvider{
		photoURLFn: func(_ context.Context, reference string, _ int) (string, error) {
			if reference == "gone" {
				return "", domainerr.New(domain.ErrPhotoNotFound, reference)
			}

			return "https://img.example/" + reference, nil
		},
	}, &fakeVenueCache{})

	urls, err := uc.PhotoURLs(t.Context(), []domain.PhotoRef{{Reference: "a"}, {Reference: "gone"}}, 400)
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"a": "https://img.example/a"}, urls)
}

func TestPhotoURLsFailsWhenGoogleIsDown(t *testing.T) {
	uc := NewVenueUseCase(&fakePlaceProvider{
		photoURLFn: func(_ context.Context, reference string, _ int) (string, error) {
			if reference == "b" {
				return "", domainerr.New(domain.ErrProviderUnavailable, "quota")
			}

			return "https://img.example/" + reference, nil
		},
	}, &fakeVenueCache{})

	urls, err := uc.PhotoURLs(t.Context(), []domain.PhotoRef{{Reference: "a"}, {Reference: "b"}}, 400)
	require.ErrorIs(t, err, domain.ErrProviderUnavailable)
	assert.Nil(t, urls)
}

func TestPhotoURLsCancelsTheOtherLookupsWhenGoogleIsDown(t *testing.T) {
	var (
		mu        sync.Mutex
		started   int
		cancelled int
	)

	uc := NewVenueUseCase(&fakePlaceProvider{
		photoURLFn: func(ctx context.Context, reference string, _ int) (string, error) {
			mu.Lock()
			started++
			mu.Unlock()

			if reference == "down" {
				return "", domainerr.New(domain.ErrProviderUnavailable, "quota")
			}

			// Every other lookup hangs, like a slow Google, until cancelled.
			select {
			case <-ctx.Done():
				mu.Lock()
				cancelled++
				mu.Unlock()

				return "", domainerr.New(domain.ErrProviderUnavailable, ctx.Err().Error())
			case <-time.After(5 * time.Second):
				return "https://img.example/" + reference, nil
			}
		},
	}, &fakeVenueCache{})

	// More refs than slots, so some wait for a slot when the failure lands.
	refs := []domain.PhotoRef{{Reference: "down"}}
	for i := range 3 * maxPhotoLookups {
		refs = append(refs, domain.PhotoRef{Reference: fmt.Sprintf("slow-%d", i)})
	}

	begin := time.Now()
	_, err := uc.PhotoURLs(t.Context(), refs, 400)

	require.ErrorIs(t, err, domain.ErrProviderUnavailable)
	assert.Less(t, time.Since(begin), time.Second, "the hanging lookups are cancelled, not waited out")
	assert.Equal(t, started-1, cancelled, "every lookup that started, other than the failed one, saw the cancel")
	assert.Less(t, started, len(refs), "lookups still waiting for a slot never start")
}
