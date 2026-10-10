// Package integration holds database-backed tests.
//
// They need a reachable Postgres. Connection details come from POSTGRES_URL,
// defaulting to a local pricebase_test database. Without a database the
// tests skip instead of failing, so plain `go test ./...` stays green:
//
//	POSTGRES_URL="postgres://pro@/pricebase_test?host=/tmp/pg&port=55433&sslmode=disable" go test ./tests/integration/ -v
//
// Logs: the logger resolves LOGS_PATH at import relative to each test
// package's working directory, defaulting to internal/_data/logs. To collect
// every suite's logs in the shared tests/logs folder instead, run with:
//
//	LOGS_PATH=../logs go test ./tests/... -v
package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/damolaoladipo/pricebase/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// testDB connects, runs migrations, and returns a pool. It skips the test
// when Postgres is unreachable.
func testDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	url := os.Getenv("POSTGRES_URL")
	if url == "" {
		url = "postgres://localhost:5432/pricebase_test?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := config.ConnectDB(url)
	if err != nil {
		t.Skipf("postgres unreachable (%v), skipping integration test", err)
	}

	t.Cleanup(func() { db.Close() })

	if err := db.Ping(ctx); err != nil {
		t.Skipf("postgres ping failed (%v), skipping integration test", err)
	}

	if err := config.RunMigrations(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	return db
}

// truncateData clears seeded tables in foreign-key-safe order.
// The migrations tracking table is left alone.
func truncateData(t *testing.T, db *pgxpool.Pool) {
	t.Helper()

	_, err := db.Exec(context.Background(), `
		TRUNCATE user_roles, permissions, roles, users, devices, notifications, settings, role_teams;
	`)
	if err != nil {
		t.Fatalf("truncate: %v", err)
	}
}
