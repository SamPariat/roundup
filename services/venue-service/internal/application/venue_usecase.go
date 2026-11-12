package application

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"sync"

	"github.com/SamPariat/roundup/services/venue-service/internal/domain"
)

// searchKeyPrefix namespaces search entries in the cache. Each entry holds a
// whole page (venues and next token); bump the version when that shape
// changes, so old entries are never read as the new one.
const searchKeyPrefix = "venue:search:v1:"

// maxPhotoLookups caps how many photo URL lookups one call to PhotoURLs runs at
// once; a search page needs up to 20.
const maxPhotoLookups = 10

// VenueLookup returns a venue's details. FavoriteUseCase and HistoryUseCase
// use it to snapshot a venue; VenueUseCase implements it, so lookups go
// through the cache before the provider.
type VenueLookup interface {
	GetDetail(ctx context.Context, placeID string) (*domain.VenueDetail, error)
}

// VenueUseCase orchestrates venue search and detail lookups using a cache-first strategy.
type VenueUseCase struct {
	provider domain.PlaceProvider
	cache    domain.VenueCache
}

// NewVenueUseCase constructs a VenueUseCase with the given provider and cache.
func NewVenueUseCase(provider domain.PlaceProvider, cache domain.VenueCache) *VenueUseCase {
	return &VenueUseCase{provider: provider, cache: cache}
}

// Search returns one page of venues near the location described by params.
// It checks the cache first; on a miss it calls the provider and warms the cache.
// Cache errors are ignored — the provider is always the fallback source of truth.
// Returns domain.ErrInvalidSearch if params fail validation.
func (u *VenueUseCase) Search(ctx context.Context, params domain.SearchParams) (*domain.SearchResult, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}

	key := buildCacheKey(params)

	result, _ := u.cache.GetNearby(ctx, key)
	if result != nil {
		return result, nil
	}

	result, err := u.provider.SearchNearby(ctx, params)
	if err != nil {
		return nil, err
	}

	_ = u.cache.SetNearby(ctx, key, result)

	return result, nil
}

// GetDetail returns full venue information for the given placeID.
// It checks the cache first; on a miss it calls the provider and warms the cache.
// Returns nil and domain.ErrVenueNotFound if the place does not exist.
func (u *VenueUseCase) GetDetail(ctx context.Context, placeID string) (*domain.VenueDetail, error) {
	detail, _ := u.cache.GetDetail(ctx, placeID)
	if detail != nil {
		return detail, nil
	}

	detail, err := u.provider.GetDetail(ctx, placeID)
	if err != nil {
		return nil, err
	}

	_ = u.cache.SetDetail(ctx, placeID, detail)

	return detail, nil
}

// PhotoURLs returns a direct URL, at most maxWidth pixels wide, for each of
// refs, by reference. URLs cached for maxWidth are reused; the rest are looked
// up from the provider and cached. Refs with an empty Reference are skipped,
// and a photo missing from the map is one the response leaves out. Cache
// errors are ignored, like the other cache reads and writes here.
func (u *VenueUseCase) PhotoURLs(ctx context.Context, refs []domain.PhotoRef, maxWidth int) (map[string]string, error) {
	wanted := make([]string, 0, len(refs))
	for _, ref := range refs {
		if ref.Reference != "" && !slices.Contains(wanted, ref.Reference) {
			wanted = append(wanted, ref.Reference)
		}
	}

	cached, _ := u.cache.GetPhotoURLs(ctx, wanted, maxWidth)

	missing := make([]string, 0, len(wanted))
	for _, ref := range wanted {
		if _, ok := cached[ref]; !ok {
			missing = append(missing, ref)
		}
	}

	urls, err := u.lookUpPhotoURLs(ctx, missing, maxWidth)
	if err != nil {
		return nil, err
	}

	_ = u.cache.SetPhotoURLs(ctx, urls, maxWidth)

	maps.Copy(urls, cached)

	return urls, nil
}

// lookUpPhotoURLs asks the provider for each of refs at once (at most
// maxPhotoLookups in flight) and returns the URLs by reference. The first
// domain.ErrProviderUnavailable cancels the other lookups and is returned.
func (u *VenueUseCase) lookUpPhotoURLs(ctx context.Context, refs []string, maxWidth int) (map[string]string, error) {
	var (
		mu      sync.Mutex
		wg      sync.WaitGroup
		failure error
	)

	urls := make(map[string]string, len(refs))
	slots := make(chan struct{}, maxPhotoLookups)

	// Cancelled by the first failure, so the lookups still running stop and
	// the ones waiting for a slot never start: the response fails either way.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// fail makes lookUpPhotoURLs return err instead of the URLs and cancels the
	// remaining lookups; the first call wins.
	fail := func(err error) {
		mu.Lock()
		defer mu.Unlock()

		if failure == nil {
			failure = err

			cancel()
		}
	}

	for _, ref := range refs {
		wg.Go(func() {
			// Wait for a slot, unless a failure already cancelled the rest.
			select {
			case slots <- struct{}{}:
			case <-ctx.Done():
				return
			}

			defer func() { <-slots }()

			url, err := u.provider.PhotoURL(ctx, ref, maxWidth)
			if err != nil {
				// A photo Google no longer has is left out; Google itself
				// failing fails the response, so an outage or a bad key is
				// seen instead of hiding behind cards without photos.
				if errors.Is(err, domain.ErrProviderUnavailable) {
					fail(err)
				}

				return
			}

			mu.Lock()
			urls[ref] = url
			mu.Unlock()
		})
	}

	wg.Wait()

	if failure != nil {
		return nil, failure
	}

	return urls, nil
}

// buildCacheKey returns a deterministic cache key for a given set of SearchParams.
// The key encodes all filter fields so that different searches never collide.
func buildCacheKey(params domain.SearchParams) string {
	return fmt.Sprintf(
		searchKeyPrefix+"%f:%f:%d:%s:%s:%s",
		params.Latitude,
		params.Longitude,
		params.RadiusInMeters,
		params.Query,
		params.Type,
		params.PageToken,
	)
}
