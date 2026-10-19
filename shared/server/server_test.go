package server

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestHealthIgnoresDependencies(t *testing.T) {
	app := fiber.New()
	Health(app, zap.NewNop(), map[string]Check{
		"postgres": func(context.Context) error { return errors.New("down") },
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/health", nil))
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode, "liveness must not depend on the database")
}

func TestReady(t *testing.T) {
	tests := []struct {
		name       string
		checkErr   error
		wantStatus int
		wantBody   string
	}{
		{"AllChecksPass", nil, http.StatusOK, "ready"},
		{"CheckFailsHidesError", errors.New("dial tcp 10.0.0.7:5432: refused"), http.StatusServiceUnavailable, "postgres not ready"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			Health(app, zap.NewNop(), map[string]Check{
				"postgres": func(context.Context) error { return tt.checkErr },
			})

			resp, err := app.Test(httptest.NewRequest("GET", "/ready", nil))
			require.NoError(t, err)

			body, _ := io.ReadAll(resp.Body)

			assert.Equal(t, tt.wantStatus, resp.StatusCode)
			assert.Contains(t, string(body), tt.wantBody)
		})
	}
}

func TestRunDrainsInFlightRequestOnSIGTERM(t *testing.T) {
	port := freePort(t)
	started := make(chan struct{})

	app := fiber.New()
	app.Get("/slow", func(c fiber.Ctx) error {
		close(started)
		time.Sleep(300 * time.Millisecond)

		return c.SendString("finished")
	})

	runErr := make(chan error, 1)
	go func() { runErr <- Run(app, port, zap.NewNop()) }()

	waitForPort(t, port)

	type result struct {
		body string
		err  error
	}

	got := make(chan result, 1)

	go func() {
		resp, err := http.Get("http://127.0.0.1:" + port + "/slow")
		if err != nil {
			got <- result{err: err}

			return
		}
		defer func() { _ = resp.Body.Close() }()

		b, err := io.ReadAll(resp.Body)
		got <- result{string(b), err}
	}()

	<-started
	require.NoError(t, syscall.Kill(syscall.Getpid(), syscall.SIGTERM))

	r := <-got
	require.NoError(t, r.err)
	assert.Equal(t, "finished", r.body, "a request already running must complete")

	select {
	case err := <-runErr:
		require.NoError(t, err)
	case <-time.After(ShutdownTimeout + time.Second):
		t.Fatal("Run did not return after SIGTERM")
	}
}

func freePort(t *testing.T) string {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	defer func() { _ = l.Close() }()

	return strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
}

func waitForPort(t *testing.T, port string) {
	t.Helper()
	require.Eventually(t, func() bool {
		conn, err := net.Dial("tcp", "127.0.0.1:"+port)
		if err == nil {
			_ = conn.Close()
		}

		return err == nil
	}, 2*time.Second, 10*time.Millisecond)
}
