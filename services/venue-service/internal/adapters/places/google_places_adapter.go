package places

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/SamPariat/roundup/services/venue-service/internal/domain"
	"github.com/SamPariat/roundup/shared/domainerr"
	"github.com/SamPariat/roundup/shared/serviceclient"
)

// BaseURL is Places API (New).
const BaseURL = "https://places.googleapis.com/v1"

// requestTimeout bounds every call to Google, photo downloads included.
const requestTimeout = 10 * time.Second

// defaultTextQuery is what a search looks for when the caller typed nothing.
const defaultTextQuery = "bars and restaurants"

// pageSize is the most results Text Search returns per page.
const pageSize = 20

// Field masks, sent as ?fields=: Google bills by the most expensive field
// asked for, and returns only what is listed.
const (
	searchFields = "nextPageToken,places.id,places.displayName,places.formattedAddress,places.location," +
		"places.rating,places.userRatingCount,places.priceLevel,places.types,places.currentOpeningHours.openNow,places.photos"
	detailFields = "id,displayName,formattedAddress,location,rating,userRatingCount,priceLevel,types," +
		"currentOpeningHours.openNow,regularOpeningHours.weekdayDescriptions,photos," +
		"internationalPhoneNumber,websiteUri,reviews,editorialSummary"
)

// Geometry for the search box.
const (
	// metresPerDegree is the length of one degree of latitude.
	metresPerDegree = 111_320.0
	maxLatitude     = 90.0
	maxLongitude    = 180.0
	// maxHalfWidth keeps the box under Google's 180° width limit near the poles.
	maxHalfWidth = 89.0
)

// photoName is the only shape of photo name the adapter will fetch.
var photoName = regexp.MustCompile(`^places/[^/]+/photos/[^/]+$`)

// GooglePlacesAdapter implements domain.PlaceProvider with Places API (New).
type GooglePlacesAdapter struct {
	client *serviceclient.Client
}

// NewGooglePlacesAdapter calls Places API (New) at baseURL (BaseURL outside
// tests) with apiKey, sent as a header so it never appears in a URL or log.
func NewGooglePlacesAdapter(baseURL, apiKey string) *GooglePlacesAdapter {
	return &GooglePlacesAdapter{
		client: serviceclient.New(baseURL, requestTimeout, serviceclient.WithHeader("X-Goog-Api-Key", apiKey)),
	}
}

// SearchNearby runs a Text Search for one page (up to 20) of venues inside the
// square around params' circle: params.Query, or the type, or
// defaultTextQuery when both are empty. An area with no matches returns an
// empty page. A search Google rejects, such as a stale cursor, is
// domain.ErrInvalidSearch.
func (a *GooglePlacesAdapter) SearchNearby(ctx context.Context, params domain.SearchParams) (*domain.SearchResult, error) {
	body := searchTextRequest{
		TextQuery:           textQuery(params),
		IncludedType:        params.Type,
		StrictTypeFiltering: params.Type != "",
		PageSize:            pageSize,
		PageToken:           params.PageToken,
		LocationRestriction: locationRestriction{Rectangle: boundingBox(params)},
	}

	var res searchTextResponse
	if err := a.do(ctx, http.MethodPost, "/places:searchText?fields="+searchFields, body, &res); err != nil {
		var apiErr *apiError
		if errors.As(err, &apiErr) && apiErr.status == http.StatusBadRequest && !apiErr.keyInvalid() {
			detail := apiErr.message
			if params.PageToken != "" {
				detail = "cursor is invalid or expired" // Google says "decryption failed"
			}

			return nil, domainerr.New(domain.ErrInvalidSearch, detail)
		}

		return nil, mapProviderError(err)
	}

	venues := make([]domain.Venue, 0, len(res.Places))
	for _, p := range res.Places {
		venues = append(venues, p.toVenue())
	}

	return &domain.SearchResult{Venues: venues, NextPageToken: res.NextPageToken}, nil
}

// GetDetail fetches full venue information for placeID. Returns
// domain.ErrVenueNotFound if the place does not exist or the ID is malformed.
func (a *GooglePlacesAdapter) GetDetail(ctx context.Context, placeID string) (*domain.VenueDetail, error) {
	var p place
	if err := a.do(ctx, http.MethodGet, "/places/"+url.PathEscape(placeID)+"?fields="+detailFields, nil, &p); err != nil {
		if isNotFound(err) {
			return nil, domainerr.New(domain.ErrVenueNotFound, placeID)
		}

		return nil, mapProviderError(err)
	}

	return p.toVenueDetail(), nil
}

