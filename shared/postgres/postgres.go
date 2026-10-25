// Package postgres holds the Postgres plumbing every service repeats: opening
// a connection pool and recognising the driver errors that repositories map
// to domain errors. Queries are NOT shared — each service generates its own
// with sqlc from its own schema.
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/SamPariat/roundup/shared/domainerr"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// uniqueViolation is the Postgres error code for unique constraint violations.
const uniqueViolation = "23505"

// Connect opens a connection pool and pings the database once, so a wrong DSN
// or unreachable database fails at startup instead of on the first request.
// pgxpool.New alone connects lazily and would not catch either.
func Connect(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("parse postgres config: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()

		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return pool, nil
}

// IsUniqueViolation reports whether err (or anything it wraps) is a Postgres
// unique constraint violation.
func IsUniqueViolation(err error) bool {
	pgErr, ok := errors.AsType[*pgconn.PgError](err)

	return ok && pgErr.Code == uniqueViolation
}

// IsNoRows reports whether err (or anything it wraps) means a single-row
// query matched nothing.
func IsNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

// NoRowsAs returns err as sentinel (with detail) when it means a single-row
// query matched nothing, and err unchanged otherwise, so repositories map
// "no rows" to their own not-found error in one line.
func NoRowsAs(err, sentinel error, detail string) error {
	if IsNoRows(err) {
		return domainerr.New(sentinel, detail)
	}

	return err
}

// InTx runs fn in a transaction, committing only if fn returns nil; any error
// rolls back. sqlc's Queries type differs per service, so fn gets the raw
// pgx.Tx and calls queries.WithTx(tx) itself.
func InTx(ctx context.Context, pool *pgxpool.Pool, fn func(tx pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }() // no-op after a successful commit

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
