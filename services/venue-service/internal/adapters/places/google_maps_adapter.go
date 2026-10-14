package places

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/SamPariat/roundup/services/venue-service/internal/domain"
	"googlemaps.github.io/maps"
)

// GoogleMapsAdapter implements domain.PlaceProvider using the Google Maps Places API.
type GoogleMapsAdapter struct {
	client *maps.Client
}

// NewGoogleMapsAdapter creates a GoogleMapsAdapter. It returns an error only if
// apiKey is empty; Google checks the key itself on the first request.
func NewGoogleMapsAdapter(apiKey string) (*GoogleMapsAdapter, error) {
	client, err := maps.NewClient(maps.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}

	return &GoogleMapsAdapter{client: client}, nil
}

// SearchNearby queries the Google Maps Places API for venues near the location in params.
// An area with no matches returns an empty slice: the library treats ZERO_RESULTS
// as success. Returns domain.ErrProviderUnavailable on auth, rate-limit or server errors.
func (a *GoogleMapsAdapter) SearchNearby(ctx context.Context, params domain.SearchParams) ([]domain.Venue, error) {
	req := &maps.NearbySearchRequest{
		Location: &maps.LatLng{Lat: params.Latitude, Lng: params.Longitude},
		Radius:   uint(params.RadiusInMeters),
		Keyword:  params.Query,
		Type:     maps.PlaceType(params.Type),
	}

	res, err := a.client.NearbySearch(ctx, req)
	if err != nil {
		return nil, mapProviderError(err)
	}

	venues := make([]domain.Venue, 0, len(res.Results))
	for _, r := range res.Results {
		venues = append(venues, mapPlaceResultToVenue(r))
	}

	return venues, nil
}

// GetDetail fetches full venue information for the given placeID from the Google Maps
// Places API. Only the fields required by domain.VenueDetail are requested to minimize
// billing costs. Returns domain.ErrVenueNotFound if the place does not exist or the
// ID is malformed.
func (a *GoogleMapsAdapter) GetDetail(ctx context.Context, placeID string) (*domain.VenueDetail, error) {
	req := &maps.PlaceDetailsRequest{
		PlaceID: placeID,
		Fields: []maps.PlaceDetailsFieldMask{
			maps.PlaceDetailsFieldMaskName,
			maps.PlaceDetailsFieldMaskFormattedAddress,
			maps.PlaceDetailsFieldMaskGeometry,
			maps.PlaceDetailsFieldMaskRatings,
			maps.PlaceDetailsFieldMaskPriceLevel,
			maps.PlaceDetailsFieldMaskTypes,
			maps.PlaceDetailsFieldMaskOpeningHours,
			maps.PlaceDetailsFieldMaskPhotos,
			maps.PlaceDetailsFieldMaskFormattedPhoneNumber,
			maps.PlaceDetailsFieldMaskWebsite,
			maps.PlaceDetailsFieldMaskReviews,
			maps.PlaceDetailsFieldMaskEditorialSummary,
		},
	}

	res, err := a.client.PlaceDetails(ctx, req)
	if err != nil {
		return nil, mapDetailError(err, placeID)
	}

	return mapPlaceDetailsToVenueDetail(res), nil
}

// mapPlaceResultToVenue converts a single Google Maps nearby-search result into a
// domain.Venue. IsOpen is set only when the provider returns opening hours.
func mapPlaceResultToVenue(res maps.PlacesSearchResult) domain.Venue {
	venue := domain.Venue{
		PlaceID:       res.PlaceID,
		Name:          res.Name,
		Address:       res.FormattedAddress,
		Latitude:      res.Geometry.Location.Lat,
		Longitude:     res.Geometry.Location.Lng,
		AverageRating: res.Rating,
		PriceLevel:    res.PriceLevel,
		Types:         res.Types,
	}

	if res.OpeningHours != nil {
		venue.IsOpen = res.OpeningHours.OpenNow
	}

	venue.PhotoRefs = make([]domain.PhotoRef, 0, len(res.Photos))

	for _, photo := range res.Photos {
		venue.PhotoRefs = append(venue.PhotoRefs, domain.PhotoRef{
			Reference: photo.PhotoReference,
			Width:     photo.Width,
			Height:    photo.Height,
		})
	}

	return venue
}

// mapPlaceDetailsToVenueDetail converts a Google Maps place-details result into a
// domain.VenueDetail. The common Venue fields are populated by reusing mapPlaceResultToVenue.
func mapPlaceDetailsToVenueDetail(res maps.PlaceDetailsResult) *domain.VenueDetail {
	base := mapPlaceResultToVenue(maps.PlacesSearchResult{
		PlaceID:          res.PlaceID,
		Name:             res.Name,
		FormattedAddress: res.FormattedAddress,
		Geometry:         res.Geometry,
		Rating:           res.Rating,
		PriceLevel:       res.PriceLevel,
		Types:            res.Types,
		OpeningHours:     res.OpeningHours,
		Photos:           res.Photos,
	})

	detail := &domain.VenueDetail{
		Venue:       base,
		PhoneNumber: res.FormattedPhoneNumber,
		Website:     res.Website,
	}

	if res.EditorialSummary != nil {
		detail.EditorialSummary = res.EditorialSummary.Overview
	}

	if res.OpeningHours != nil {
		detail.OpeningHours = res.OpeningHours.WeekdayText
	}

	detail.Reviews = make([]domain.Review, 0, len(res.Reviews))
	for _, r := range res.Reviews {
		detail.Reviews = append(detail.Reviews, domain.Review{
			AuthorName:  r.AuthorName,
			Rating:      float32(r.Rating),
			Text:        r.Text,
			PublishedAt: time.Unix(int64(r.Time), 0),
		})
	}

	return detail
}

// mapDetailError translates a place-details error. A place-details request has
// only one input, so INVALID_REQUEST means the place ID is malformed — the
// client asked for a venue that cannot exist.
func mapDetailError(err error, placeID string) error {
	if strings.Contains(err.Error(), "INVALID_REQUEST") {
		return domain.NewDomainError(domain.ErrVenueNotFound, placeID)
	}

	return mapProviderError(err)
}

// mapProviderError translates a Google Maps library error into a domain sentinel.
// NOT_FOUND → ErrVenueNotFound; auth, rate-limit and server errors → ErrProviderUnavailable.
// Unrecognized errors are wrapped with fmt.Errorf and surface as 500s at the handler.
// ZERO_RESULTS never reaches here: the library treats it as success.
func mapProviderError(err error) error {
	s := err.Error()

	if strings.Contains(s, "NOT_FOUND") {
		return domain.NewDomainError(domain.ErrVenueNotFound, s)
	}

	if strings.Contains(s, "REQUEST_DENIED") ||
		strings.Contains(s, "OVER_QUERY_LIMIT") ||
		strings.Contains(s, "UNKNOWN_ERROR") {
		return domain.NewDomainError(domain.ErrProviderUnavailable, s)
	}

	return fmt.Errorf("places provider: %w", err)
}