// Photo streams a venue photo at most maxWidth pixels wide, so the API key
// never reaches clients. Returns domain.ErrPhotoNotFound for a reference this
// adapter did not issue or Google no longer knows.
func (a *GooglePlacesAdapter) Photo(ctx context.Context, reference string, maxWidth int) (*domain.Photo, error) {
	name, ok := decodePhotoReference(reference)
	if !ok {
		return nil, domainerr.New(domain.ErrPhotoNotFound, reference)
	}

	resp, err := a.send(ctx, http.MethodGet, "/"+name+"/media?maxWidthPx="+strconv.Itoa(maxWidth), nil)
	if err != nil {
		if isNotFound(err) {
			return nil, domainerr.New(domain.ErrPhotoNotFound, reference)
		}

		return nil, mapProviderError(err)
	}

	return &domain.Photo{ContentType: resp.Header.Get("Content-Type"), Body: resp.Body}, nil
}

// do sends a request and decodes a 2xx JSON answer into out.
func (a *GooglePlacesAdapter) do(ctx context.Context, method, path string, body, out any) error {
	resp, err := a.send(ctx, method, path, body)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("places: decode %s: %w", path, err)
	}

	return nil
}

// send returns the response for any 2xx status. A network failure is
// domain.ErrProviderUnavailable; any other status is an *apiError.
func (a *GooglePlacesAdapter) send(ctx context.Context, method, path string, body any) (*http.Response, error) {
	resp, err := a.client.Do(ctx, method, path, "", body)
	if err != nil {
		return nil, domainerr.New(domain.ErrProviderUnavailable, err.Error())
	}

	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return resp, nil
	}

	defer func() { _ = resp.Body.Close() }()

	return nil, readAPIError(resp)
}

// textQuery is what the search looks for: the caller's words, else the
// type ("bar_and_grill" → "bar and grill"), else defaultTextQuery.
func textQuery(params domain.SearchParams) string {
	switch {
	case strings.TrimSpace(params.Query) != "":
		return params.Query
	case params.Type != "":
		return strings.ReplaceAll(params.Type, "_", " ")
	default:
		return defaultTextQuery
	}
}

// boundingBox is the square that contains params' circle. Text Search only
// restricts to rectangles; a low longitude above the high one crosses 180°.
func boundingBox(params domain.SearchParams) rectangle {
	radius := float64(params.RadiusInMeters)
	dLat := radius / metresPerDegree
	dLng := math.Min(radius/(metresPerDegree*math.Cos(params.Latitude*math.Pi/maxLongitude)), maxHalfWidth)

	return rectangle{
		Low:  latLng{Latitude: math.Max(params.Latitude-dLat, -maxLatitude), Longitude: wrapLongitude(params.Longitude - dLng)},
		High: latLng{Latitude: math.Min(params.Latitude+dLat, maxLatitude), Longitude: wrapLongitude(params.Longitude + dLng)},
	}
}

// wrapLongitude brings lng back into -180–180.
func wrapLongitude(lng float64) float64 {
	switch {
	case lng > maxLongitude:
		return lng - 2*maxLongitude
	case lng < -maxLongitude:
		return lng + 2*maxLongitude
	default:
		return lng
	}
}

// encodePhotoReference turns a Google photo name, which contains slashes,
// into one URL path segment.
func encodePhotoReference(name string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(name))
}

// decodePhotoReference reverses encodePhotoReference, accepting only a photo
// name, so a crafted reference can't reach any other Google path.
func decodePhotoReference(reference string) (string, bool) {
	raw, err := base64.RawURLEncoding.DecodeString(reference)
	if err != nil || !photoName.Match(raw) {
		return "", false
	}

	return string(raw), true
}

// mapProviderError turns an adapter error into a domain error. Bad keys,
// quota and Google's own failures are domain.ErrProviderUnavailable; anything
// else surfaces as a 500.
func mapProviderError(err error) error {
	var apiErr *apiError
	if !errors.As(err, &apiErr) {
		return err
	}

	if apiErr.keyInvalid() ||
		apiErr.status == http.StatusUnauthorized ||
		apiErr.status == http.StatusForbidden ||
		apiErr.status == http.StatusTooManyRequests ||
		apiErr.status >= http.StatusInternalServerError {
		return domainerr.New(domain.ErrProviderUnavailable, apiErr.Error())
	}

	return fmt.Errorf("places provider: %w", err)
}

// isNotFound reports whether Google said the place or photo doesn't exist:
// a 404, or a 400 for an ID it can't parse.
func isNotFound(err error) bool {
	var apiErr *apiError

	return errors.As(err, &apiErr) && !apiErr.keyInvalid() &&
		(apiErr.status == http.StatusNotFound || apiErr.status == http.StatusBadRequest)
}
