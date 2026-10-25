package serviceclient

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	sharedmw "github.com/SamPariat/roundup/shared/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDoSendsIdentityRequestIDAndJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)

		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/things", r.URL.Path)
		assert.Equal(t, "u1", r.Header.Get(sharedmw.UserIDHeader))
		assert.Equal(t, "req-1", r.Header.Get(sharedmw.RequestIDHeader))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.JSONEq(t, `{"name":"x"}`, string(body))

		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	ctx := sharedmw.WithRequestID(t.Context(), "req-1")

	resp, err := New(server.URL, time.Second).Do(ctx, http.MethodPost, "/api/v1/things", "u1", map[string]string{"name": "x"})
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestDoWithoutUserOrBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.Header.Values(sharedmw.UserIDHeader))
		assert.Empty(t, r.Header.Get("Content-Type"))
		assert.Empty(t, r.Header.Get(sharedmw.RequestIDHeader))
	}))
	defer server.Close()

	resp, err := New(server.URL, time.Second).Do(context.Background(), http.MethodGet, "/", "", nil)
	require.NoError(t, err)

	_ = resp.Body.Close()
}

func TestDoTimesOut(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		time.Sleep(200 * time.Millisecond)
	}))
	defer server.Close()

	_, err := New(server.URL, 50*time.Millisecond).Do(t.Context(), http.MethodGet, "/", "", nil)
	assert.Error(t, err)
}

func TestDecodeData(t *testing.T) {
	var out struct {
		ID string `json:"id"`
	}

	require.NoError(t, DecodeData(strings.NewReader(`{"success":true,"data":{"id":"42"}}`), &out))
	assert.Equal(t, "42", out.ID)

	assert.Error(t, DecodeData(strings.NewReader(`not json`), &out))
}

func TestWithHeaderIsSentButNotOffHost(t *testing.T) {
	other := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.Header.Get("X-Api-Key"), "another host never sees the key")
	}))
	defer other.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "secret", r.Header.Get("X-Api-Key"))

		switch r.URL.Path {
		case "/same":
			http.Redirect(w, r, "/kept", http.StatusFound)
		case "/away":
			http.Redirect(w, r, other.URL, http.StatusFound)
		}
	}))
	defer server.Close()

	client := New(server.URL, time.Second, WithHeader("X-Api-Key", "secret"))

	for _, path := range []string{"/same", "/away"} {
		resp, err := client.Do(t.Context(), http.MethodGet, path, "", nil)
		require.NoError(t, err)

		_ = resp.Body.Close()
	}
}

func TestRedirectLoopStops(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/loop", http.StatusFound)
	}))
	defer server.Close()

	_, err := New(server.URL, time.Second).Do(t.Context(), http.MethodGet, "/loop", "", nil) //nolint:bodyclose // errors before a body exists
	require.ErrorContains(t, err, "stopped after 10 redirects")
}
