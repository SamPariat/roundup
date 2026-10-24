package http

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/SamPariat/roundup/services/gateway/internal/config"
	"github.com/SamPariat/roundup/services/gateway/internal/middleware"
	"github.com/SamPariat/roundup/shared/api"
	"github.com/SamPariat/roundup/shared/authtoken"
	sharedmw "github.com/SamPariat/roundup/shared/middleware"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const userID = "7dc92d84-1950-43d6-b676-debacf84d830"

// seen is what a fake upstream received.
type seen struct {
	Service, Method, URI, UserID, RequestID, Body string
}

// upstream is a fake service that records each request and answers 200.
type upstream struct {
	name string
	srv  *httptest.Server
	mu   sync.Mutex
	got  []seen
}

func newUpstream(t *testing.T, name string) *upstream {
	t.Helper()

	u := &upstream{name: name}
	u.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)

		u.mu.Lock()
		u.got = append(u.got, seen{name, r.Method, r.URL.RequestURI(), r.Header.Get(sharedmw.UserIDHeader), r.Header.Get(sharedmw.RequestIDHeader), string(body)})
		u.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"from": name})
	}))
	t.Cleanup(u.srv.Close)

	return u
}

func (u *upstream) last(t *testing.T) seen {
	t.Helper()
	u.mu.Lock()
	defer u.mu.Unlock()

	require.NotEmpty(t, u.got, "%s received nothing", u.name)

	return u.got[len(u.got)-1]
}

type fixture struct {
	app               *fiber.App
	auth, user, venue *upstream
	token             string
}

func newFixture(t *testing.T) *fixture {
	t.Helper()

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	token, _, err := authtoken.Sign(priv, userID, time.Minute)
	require.NoError(t, err)

	f := &fixture{auth: newUpstream(t, "auth"), user: newUpstream(t, "user"), venue: newUpstream(t, "venue"), token: token}
	f.app = fiber.New()
	f.app.Use(sharedmw.GenerateRequestID(), middleware.StripIdentity())
	RegisterRoutes(f.app, config.Upstreams{Auth: f.auth.srv.URL, User: f.user.srv.URL, Venue: f.venue.srv.URL}, pub, 2*time.Second)

	return f
}

func (f *fixture) do(t *testing.T, method, path, body string, withToken bool) *http.Response {
	t.Helper()

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	if withToken {
		req.Header.Set("Authorization", "Bearer "+f.token)
	}

	resp, err := f.app.Test(req)
	require.NoError(t, err)

	return resp
}

func TestRoutesToUpstreams(t *testing.T) {
	tests := []struct {
		name, method, path, body string
		withToken                bool
		target                   func(f *fixture) *upstream
		wantUserID               string
	}{
		{"AuthIsPublic", http.MethodPost, "/api/v1/auth/otp/request", `{"phone":"+919876543210"}`, false, func(f *fixture) *upstream { return f.auth }, ""},
		{"UsersMe", http.MethodGet, "/api/v1/users/me", "", true, func(f *fixture) *upstream { return f.user }, userID},
		{"DeviceTokens", http.MethodDelete, "/api/v1/users/me/device-tokens/abc", "", true, func(f *fixture) *upstream { return f.user }, userID},
		{"VenueSearchKeepsQuery", http.MethodGet, "/api/v1/venues/?lat=12.97&lng=77.59&radius=2000", "", true, func(f *fixture) *upstream { return f.venue }, userID},
		{"SquadFavorites", http.MethodPost, "/api/v1/squads/s1/favorites/", `{"placeID":"p1"}`, true, func(f *fixture) *upstream { return f.venue }, userID},
		{"SquadFavoriteCheck", http.MethodGet, "/api/v1/squads/s1/favorites/p1/check", "", true, func(f *fixture) *upstream { return f.venue }, userID},
		{"SquadHistory", http.MethodGet, "/api/v1/squads/s1/history", "", true, func(f *fixture) *upstream { return f.venue }, userID},
		{"CreateSquad", http.MethodPost, "/api/v1/squads", `{"name":"Friday crew"}`, true, func(f *fixture) *upstream { return f.user }, userID},
		{"ListSquads", http.MethodGet, "/api/v1/squads/", "", true, func(f *fixture) *upstream { return f.user }, userID},
		{"GetSquad", http.MethodGet, "/api/v1/squads/s1", "", true, func(f *fixture) *upstream { return f.user }, userID},
		{"ChangeSquadRole", http.MethodPatch, "/api/v1/squads/s1/members/u1", `{"role":"ADMIN"}`, true, func(f *fixture) *upstream { return f.user }, userID},
		{"UserLookup", http.MethodPost, "/api/v1/users/lookup", `{"phones":["+919876543210"]}`, true, func(f *fixture) *upstream { return f.user }, userID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)

			resp := f.do(t, tt.method, tt.path, tt.body, tt.withToken)

			require.Equal(t, http.StatusOK, resp.StatusCode)
			got := tt.target(f).last(t)
			assert.Equal(t, tt.method, got.Method)
			assert.Equal(t, tt.path, got.URI, "path and query forwarded unchanged")
			assert.Equal(t, tt.body, got.Body)
			assert.Equal(t, tt.wantUserID, got.UserID)
			assert.NotEmpty(t, got.RequestID, "request ID is forwarded")
			assert.Equal(t, got.RequestID, resp.Header.Get(sharedmw.RequestIDHeader), "caller sees the same request ID")
		})
	}
}

