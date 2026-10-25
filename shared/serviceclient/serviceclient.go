// Package serviceclient is the HTTP plumbing every service uses to call
// another service directly: one of ours (not through the gateway) or an
// outside API such as Google's. It sends JSON, sends the request as a user,
// passes the request ID on so both sides log the same ID, and decodes our
// response envelope. Each client keeps its own routes, statuses and sentinel
// errors on top.
package serviceclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"time"

	sharedmw "github.com/SamPariat/roundup/shared/middleware"
	"github.com/SamPariat/roundup/shared/response"
)

// Client calls one service at a fixed base URL.
type Client struct {
	baseURL string
	http    *http.Client
	header  http.Header
}

// Option configures a Client.
type Option func(*Client)

// WithHeader sends key: value on every request, e.g. an outside API's key.
// It is dropped on a redirect to another host, so a secret never follows the
// redirect off the service it was meant for.
func WithHeader(key, value string) Option {
	return func(c *Client) { c.header.Set(key, value) }
}

// New returns a Client for the service at baseURL, e.g.
// "http://localhost:3002". timeout bounds each call, so a slow service fails
// the request fast instead of hanging it.
func New(baseURL string, timeout time.Duration, opts ...Option) *Client {
	c := &Client{baseURL: baseURL, header: http.Header{}}
	for _, opt := range opts {
		opt(c)
	}

	c.http = &http.Client{Timeout: timeout, CheckRedirect: c.dropHeaderOffHost}

	return c
}

// Do sends method to path. userID, if not empty, is sent as X-User-ID so the
// other service applies its own rules for that user; body, if not nil, is
// sent as JSON. The returned error is a transport failure (unreachable,
// timed out); callers wrap it in their own "unavailable" sentinel. The
// caller closes the response body.
func (c *Client) Do(ctx context.Context, method, path, userID string, body any) (*http.Response, error) {
	var reader io.Reader

	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("encode request: %w", err)
		}

		reader = bytes.NewReader(encoded)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return nil, err
	}

	maps.Copy(req.Header, c.header)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if userID != "" {
		req.Header.Set(sharedmw.UserIDHeader, userID)
	}

	if id := sharedmw.RequestIDFromContext(ctx); id != "" {
		req.Header.Set(sharedmw.RequestIDHeader, id)
	}

	return c.http.Do(req)
}

// DecodeData decodes a {"success": true, "data": ...} body's data into out.
func DecodeData(body io.Reader, out any) error {
	if err := json.NewDecoder(body).Decode(&response.Response[any]{Data: out}); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

// maxRedirects is Go's own limit, which a custom CheckRedirect replaces.
const maxRedirects = 10

// dropHeaderOffHost removes the WithHeader headers when a redirect leaves the
// original host. Go itself only drops Authorization and Cookie.
func (c *Client) dropHeaderOffHost(req *http.Request, via []*http.Request) error {
	if len(via) >= maxRedirects {
		return fmt.Errorf("stopped after %d redirects", maxRedirects)
	}

	if req.URL.Host != via[0].URL.Host {
		for key := range c.header {
			req.Header.Del(key)
		}
	}

	return nil
}
