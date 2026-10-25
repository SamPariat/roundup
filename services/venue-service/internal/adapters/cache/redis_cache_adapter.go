package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/SamPariat/roundup/services/venue-service/internal/domain"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// DefaultTTL is how long cached search results and details stay valid.
const DefaultTTL = 10 * time.Minute

// detailKeyPrefix namespaces detail entries, so a place ID can never collide
// with another key in the same Redis.
const detailKeyPrefix = "venue:detail:"

// RedisCacheAdapter implements domain.VenueCache by storing domain types as JSON
// in Redis. Cache misses (redis.Nil) are returned as (nil, nil), not as errors.
// Every failure is logged at warn level before it is returned.
type RedisCacheAdapter struct {
	client *redis.Client
	ttl    time.Duration
	log    *zap.Logger
}

// NewRedisCacheAdapter constructs a RedisCacheAdapter with the given client and TTL.
// The TTL is applied uniformly to all cache writes.
func NewRedisCacheAdapter(client *redis.Client, ttl time.Duration, log *zap.Logger) *RedisCacheAdapter {
	return &RedisCacheAdapter{client: client, ttl: ttl, log: log}
}

// GetNearby returns the cached search page for the given key.
// Returns nil, nil on a cache miss — the caller should fall through to the provider.
func (r *RedisCacheAdapter) GetNearby(ctx context.Context, key string) (*domain.SearchResult, error) {
	val, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}

	if err != nil {
		r.log.Warn("failed to get nearby venues", zap.String("key", key), zap.Error(err))

		return nil, err
	}

	var result domain.SearchResult
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		r.log.Warn("failed to unmarshal nearby venues", zap.String("key", key), zap.Error(err))

		return nil, err
	}

	return &result, nil
}

// SetNearby serialises a search page to JSON and stores it under the key with the configured TTL.
func (r *RedisCacheAdapter) SetNearby(ctx context.Context, key string, result *domain.SearchResult) error {
	val, err := json.Marshal(result)
	if err != nil {
		r.log.Warn("failed to marshal nearby venues", zap.String("key", key), zap.Error(err))

		return err
	}

	if err := r.client.Set(ctx, key, val, r.ttl).Err(); err != nil {
		r.log.Warn("failed to set nearby venues", zap.String("key", key), zap.Error(err))

		return err
	}

	return nil
}

// GetDetail returns the cached VenueDetail for the given placeID.
// Returns nil, nil on a cache miss.
func (r *RedisCacheAdapter) GetDetail(ctx context.Context, placeID string) (*domain.VenueDetail, error) {
	val, err := r.client.Get(ctx, detailKeyPrefix+placeID).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}

	if err != nil {
		r.log.Warn("failed to get venue detail", zap.String("placeID", placeID), zap.Error(err))

		return nil, err
	}

	var detail domain.VenueDetail
	if err := json.Unmarshal([]byte(val), &detail); err != nil {
		r.log.Warn("failed to unmarshal venue detail", zap.String("placeID", placeID), zap.Error(err))

		return nil, err
	}

	return &detail, nil
}

// SetDetail serialises detail to JSON and stores it under placeID with the configured TTL.
func (r *RedisCacheAdapter) SetDetail(ctx context.Context, placeID string, detail *domain.VenueDetail) error {
	val, err := json.Marshal(detail)
	if err != nil {
		r.log.Warn("failed to marshal venue detail", zap.String("placeID", placeID), zap.Error(err))

		return err
	}

	if err := r.client.Set(ctx, detailKeyPrefix+placeID, val, r.ttl).Err(); err != nil {
		r.log.Warn("failed to set venue detail", zap.String("placeID", placeID), zap.Error(err))

		return err
	}

	return nil
}