func TestProtectedRoutesNeedToken(t *testing.T) {
	f := newFixture(t)

	for _, path := range []string{"/api/v1/users/me", "/api/v1/venues/", "/api/v1/squads/s1/history/", "/api/v1/squads", "/api/v1/squads/s1/members"} {
		resp := f.do(t, http.MethodGet, path, "", false)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, path)
	}

	assert.Empty(t, f.user.got)
	assert.Empty(t, f.venue.got)
}

func TestCreateUserIsBlocked(t *testing.T) {
	f := newFixture(t)

	for _, path := range []string{"/api/v1/users", "/api/v1/users/"} {
		resp := f.do(t, http.MethodPost, path, `{"phone":"+919876543210"}`, true)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode, path)
	}

	assert.Empty(t, f.user.got, "user-service must never see an external create")
}

func TestUpstreamDownIsBadGateway(t *testing.T) {
	f := newFixture(t)
	f.user.srv.Close()

	resp := f.do(t, http.MethodGet, "/api/v1/users/me", "", true)

	assert.Equal(t, http.StatusBadGateway, resp.StatusCode)
}

func TestUnbuiltServices(t *testing.T) {
	f := newFixture(t)

	assert.Equal(t, http.StatusNotImplemented, f.do(t, http.MethodGet, "/api/v1/events/123", "", false).StatusCode)
}

func TestClientRequestIDIsReplaced(t *testing.T) {
	f := newFixture(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+f.token)
	req.Header.Set(sharedmw.RequestIDHeader, "client-chosen-id")

	resp, err := f.app.Test(req)
	require.NoError(t, err)

	got := f.user.last(t).RequestID
	require.NoError(t, uuid.Validate(got), "the gateway generates a UUID")
	assert.NotEqual(t, "client-chosen-id", got, "a client-sent request ID never reaches upstream")
	assert.Equal(t, got, resp.Header.Get(sharedmw.RequestIDHeader), "caller sees the gateway's ID")
}

func TestRegisterDocs(t *testing.T) {
	app := fiber.New()
	RegisterDocs(app)

	page, err := app.Test(httptest.NewRequest(http.MethodGet, "/docs", nil))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, page.StatusCode)
	assert.Contains(t, page.Header.Get(fiber.HeaderContentType), "text/html")

	spec, err := app.Test(httptest.NewRequest(http.MethodGet, "/docs/openapi.yaml", nil))
	require.NoError(t, err)

	body, _ := io.ReadAll(spec.Body)
	assert.Equal(t, http.StatusOK, spec.StatusCode)
	assert.Equal(t, api.Spec, body)
}
