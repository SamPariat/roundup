// Package postgrestest gives integration tests a real, freshly migrated
// Postgres database. Each call to New creates its own database, so tests never
// see each other's rows and can run in any order.
//
// Tests are skipped unless TEST_POSTGRES_DSN points at a Postgres server whose
// user may create databases, e.g. one of the docker-compose containers:
//
//	TEST_POSTGRES_DSN="postgres://user_user:user_pass@localhost:5434/users?sslmode=disable" go test ./...
package postgrestest

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// EnvDSN names the environment variable holding the server to create test
// databases on. Its database part only needs to exist; it is never modified.
const EnvDSN = "TEST_POSTGRES_DSN"

// New creates a uniquely named database, applies the goose migrations in
// migrationsDir (relative to the test's package directory), and returns a pool
// connected to it. The pool is closed and the database dropped when t ends.
// The test is skipped if EnvDSN is unset.
func New(t *testing.T, migrationsDir string) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv(EnvDSN)
	if dsn == "" {
		t.Skipf("%s not set; skipping Postgres integration test", EnvDSN)
	}

	ctx := context.Background()

	admin, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("connect to %s: %v", EnvDSN, err)
	}

	// Identifiers can't be query parameters, so the name is built from a UUID
	// (hex only) rather than anything caller-supplied.
	name := "test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := admin.Exec(ctx, "CREATE DATABASE "+name); err != nil {
		_ = admin.Close(ctx)
		t.Fatalf("create test database: %v", err)
	}

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("parse %s: %v", EnvDSN, err)
	}
	cfg.ConnConfig.Database = name

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("connect to test database: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
		// FORCE ends any connection a failed test left open.
		if _, err := admin.Exec(ctx, "DROP DATABASE IF EXISTS "+name+" WITH (FORCE)"); err != nil {
			t.Errorf("drop test database %s: %v", name, err)
		}
		_ = admin.Close(ctx)
	})

	migrate(t, pool, migrationsDir)

	return pool
}

// migrate applies every up migration in dir using goose's Provider API, which
// keeps no global state and is safe for parallel tests.
func migrate(t *testing.T, pool *pgxpool.Pool, dir string) {
	t.Helper()

	db := stdlib.OpenDBFromPool(pool)
	defer func() { _ = db.Close() }()

	provider, err := goose.NewProvider(goose.DialectPostgres, db, os.DirFS(dir))
	if err != nil {
		t.Fatalf("load migrations from %s: %v", dir, err)
	}

	if _, err := provider.Up(context.Background()); err != nil {
		t.Fatalf("apply migrations from %s: %v", dir, err)
	}
}
