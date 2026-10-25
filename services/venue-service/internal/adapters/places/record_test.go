package places

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SamPariat/roundup/services/venue-service/internal/domain"
)

// fixture is one recorded answer from Google: its status and JSON body.
type fixture struct {
	Status int             `json:"status"`
	Body   json.RawMessage `json:"body"`
}

// keptPlaces is how many places a recorded page keeps; a full page is
// ~260 KB, mostly photo credits.
const keptPlaces = 3

// bkc is the search every fixture is recorded around.
var bkc = domain.SearchParams{Latitude: 19.076, Longitude: 72.8777, RadiusInMeters: 2000}

// TestRecordGoogleResponses re-records testdata/ from the real Places API
// (New). It runs only when asked:
//
//	PLACES_RECORD_KEY=<key> go test ./internal/adapters/places -run TestRecordGoogleResponses
//
// The key is sent to Google only; it is never written to a fixture.
func TestRecordGoogleResponses(t *testing.T) {
	key := os.Getenv("PLACES_RECORD_KEY")
	if key == "" {
		t.Skip("set PLACES_RECORD_KEY to re-record testdata/ from Google")
	}

	var name string

	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status, body, err := forward(r)
		if !assert.NoError(t, err) {
			return
		}

		assert.NoError(t, save(name, status, body))
		w.WriteHeader(status)
		_, _ = w.Write(body)
	}))
	defer proxy.Close()

	a := NewGooglePlacesAdapter(proxy.URL, key)
	ctx := t.Context()

	name = "search_text"
	first, err := a.SearchNearby(ctx, bkc)
	require.NoError(t, err)
	require.NotEmpty(t, first.Venues)

	name = "search_text_bars"
	_, err = a.SearchNearby(ctx, domain.SearchParams{Latitude: bkc.Latitude, Longitude: bkc.Longitude, RadiusInMeters: 2000, Type: "bar"})
	require.NoError(t, err)

	name = "search_text_empty"
	_, err = a.SearchNearby(ctx, domain.SearchParams{Latitude: -48.876, Longitude: -123.393, RadiusInMeters: 1000, Query: "rooftop bar"})
	require.NoError(t, err)

	name = "search_text_bad_cursor"
	_, _ = a.SearchNearby(ctx, domain.SearchParams{Latitude: bkc.Latitude, Longitude: bkc.Longitude, RadiusInMeters: 2000, PageToken: "not-a-token"})

	name = "place_details"
	_, err = a.GetDetail(ctx, first.Venues[0].PlaceID)
	require.NoError(t, err)

	name = "place_details_unknown"
	_, _ = a.GetDetail(ctx, "ChIJ_not_a_real_place")

	name = "invalid_key"
	_, _ = NewGooglePlacesAdapter(proxy.URL, "not-a-key").GetDetail(ctx, first.Venues[0].PlaceID)
}

// forward sends the adapter's request on to Google and returns its answer.
func forward(r *http.Request) (int, []byte, error) {
	req, err := http.NewRequestWithContext(r.Context(), r.Method, BaseURL+r.URL.RequestURI(), r.Body)
	if err != nil {
		return 0, nil, err
	}

	req.Header = r.Header.Clone()
	req.Header.Del("Accept-Encoding") // let Go negotiate and unzip

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)

	return resp.StatusCode, body, err
}

// save writes testdata/<name>.json, keeping the first keptPlaces of a page.
func save(name string, status int, body []byte) error {
	var page map[string]json.RawMessage
	if err := json.Unmarshal(body, &page); err != nil {
		return err
	}

	if raw, ok := page["places"]; ok {
		var places []json.RawMessage
		if err := json.Unmarshal(raw, &places); err != nil {
			return err
		}

		page["places"], _ = json.Marshal(places[:min(len(places), keptPlaces)])
	}

	trimmed, err := json.Marshal(page)
	if err != nil {
		return err
	}

	var pretty bytes.Buffer
	if err := json.Indent(&pretty, trimmed, "", "  "); err != nil {
		return err
	}

	saved, err := json.MarshalIndent(fixture{Status: status, Body: pretty.Bytes()}, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join("testdata", name+".json"), append(saved, '\n'), 0o600)
}
