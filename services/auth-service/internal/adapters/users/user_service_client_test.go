package users

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/SamPariat/roundup/services/auth-service/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const userID = "7dc92d84-1950-43d6-b676-debacf84d830"

func userService(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/users", r.URL.Path)
		var req map[string]string
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		assert.Equal(t, "+919876543210", req["phone"])

		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestEnsureUser(t *testing.T) {
	ok := `{"success":true,"data":{"id":"` + userID + `"}}`

	tests := []struct {
		name        string
		status      int
		body        string
		wantCreated bool
		wantErr     bool
	}{
		{name: "NewUser", status: http.StatusCreated, body: ok, wantCreated: true},
		{name: "ExistingUser", status: http.StatusOK, body: ok, wantCreated: false},
		{name: "ServerError", status: http.StatusInternalServerError, body: `{"success":false}`, wantErr: true},
		{name: "BadRequest", status: http.StatusBadRequest, body: `{"success":false}`, wantErr: true},
		{name: "NoID", status: http.StatusOK, body: `{"success":true,"data":{}}`, wantErr: true},
		{name: "NotJSON", status: http.StatusOK, body: `<html>`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewUserServiceClient(userService(t, tt.status, tt.body).URL)

			id, created, err := client.EnsureUser(context.Background(), "+919876543210")

			if tt.wantErr {
				assert.ErrorIs(t, err, domain.ErrUserServiceUnavailable)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, userID, id)
			assert.Equal(t, tt.wantCreated, created)
		})
	}
}

func TestEnsureUserUnreachable(t *testing.T) {
	client := NewUserServiceClient("http://127.0.0.1:1")

	_, _, err := client.EnsureUser(context.Background(), "+919876543210")

	assert.ErrorIs(t, err, domain.ErrUserServiceUnavailable)
}

func TestEnsureUserTimesOut(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		time.Sleep(200 * time.Millisecond)
	}))
	t.Cleanup(srv.Close)
	client := NewUserServiceClient(srv.URL)
	client.http.Timeout = 50 * time.Millisecond

	_, _, err := client.EnsureUser(context.Background(), "+919876543210")

	assert.ErrorIs(t, err, domain.ErrUserServiceUnavailable)
}
