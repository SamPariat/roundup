# adapters/places

Implements `domain.PlaceProvider` — the port that abstracts all venue search and detail lookups from a third-party places data source.

## Files

| File | Purpose |
|---|---|
| `google_maps_adapter.go` | Production adapter. Wraps the Google Maps Go SDK (`googlemaps/google-maps-services-go`), calls the Places API (Nearby Search + Place Details), and maps responses to `domain.Venue` / `domain.VenueDetail`. |

## Adding a new provider

Create a new file in this package (e.g. `foursquare_adapter.go`) and implement the two methods of `domain.PlaceProvider`:

```go
SearchNearby(ctx context.Context, params domain.SearchParams) ([]domain.Venue, error)
GetDetail(ctx context.Context, placeID string) (domain.VenueDetail, error)
```

Then wire it in `cmd/main.go` instead of (or alongside) `GoogleMapsAdapter`. No changes required anywhere else.

## Mapping responsibility

The adapter is responsible for:
- Translating `domain.SearchParams` fields (lat, lng, radius, type, keyword) into the provider's specific request format
- Normalising provider-specific fields to domain types — e.g. Google's `price_level` (0–4) maps directly to `domain.Venue.PriceLevel`
- Returning `domain.ErrVenueNotFound` when the provider returns a zero-result response
- Wrapping provider errors as `domain.NewProviderError(...)` so callers don't need to import the SDK

## Note on `placeId` portability

Google's `placeId` is specific to the Google Maps platform. If you switch providers, the IDs stored in the database (`saved_venues.place_id`, `venue_visits.place_id`) will be incompatible with the new provider. Plan a migration strategy before switching providers in production.
