package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/SamPariat/roundup/services/auth-service/internal/domain"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	expiresAt = time.Date(2026, 11, 15, 12, 0, 0, 0, time.UTC)
	loginPair = domain.TokenPair{
		AccessToken:           "access-1",
		AccessTokenExpiresAt:  expiresAt.Add(-720 * time.Hour),
		RefreshToken:          "refresh-1",
		RefreshTokenExpiresAt: expiresAt,
	}
)

type fakeOTP struct{}

func (fakeOTP) Request(context.Context, string) error { return nil }

func (fakeOTP) Verify(context.Context, string, string) (*domain.LoginResult, error) {
	return &domain.LoginResult{TokenPair: loginPair, UserID: "u1", IsNewUser: true}, nil
}

// fakeTokens accepts only the token "refresh-1", and records what it was given.
type fakeTokens struct {
	refreshedWith string
	loggedOutWith string
}

func (f *fakeTokens) Refresh(_ context.Context, token string) (*domain.TokenPair, error) {
	f.refreshedWith = token
	if token != "refresh-1" {
		return nil, domain.ErrInvalidRefreshToken
	}

	return &domain.TokenPair{
		AccessToken:           "access-2",
		AccessTokenExpiresAt:  loginPair.AccessTokenExpiresAt,
		RefreshToken:          "refresh-2",
		RefreshTokenExpiresAt: expiresAt,
	}, nil
}

func (f *fakeTokens) Logout(_ context.Context, token string) error {
	f.loggedOutWith = token
	if token != "refresh-1" {
		return domain.ErrInvalidRefreshToken
	}

	return nil
}

type call struct {
	path, body string
	web        bool
	cookie     string
}

func send(t *testing.T, h *AuthHandler, in call) *http.Response {
	t.Helper()

	app := newApp(h)
	req := httptest.NewRequest(http.MethodPost, in.path, strings.NewReader(in.body))

	if in.body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	if in.web {
		req.Header.Set("X-Client", "web")
	}

	if in.cookie != "" {
		req.Header.Set("Cookie", "roundup_refresh="+in.cookie)
	}

	resp, err := app.Test(req)
	require.NoError(t, err)

	return resp
}

// tokens decodes data.refreshToken and data.accessToken from a success envelope.
func tokens(t *testing.T, resp *http.Response) (access, refresh string, hasRefreshKey bool) {
	t.Helper()

	var body struct {
		Data map[string]any `json:"data"`
	}

	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))

	access, _ = body.Data["accessToken"].(string)
	refresh, hasRefreshKey = body.Data["refreshToken"].(string)

	return access, refresh, hasRefreshKey
}

func refreshCookieOf(t *testing.T, resp *http.Response) *http.Cookie {
	t.Helper()

	for _, c := range resp.Cookies() {
		if c.Name == "roundup_refresh" {
			return c
		}
	}

	return nil
}

func newApp(h *AuthHandler) *fiber.App {
	app := fiber.New()
	RegisterRoutes(app, &Handlers{Auth: h})

	return app
}

func newTestHandler(secure bool) (*AuthHandler, *fakeTokens) {
	tk := &fakeTokens{}

	return &AuthHandler{otp: fakeOTP{}, tokens: tk, cookieSecure: secure}, tk
}

const (
	verifyPath  = "/api/v1/auth/otp/verify"
	refreshPath = "/api/v1/auth/refresh"
	logoutPath  = "/api/v1/auth/logout"
	verifyBody  = `{"phone":"+919876543210","code":"123456"}`
)

func TestVerifyNativeGetsTokenInBodyAndNoCookie(t *testing.T) {
	h, _ := newTestHandler(true)

	resp := send(t, h, call{path: verifyPath, body: verifyBody})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	access, refresh, _ := tokens(t, resp)
	assert.Equal(t, "access-1", access)
	assert.Equal(t, "refresh-1", refresh)
	assert.Nil(t, refreshCookieOf(t, resp))
}

func TestVerifyWebGetsCookieAndNoTokenInBody(t *testing.T) {
	h, _ := newTestHandler(true)

	resp := send(t, h, call{path: verifyPath, body: verifyBody, web: true})

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var body struct {
		Data map[string]any `json:"data"`
	}

	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.NotContains(t, body.Data, "refreshToken")
	assert.Equal(t, "access-1", body.Data["accessToken"])
	assert.Contains(t, body.Data, "refreshTokenExpiresAt")
	assert.Equal(t, true, body.Data["isNewUser"])

	cookie := refreshCookieOf(t, resp)
	require.NotNil(t, cookie)
	assert.Equal(t, "refresh-1", cookie.Value)
	assert.Equal(t, "/api/v1/auth", cookie.Path)
	assert.True(t, cookie.HttpOnly)
	assert.True(t, cookie.Secure)
	assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
	assert.True(t, expiresAt.Equal(cookie.Expires), "expires with the refresh token, got %v", cookie.Expires)
}

