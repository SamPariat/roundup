package places

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SamPariat/roundup/services/venue-service/internal/domain"
)

const testKey = "test-key"

// replay serves the recorded answer testdata/<name>.json for every request,
// after check has looked at the request.
func replay(t *testing.T, name string, check func(r *http.Request)) *GooglePlacesAdapter {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join("testdata", name+".json"))
	require.NoError(t, err)

	var f fixture
	require.NoError(t, json.Unmarshal(raw, &f))

	return fakeGoogle(t, func(w http.ResponseWriter, r *http.Request) {
		if check != nil {
			check(r)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(f.Status)
		_, _ = w.Write(f.Body)
	})
}

// fakeGoogle serves handler and returns an adapter pointed at it.
func fakeGoogle(t *testing.T, handler http.HandlerFunc) *GooglePlacesAdapter {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	return NewGooglePlacesAdapter(srv.URL, testKey)
}

func TestSearchSendsATextSearchInsideTheRadius(t *testing.T) {
	var got searchTextRequest

	a := replay(t, "search_text", func(r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/places:searchText", r.URL.Path)
		assert.Equal(t, searchFields, r.URL.Query().Get("fields"))
		assert.Equal(t, testKey, r.Header.Get("X-Goog-Api-Key"))
		assert.Empty(t, r.URL.Query().Get("key"), "the key never goes in the URL")
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&got))
	})

	_, err := a.SearchNearby(t.Context(), bkc)
	require.NoError(t, err)

	assert.Equal(t, defaultTextQuery, got.TextQuery)
	assert.Equal(t, pageSize, got.PageSize)
	assert.Empty(t, got.IncludedType)

	box := got.LocationRestriction.Rectangle
	assert.Less(t, box.Low.Latitude, bkc.Latitude)
	assert.Greater(t, box.High.Latitude, bkc.Latitude)
	assert.Less(t, box.Low.Longitude, bkc.Longitude)
	assert.Greater(t, box.High.Longitude, bkc.Longitude)
}

func TestSearchMapsEveryRecordedPlace(t *testing.T) {
	for _, name := range []string{"search_text", "search_text_bars"} {
		body := rawBody(t, name)

		res, err := replay(t, name, nil).SearchNearby(t.Context(), bkc)
		require.NoError(t, err)

		assert.Equal(t, body["nextPageToken"], res.NextPageToken, name)

		places := body["places"].([]any)
		require.Len(t, res.Venues, len(places), name)

		for i, v := range res.Venues {
			assertMapped(t, places[i].(map[string]any), v)
		}
	}
}

func TestPriceLevels(t *testing.T) {
	for want, level := range []string{"PRICE_LEVEL_FREE", "PRICE_LEVEL_INEXPENSIVE", "PRICE_LEVEL_MODERATE", "PRICE_LEVEL_EXPENSIVE", "PRICE_LEVEL_VERY_EXPENSIVE"} {
		assert.Equal(t, want, place{PriceLevel: level}.toVenue().PriceLevel, level)
	}

	assert.Zero(t, place{}.toVenue().PriceLevel, "no price level is 0")
	assert.Zero(t, place{PriceLevel: "PRICE_LEVEL_UNSPECIFIED"}.toVenue().PriceLevel)
}

func TestOpenStatusIsUnknownWithoutCurrentHours(t *testing.T) {
	assert.Nil(t, place{}.toVenue().IsOpen, "no current hours means unknown, not closed")
}

func TestSearchWithTypeFiltersStrictly(t *testing.T) {
	var got searchTextRequest

	a := replay(t, "search_text_bars", func(r *http.Request) {
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&got))
	})

	_, err := a.SearchNearby(t.Context(), domain.SearchParams{Latitude: 1, Longitude: 1, RadiusInMeters: 10, Type: "bar", PageToken: "t"})
	require.NoError(t, err)
	assert.Equal(t, "bar", got.TextQuery)
	assert.Equal(t, "bar", got.IncludedType)
	assert.True(t, got.StrictTypeFiltering)
	assert.Equal(t, "t", got.PageToken)
}

func TestSearchEmptyAreaIsAnEmptyPage(t *testing.T) {
	res, err := replay(t, "search_text_empty", nil).SearchNearby(t.Context(), bkc)
	require.NoError(t, err)
	assert.NotNil(t, res.Venues)
	assert.Empty(t, res.Venues)
	assert.Empty(t, res.NextPageToken)
}

