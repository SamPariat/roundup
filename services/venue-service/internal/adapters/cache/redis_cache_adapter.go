package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/SamPariatIL/roundup/services/venue-service/internal/domain"
	"github.com/redis/go-redis/v9"
)

// redis_cache_adapter.go implements domain.VenueCache using Redis.
// All methods serialize domain types to JSON before writing and deserialize on read.
// Cache misses (redis.Nil) are returned as nil, nil — not as errors.

// RedisCacheAdapter implements domain.VenueCache backed by a Redis instance.
type RedisCacheAdapter struct {
	client *redis.Client
	ttl    time.Duration
}

// NewRedisCacheAdapter constructs a RedisCacheAdapter with the given client and TTL.
// The TTL is applied uniformly to all cache writes.
func NewRedisCacheAdapter(client *redis.Client, ttl time.Duration) *RedisCacheAdapter {
	return &RedisCacheAdapter{client: client, ttl: ttl}
}

// GetNearby returns cached nearby venues for the given key.
// Returns nil, nil on a cache miss — the caller should fall through to the provider.
func (r *RedisCacheAdapter) GetNearby(ctx context.Context, key string) ([]domain.Venue, error) {
	val, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var venues []domain.Venue
	if err := json.Unmarshal([]byte(val), &venues); err != nil {
		return nil, err
	}

	return venues, nil
}

// SetNearby serialises venues to JSON and stores them under the key with the configured TTL.
func (r *RedisCacheAdapter) SetNearby(ctx context.Context, key string, venues []domain.Venue) error {
	val, err := json.Marshal(venues)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, val, r.ttl).Err()
}

// GetDetail returns the cached VenueDetail for the given placeID.
// Returns nil, nil on a cache miss.
func (r *RedisCacheAdapter) GetDetail(ctx context.Context, placeID string) (*domain.VenueDetail, error) {
	val, err := r.client.Get(ctx, placeID).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var detail domain.VenueDetail
	if err := json.Unmarshal([]byte(val), &detail); err != nil {
		return nil, err
	}

	return &detail, nil
}

// SetDetail serialises detail to JSON and stores it under placeID with the configured TTL.
func (r *RedisCacheAdapter) SetDetail(ctx context.Context, placeID string, detail *domain.VenueDetail) error {
	val, err := json.Marshal(detail)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, placeID, val, r.ttl).Err()
}

// InvalidateDetail deletes the cached VenueDetail for the given placeID.
func (r *RedisCacheAdapter) InvalidateDetail(ctx context.Context, placeID string) error {
	return r.client.Del(ctx, placeID).Err()
}
