package logger

import (
	"bufio"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name  string
		level string
		want  zapcore.Level
	}{
		{name: "debug", level: "debug", want: zapcore.DebugLevel},
		{name: "info", level: "info", want: zapcore.InfoLevel},
		{name: "warn", level: "warn", want: zapcore.WarnLevel},
		{name: "error", level: "error", want: zapcore.ErrorLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := New("test-service", tt.level)
			defer func() { _ = log.Sync() }()

			assert.True(t, log.Core().Enabled(tt.want))
		})
	}
}

func TestNewInvalidLevel(t *testing.T) {
	assert.Panics(t, func() { New("test-service", "not-a-level") })
}

func TestNewTagsOutputWithServiceAndLevel(t *testing.T) {
	line := captureStdout(t, func() {
		log := New("gateway", "info")
		log.Info("starting gateway")
		_ = log.Sync()
	})

	assert.Contains(t, line, "[gateway]")
	assert.Contains(t, line, "[INFO]")
	assert.Contains(t, line, "starting gateway")
}

// captureStdout redirects os.Stdout for the duration of fn and returns
// whatever was written to it.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	original := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	fn()

	require.NoError(t, w.Close())
	os.Stdout = original

	var buf strings.Builder
	_, err = io.Copy(&buf, bufio.NewReader(r))
	require.NoError(t, err)

	return buf.String()
}
