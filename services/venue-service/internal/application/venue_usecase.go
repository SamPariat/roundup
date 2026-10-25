package application

import (
	"context"
	"fmt"

	"github.com/SamPariat/roundup/services/venue-service/internal/domain"
)

// searchKeyPrefix namespaces search entries in the cache.
const searchKeyPrefix = "venue:search:"

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

// Search returns venues near the location described by params.
// It checks the cache first; on a miss it calls the provider and warms the cache.
// Cache errors are ignored — the provider is always the fallback source of truth.
// Returns domain.ErrInvalidSearch if params fail validation.
func (u *VenueUseCase) Search(ctx context.Context, params domain.SearchParams) ([]domain.Venue, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}

	key := buildCacheKey(params)

	venues, _ := u.cache.GetNearby(ctx, key)
	if venues != nil {
		return venues, nil
	}

	venues, err := u.provider.SearchNearby(ctx, params)
	if err != nil {
		return nil, err
	}

	_ = u.cache.SetNearby(ctx, key, venues)

	return venues, nil
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

// Photo streams a venue photo. maxWidth outside 1–domain.MaxPhotoWidth falls
// back to domain.DefaultPhotoWidth or the maximum. Photos are not cached here:
// clients and CDNs cache them through the response headers.
func (u *VenueUseCase) Photo(ctx context.Context, reference string, maxWidth int) (*domain.Photo, error) {
	switch {
	case maxWidth <= 0:
		maxWidth = domain.DefaultPhotoWidth
	case maxWidth > domain.MaxPhotoWidth:
		maxWidth = domain.MaxPhotoWidth
	}

	return u.provider.Photo(ctx, reference, maxWidth)
}

// buildCacheKey returns a deterministic cache key for a given set of SearchParams.
// The key encodes all filter fields so that different searches never collide.
func buildCacheKey(params domain.SearchParams) string {
	return fmt.Sprintf(
		searchKeyPrefix+"%f:%f:%d:%s:%s",
		params.Latitude,
		params.Longitude,
		params.RadiusInMeters,
		params.Query,
		params.Type,
	)
}
