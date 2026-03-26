# adapters/cache

Implements `domain.VenueCache` — the port that abstracts all caching operations for venue data.

## Files

| File | Purpose |
|---|---|
| `redis_cache_adapter.go` | Production adapter. Uses `go-redis/v9` to store and retrieve serialised venue data in Redis. |

## TTL Policy

| Cache key pattern | TTL | Rationale |
|---|---|---|
| `venues:nearby:{lat}:{lng}:{radius}:{type}:{query}` | 24 hours | Nearby search results change slowly; long TTL reduces Places API costs |
| `venue:detail:{placeId}` | 1 hour | Place details (hours, phone, reviews) change more frequently |

TTL values are constants defined inside `redis_cache_adapter.go` — they are an adapter implementation detail, not a domain concern.

## Serialisation

Domain structs are JSON-marshalled before writing to Redis and unmarshalled on read. The adapter handles all encoding/decoding internally; callers receive and pass domain types only.

## Cache miss behaviour

`GetNearby` and `GetDetail` return `found=false` (not an error) on a cache miss. The use-case layer is responsible for deciding what to do on a miss (typically: fetch from the provider, then write back to cache).

## Best-effort writes

Cache write failures are non-fatal — a failed `SetNearby` or `SetDetail` must not cause the API request to fail. The use-case layer discards cache write errors with `_ = cache.Set(...)`. The adapter should still log the failure internally for observability.
