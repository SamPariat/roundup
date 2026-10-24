package users

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/SamPariat/roundup/services/venue-service/internal/domain"
	sharedmw "github.com/SamPariat/roundup/shared/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	squadID = "3f1c2a9e-1111-4c2b-9d7e-2b1f0a9c4e11"
	userID  = "7dc92d84-1950-43d6-b676-debacf84d830"
)

func userService(t *testing.T, status int) *httptest.Server {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/v1/squads/"+squadID, r.URL.Path)
		assert.Equal(t, userID, r.Header.Get(sharedmw.UserIDHeader))

		w.WriteHeader(status)
	}))
	t.Cleanup(srv.Close)

	return srv
}

func TestEnsureMember(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		wantErr error
	}{
		{name: "Member", status: http.StatusOK},
		{name: "NotMember", status: http.StatusNotFound, wantErr: domain.ErrSquadNotFound},
		{name: "ServerError", status: http.StatusInternalServerError, wantErr: domain.ErrUserServiceUnavailable},
		{name: "Unauthorized", status: http.StatusUnauthorized, wantErr: domain.ErrUserServiceUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewUserServiceClient(userService(t, tt.status).URL).EnsureMember(t.Context(), squadID, userID)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)

				return
			}

			require.NoError(t, err)
		})
	}
}

func TestEnsureMemberUnreachable(t *testing.T) {
	err := NewUserServiceClient("http://127.0.0.1:1").EnsureMember(t.Context(), squadID, userID)

	assert.ErrorIs(t, err, domain.ErrUserServiceUnavailable)
}

func TestEnsureMemberTimesOut(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	client := NewUserServiceClient(srv.URL)
	client.http.Timeout = 50 * time.Millisecond

	assert.ErrorIs(t, client.EnsureMember(t.Context(), squadID, userID), domain.ErrUserServiceUnavailable)
}

func TestEnsureMemberEscapesSquadID(t *testing.T) {
	var got string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.EscapedPath()

		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	_ = NewUserServiceClient(srv.URL).EnsureMember(t.Context(), "../users/me", userID)

	assert.Equal(t, "/api/v1/squads/..%2Fusers%2Fme", got, "a squad ID must not be able to change the path")
}

func TestEnsureMemberForwardsRequestID(t *testing.T) {
	var got string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get(sharedmw.RequestIDHeader)

		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	ctx := sharedmw.WithRequestID(t.Context(), "fav-req-1")
	require.NoError(t, NewUserServiceClient(srv.URL).EnsureMember(ctx, squadID, userID))

	assert.Equal(t, "fav-req-1", got)
}
