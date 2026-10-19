package postgrestest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAppliesMigrations(t *testing.T) {
	pool := New(t, "testdata/migrations")

	_, err := pool.Exec(t.Context(), "INSERT INTO widgets (name) VALUES ('a')")
	require.NoError(t, err)

	var count int
	require.NoError(t, pool.QueryRow(t.Context(), "SELECT count(*) FROM widgets").Scan(&count))
	assert.Equal(t, 1, count)
}

func TestNewIsolatesDatabases(t *testing.T) {
	first := New(t, "testdata/migrations")
	second := New(t, "testdata/migrations")

	_, err := first.Exec(t.Context(), "INSERT INTO widgets (name) VALUES ('only-in-first')")
	require.NoError(t, err)

	var count int
	require.NoError(t, second.QueryRow(t.Context(), "SELECT count(*) FROM widgets").Scan(&count))
	assert.Equal(t, 0, count, "each New call must get its own database")
}