func TestWebCookieIsNotSecureWhenConfiguredOff(t *testing.T) {
	h, _ := newTestHandler(false)

	cookie := refreshCookieOf(t, send(t, h, call{path: verifyPath, body: verifyBody, web: true}))

	require.NotNil(t, cookie)
	assert.False(t, cookie.Secure)
}

func TestRefreshNativeUsesBodyAndReturnsTokenInBody(t *testing.T) {
	h, tk := newTestHandler(true)

	resp := send(t, h, call{path: refreshPath, body: `{"refreshToken":"refresh-1"}`})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	_, refresh, _ := tokens(t, resp)
	assert.Equal(t, "refresh-1", tk.refreshedWith)
	assert.Equal(t, "refresh-2", refresh)
	assert.Nil(t, refreshCookieOf(t, resp))
}

func TestRefreshWebReadsCookieAndRotatesIt(t *testing.T) {
	h, tk := newTestHandler(true)

	resp := send(t, h, call{path: refreshPath, web: true, cookie: "refresh-1"})

	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "refresh-1", tk.refreshedWith)

	access, _, hasRefresh := tokens(t, resp)
	assert.Equal(t, "access-2", access)
	assert.False(t, hasRefresh, "the rotated token is only in the cookie")

	cookie := refreshCookieOf(t, resp)
	require.NotNil(t, cookie)
	assert.Equal(t, "refresh-2", cookie.Value)
	assert.True(t, cookie.HttpOnly)
	assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
}

func TestRefreshBodyTokenWinsOverCookie(t *testing.T) {
	h, tk := newTestHandler(true)

	send(t, h, call{path: refreshPath, body: `{"refreshToken":"refresh-1"}`, web: true, cookie: "other"})

	assert.Equal(t, "refresh-1", tk.refreshedWith)
}

func TestRefreshWithoutAnyTokenIsBadRequest(t *testing.T) {
	tests := []struct {
		name string
		in   call
	}{
		{"NoBodyNoCookie", call{path: refreshPath, web: true}},
		{"EmptyJSONNoCookie", call{path: refreshPath, body: `{}`}},
		{"BrokenJSON", call{path: refreshPath, body: `{`}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, tk := newTestHandler(true)

			resp := send(t, h, tt.in)

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			assert.Empty(t, tk.refreshedWith, "the use case is never called")
		})
	}
}

func TestRefreshWithBadCookieIsUnauthorized(t *testing.T) {
	h, _ := newTestHandler(true)

	resp := send(t, h, call{path: refreshPath, web: true, cookie: "stale"})

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestLogoutWebClearsCookieOnSuccess(t *testing.T) {
	h, tk := newTestHandler(true)

	resp := send(t, h, call{path: logoutPath, web: true, cookie: "refresh-1"})

	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Equal(t, "refresh-1", tk.loggedOutWith)
	assertCookieCleared(t, resp)
}

func TestLogoutWebClearsCookieOnUnauthorized(t *testing.T) {
	h, _ := newTestHandler(true)

	resp := send(t, h, call{path: logoutPath, web: true, cookie: "stale"})

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assertCookieCleared(t, resp)
}

func TestLogoutWebClearsCookieWhenNothingWasSent(t *testing.T) {
	h, _ := newTestHandler(true)

	resp := send(t, h, call{path: logoutPath, web: true})

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assertCookieCleared(t, resp)
}

func TestLogoutNativeUsesBodyAndSetsNoCookie(t *testing.T) {
	h, tk := newTestHandler(true)

	resp := send(t, h, call{path: logoutPath, body: `{"refreshToken":"refresh-1"}`})

	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Equal(t, "refresh-1", tk.loggedOutWith)
	assert.Nil(t, refreshCookieOf(t, resp))
}

func assertCookieCleared(t *testing.T, resp *http.Response) {
	t.Helper()

	cookie := refreshCookieOf(t, resp)
	require.NotNil(t, cookie, "a Set-Cookie clears the stored token")
	assert.Empty(t, cookie.Value)
	assert.Equal(t, "/api/v1/auth", cookie.Path)
	assert.Negative(t, cookie.MaxAge, "Max-Age=0 on the wire")
	assert.True(t, cookie.HttpOnly)
}
