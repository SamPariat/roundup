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

// userService fakes user-service at path, answering status with body.
func userService(t *testing.T, path string, status int, body string) *httptest.Server {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, path, r.URL.Path)
		assert.Equal(t, userID, r.Header.Get(sharedmw.UserIDHeader))

		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	return srv
}

const squadBody = `{"success":true,"data":{"id":"` + squadID + `","name":"Friday crew","members":[
	{"userID":"` + userID + `","name":"Meera","avatarURL":"https://a/m.png","role":"ADMIN"},
	{"userID":"u-2","name":"","avatarURL":"","role":"MEMBER"}]}}`

func TestMembers(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		body    string
		want    []domain.SquadMember
		wantErr error
	}{
		{name: "Member", status: http.StatusOK, body: squadBody, want: []domain.SquadMember{
			{UserID: userID, Name: "Meera", AvatarURL: "https://a/m.png"},
			{UserID: "u-2"},
		}},
		{name: "NotMember", status: http.StatusNotFound, body: `{"success":false}`, wantErr: domain.ErrSquadNotFound},
		{name: "BadBody", status: http.StatusOK, body: `not json`, wantErr: domain.ErrUserServiceUnavailable},
		{name: "ServerError", status: http.StatusInternalServerError, wantErr: domain.ErrUserServiceUnavailable},
		{name: "Unauthorized", status: http.StatusUnauthorized, wantErr: domain.ErrUserServiceUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := userService(t, "/api/v1/squads/"+squadID, tt.status, tt.body)

			got, err := NewUserServiceClient(srv.URL).Members(t.Context(), squadID, userID)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSquads(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		body    string
		want    []domain.Squad
		wantErr error
	}{
		{
			name: "Some", status: http.StatusOK, body: `{"success":true,"data":{"squads":[
			{"id":"s-1","name":"Friday crew","role":"ADMIN","memberCount":4},
			{"id":"s-2","name":"Office","role":"MEMBER","memberCount":9}]}}`,
			want: []domain.Squad{{ID: "s-1", Name: "Friday crew"}, {ID: "s-2", Name: "Office"}},
		},
		{name: "None", status: http.StatusOK, body: `{"success":true,"data":{"squads":[]}}`, want: []domain.Squad{}},
		{name: "NotFoundIsUnavailable", status: http.StatusNotFound, wantErr: domain.ErrUserServiceUnavailable},
		{name: "ServerError", status: http.StatusInternalServerError, wantErr: domain.ErrUserServiceUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := userService(t, "/api/v1/squads", tt.status, tt.body)

			got, err := NewUserServiceClient(srv.URL).Squads(t.Context(), userID)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMembersUnreachable(t *testing.T) {
	_, err := NewUserServiceClient("http://127.0.0.1:1").Members(t.Context(), squadID, userID)

	assert.ErrorIs(t, err, domain.ErrUserServiceUnavailable)
}

func TestMembersTimesOut(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	client := NewUserServiceClient(srv.URL)
	client.http.Timeout = 50 * time.Millisecond

	_, err := client.Members(t.Context(), squadID, userID)

	assert.ErrorIs(t, err, domain.ErrUserServiceUnavailable)
}

func TestMembersEscapesSquadID(t *testing.T) {
	var got string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.EscapedPath()

		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	_, _ = NewUserServiceClient(srv.URL).Members(t.Context(), "../users/me", userID)

	assert.Equal(t, "/api/v1/squads/..%2Fusers%2Fme", got, "a squad ID must not be able to change the path")
}

func TestMembersForwardsRequestID(t *testing.T) {
	var got string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get(sharedmw.RequestIDHeader)

		_, _ = w.Write([]byte(squadBody))
	}))
	t.Cleanup(srv.Close)

	ctx := sharedmw.WithRequestID(t.Context(), "fav-req-1")
	_, err := NewUserServiceClient(srv.URL).Members(ctx, squadID, userID)
	require.NoError(t, err)

	assert.Equal(t, "fav-req-1", got)
}

func TestSquadsFollowsEveryPage(t *testing.T) {
	pages := map[string]string{
		"":   `{"success":true,"data":{"squads":[{"id":"s-1","name":"A"},{"id":"s-2","name":"B"}],"nextCursor":"c2"}}`,
		"c2": `{"success":true,"data":{"squads":[{"id":"s-3","name":"C"}],"nextCursor":"c3"}}`,
		"c3": `{"success":true,"data":{"squads":[{"id":"s-4","name":"D"}],"nextCursor":null}}`,
	}

	var limits []string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limits = append(limits, r.URL.Query().Get("limit"))

		body, ok := pages[r.URL.Query().Get("cursor")]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	got, err := NewUserServiceClient(srv.URL).Squads(t.Context(), userID)

	require.NoError(t, err)
	assert.Equal(t, []domain.Squad{{ID: "s-1", Name: "A"}, {ID: "s-2", Name: "B"}, {ID: "s-3", Name: "C"}, {ID: "s-4", Name: "D"}}, got)
	assert.Equal(t, []string{"100", "100", "100"}, limits, "asks for the largest page to keep calls few")
}
