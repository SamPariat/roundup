package postgres

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/SamPariat/roundup/shared/postgres/postgrestest"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsUniqueViolation(t *testing.T) {
	unique := &pgconn.PgError{Code: uniqueViolation}
	foreignKey := &pgconn.PgError{Code: "23503"}

	assert.True(t, IsUniqueViolation(unique))
	assert.True(t, IsUniqueViolation(fmt.Errorf("insert user: %w", unique)))
	assert.False(t, IsUniqueViolation(foreignKey))
	assert.False(t, IsUniqueViolation(errors.New("boom")))
	assert.False(t, IsUniqueViolation(nil))
}

func TestIsNoRows(t *testing.T) {
	assert.True(t, IsNoRows(pgx.ErrNoRows))
	assert.True(t, IsNoRows(fmt.Errorf("get user: %w", pgx.ErrNoRows)))
	assert.False(t, IsNoRows(errors.New("boom")))
	assert.False(t, IsNoRows(nil))
}

func TestConnectRejectsInvalidDSN(t *testing.T) {
	_, err := Connect(t.Context(), "not a dsn ::")

	require.Error(t, err)
}

func TestConnectFailsWhenDatabaseUnreachable(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()

	// Port 1 is never a Postgres server, so the ping must fail.
	_, err := Connect(ctx, "postgres://user:pass@127.0.0.1:1/db?sslmode=disable")

	require.Error(t, err)
}

func TestInTx(t *testing.T) {
	pool := postgrestest.New(t, "postgrestest/testdata/migrations")
	ctx := t.Context()

	insert := func(name string, fnErr error) error {
		return InTx(ctx, pool, func(tx pgx.Tx) error {
			if _, err := tx.Exec(ctx, "INSERT INTO widgets (name) VALUES ($1)", name); err != nil {
				return err
			}

			return fnErr
		})
	}

	require.NoError(t, insert("kept", nil))

	errBoom := errors.New("boom")
	require.ErrorIs(t, insert("rolled-back", errBoom), errBoom)

	rows, err := pool.Query(ctx, "SELECT name FROM widgets")
	require.NoError(t, err)

	names, err := pgx.CollectRows(rows, pgx.RowTo[string])
	require.NoError(t, err)

	assert.Equal(t, []string{"kept"}, names, "a failed fn must roll its writes back")
}