func TestSearchBadCursor(t *testing.T) {
	_, err := replay(t, "search_text_bad_cursor", nil).SearchNearby(t.Context(), domain.SearchParams{PageToken: "stale"})
	require.ErrorIs(t, err, domain.ErrInvalidSearch)
	assert.ErrorContains(t, err, "cursor is invalid or expired")
}

func TestTextQuery(t *testing.T) {
	assert.Equal(t, "rooftop bar", textQuery(domain.SearchParams{Query: "rooftop bar", Type: "bar"}))
	assert.Equal(t, "bar and grill", textQuery(domain.SearchParams{Type: "bar_and_grill"}))
	assert.Equal(t, defaultTextQuery, textQuery(domain.SearchParams{Query: "  "}))
}

func TestBoundingBox(t *testing.T) {
	box := boundingBox(domain.SearchParams{Latitude: 0, Longitude: 179.99, RadiusInMeters: 5000})
	assert.Greater(t, box.Low.Longitude, box.High.Longitude, "a low longitude above the high one crosses 180°")
	assert.GreaterOrEqual(t, box.High.Longitude, -maxLongitude)

	polar := boundingBox(domain.SearchParams{Latitude: maxLatitude, Longitude: 0, RadiusInMeters: 50000})
	assert.LessOrEqual(t, polar.High.Latitude, maxLatitude)
	assert.InDelta(t, -maxHalfWidth, polar.Low.Longitude, 0.001)
}

func TestGetDetail(t *testing.T) {
	body := rawBody(t, "place_details")
	id := body["id"].(string)

	a := replay(t, "place_details", func(r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/places/"+id, r.URL.Path)
		assert.Equal(t, detailFields, r.URL.Query().Get("fields"))
	})

	d, err := a.GetDetail(t.Context(), id)
	require.NoError(t, err)

	assertMapped(t, body, d.Venue)
	assert.Equal(t, body["internationalPhoneNumber"], d.PhoneNumber)
	assert.Equal(t, body["websiteUri"], d.Website)
	assert.Equal(t, get(body, "editorialSummary", "text"), d.EditorialSummary)

	hours := get(body, "regularOpeningHours", "weekdayDescriptions").([]any)
	require.Len(t, d.OpeningHours, len(hours))

	for i, h := range hours {
		assert.Equal(t, h, d.OpeningHours[i])
	}

	reviews := body["reviews"].([]any)
	require.Len(t, d.Reviews, len(reviews))

	for i, raw := range reviews {
		r := raw.(map[string]any)
		published, err := time.Parse(time.RFC3339Nano, r["publishTime"].(string))
		require.NoError(t, err)

		assert.Equal(t, get(r, "authorAttribution", "displayName"), d.Reviews[i].AuthorName)
		assert.Equal(t, get(r, "text", "text"), d.Reviews[i].Text)
		assert.InDelta(t, r["rating"], d.Reviews[i].Rating, 0.001)
		assert.Equal(t, published, d.Reviews[i].PublishedAt)
	}
}

func TestGetDetailUnknownPlace(t *testing.T) {
	_, err := replay(t, "place_details_unknown", nil).GetDetail(t.Context(), "ChIJ_not_a_real_place")
	require.ErrorIs(t, err, domain.ErrVenueNotFound, "Google answers an unknown ID with a 400")
}

func TestInvalidKeyIsUnavailableNotTheCallersFault(t *testing.T) {
	a := replay(t, "invalid_key", nil)

	_, err := a.GetDetail(t.Context(), "ChIJOfd6YX3I5zsRwtCk41iQ47Y")
	require.ErrorIs(t, err, domain.ErrProviderUnavailable)

	_, err = a.SearchNearby(t.Context(), bkc)
	require.ErrorIs(t, err, domain.ErrProviderUnavailable)
}

func TestProviderStatusesAreUnavailable(t *testing.T) {
	for _, status := range []int{http.StatusForbidden, http.StatusTooManyRequests, http.StatusServiceUnavailable} {
		a := fakeGoogle(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(status)
			_, _ = fmt.Fprintf(w, `{"error":{"code":%d,"message":"nope"}}`, status)
		})

		_, err := a.SearchNearby(t.Context(), bkc)
		require.ErrorIs(t, err, domain.ErrProviderUnavailable, status)
	}
}

