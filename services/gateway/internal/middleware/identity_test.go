package middleware

import (
	"crypto/ed25519"
	"crypto/rand"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/SamPariat/roundup/shared/authtoken"
	sharedmw "github.com/SamPariat/roundup/shared/middleware"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const userID = "7dc92d84-1950-43d6-b676-debacf84d830"

// echoApp returns an app whose only route replies with the X-User-ID it received.
func echoApp(pub ed25519.PublicKey) *fiber.App {
	app := fiber.New()
	app.Use(StripInternalHeaders())
	app.Get("/private", RequireAuth(pub), func(c fiber.Ctx) error {
		return c.SendString(c.Get(sharedmw.UserIDHeader))
	})
	app.Get("/public", func(c fiber.Ctx) error {
		return c.SendString(c.Get(sharedmw.UserIDHeader))
	})
	return app
}

func call(t *testing.T, app *fiber.App, path string, headers map[string]string) (int, string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := app.Test(req)
	require.NoError(t, err)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp.StatusCode, string(body)
}

func TestRequireAuth(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	_, otherPriv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	valid, _, err := authtoken.Sign(priv, userID, time.Minute)
	require.NoError(t, err)
	expired, _, err := authtoken.Sign(priv, userID, -time.Minute)
	require.NoError(t, err)
	forged, _, err := authtoken.Sign(otherPriv, userID, time.Minute)
	require.NoError(t, err)

	tests := []struct {
		name     string
		headers  map[string]string
		wantCode int
		wantBody string
	}{
		{name: "ValidTokenForwardsUser", headers: map[string]string{"Authorization": "Bearer " + valid}, wantCode: 200, wantBody: userID},
		{name: "SpoofedHeaderReplacedByTokenUser", headers: map[string]string{"Authorization": "Bearer " + valid, sharedmw.UserIDHeader: "00000000-0000-0000-0000-000000000000"}, wantCode: 200, wantBody: userID},
		{name: "NoToken", wantCode: 401},
		{name: "SpoofedHeaderWithoutToken", headers: map[string]string{sharedmw.UserIDHeader: userID}, wantCode: 401},
		{name: "Expired", headers: map[string]string{"Authorization": "Bearer " + expired}, wantCode: 401},
		{name: "SignedByAnotherKey", headers: map[string]string{"Authorization": "Bearer " + forged}, wantCode: 401},
		{name: "NotBearer", headers: map[string]string{"Authorization": "Basic " + valid}, wantCode: 401},
		{name: "EmptyBearer", headers: map[string]string{"Authorization": "Bearer "}, wantCode: 401},
	}

	app := echoApp(pub)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, body := call(t, app, "/private", tt.headers)

			assert.Equal(t, tt.wantCode, code)
			if tt.wantBody != "" {
				assert.Equal(t, tt.wantBody, body)
			}
		})
	}
}

func TestStripInternalHeadersRemovesUserIDOnPublicRoutes(t *testing.T) {
	pub, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	code, body := call(t, echoApp(pub), "/public", map[string]string{sharedmw.UserIDHeader: userID})

	assert.Equal(t, 200, code)
	assert.Empty(t, body, "a client-sent X-User-ID must never pass through")
}