func TestGoogleUnreachableIsUnavailable(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	srv.Close()

	_, err := NewGooglePlacesAdapter(srv.URL, testKey).GetDetail(t.Context(), "p1")
	require.ErrorIs(t, err, domain.ErrProviderUnavailable)
}

func TestPhotoFetchesTheMedia(t *testing.T) {
	a := fakeGoogle(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/places/p1/photos/abc/media", r.URL.Path)
		assert.Equal(t, "400", r.URL.Query().Get("maxWidthPx"))
		w.Header().Set("Content-Type", "image/jpeg")
		_, _ = io.WriteString(w, "jpeg-bytes")
	})

	photo, err := a.Photo(t.Context(), encodePhotoReference("places/p1/photos/abc"), 400)
	require.NoError(t, err)

	defer func() { _ = photo.Body.Close() }()

	body, err := io.ReadAll(photo.Body)
	require.NoError(t, err)
	assert.Equal(t, "image/jpeg", photo.ContentType)
	assert.Equal(t, "jpeg-bytes", string(body))
}

func TestPhotoRejectsReferencesItDidNotIssue(t *testing.T) {
	a := fakeGoogle(t, func(_ http.ResponseWriter, r *http.Request) {
		t.Errorf("no request should reach Google, got %s", r.URL.Path)
	})

	for _, ref := range []string{
		"AWU5eFjLegacyReference", // saved before the move to Places API (New)
		encodePhotoReference("places/p1"),
		encodePhotoReference("places/p1/photos/abc/../../x"),
		"",
	} {
		_, err := a.Photo(t.Context(), ref, 400)
		require.ErrorIs(t, err, domain.ErrPhotoNotFound, ref)
	}
}

func TestPhotoGoogleNoLongerHas(t *testing.T) {
	a := fakeGoogle(t, func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNotFound) })

	_, err := a.Photo(t.Context(), encodePhotoReference("places/p1/photos/gone"), 400)
	require.ErrorIs(t, err, domain.ErrPhotoNotFound)
}

// rawBody reads testdata/<name>.json's body as plain JSON, independent of the
// adapter's own types, so a wrong JSON tag can't hide in both.
func rawBody(t *testing.T, name string) map[string]any {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join("testdata", name+".json"))
	require.NoError(t, err)

	var f struct {
		Body map[string]any `json:"body"`
	}
	require.NoError(t, json.Unmarshal(raw, &f))

	return f.Body
}

// get walks keys into nested JSON objects; nil if any is missing.
func get(v any, keys ...string) any {
	for _, k := range keys {
		m, ok := v.(map[string]any)
		if !ok {
			return nil
		}

		v = m[k]
	}

	return v
}

// assertMapped checks v against the raw place p that Google sent.
func assertMapped(t *testing.T, p map[string]any, v domain.Venue) {
	t.Helper()

	assert.Equal(t, p["id"], v.PlaceID)
	assert.Equal(t, get(p, "displayName", "text"), v.Name)
	assert.Equal(t, p["formattedAddress"], v.Address)
	assert.InDelta(t, get(p, "location", "latitude"), v.Latitude, 1e-9)
	assert.InDelta(t, get(p, "location", "longitude"), v.Longitude, 1e-9)
	assert.InDelta(t, p["rating"], v.AverageRating, 0.001)
	assert.InDelta(t, p["userRatingCount"], v.RatingCount, 0)
	assert.Equal(t, get(p, "currentOpeningHours", "openNow") == nil, v.IsOpen == nil, "IsOpen is set only when Google sends it")

	photos, _ := p["photos"].([]any)
	require.Len(t, v.PhotoRefs, len(photos))

	for i, raw := range photos {
		name, ok := decodePhotoReference(v.PhotoRefs[i].Reference)
		require.True(t, ok)
		assert.Equal(t, get(raw, "name"), name)
		assert.InDelta(t, get(raw, "widthPx"), v.PhotoRefs[i].Width, 0)
		assert.InDelta(t, get(raw, "heightPx"), v.PhotoRefs[i].Height, 0)
	}
}
