package integration

import (
	"context"
	"testing"

	"github.com/damolaoladipo/pricebase/internal/config/seeds"
	"github.com/jackc/pgx/v5/pgxpool"
)

func count(t *testing.T, db *pgxpool.Pool, table string) int {
	t.Helper()

	var n int
	if err := db.QueryRow(context.Background(), "SELECT COUNT(*) FROM "+table).Scan(&n); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return n
}

func TestSeedData(t *testing.T) {
	db := testDB(t)
	truncateData(t, db)
	ctx := context.Background()

	if err := seeds.SeedData(ctx, db); err != nil {
		t.Fatalf("seed: %v", err)
	}

	if got := count(t, db, "roles"); got == 0 {
		t.Error("expected roles to be seeded")
	}
	if got := count(t, db, "permissions"); got == 0 {
		t.Error("expected permissions to be seeded")
	}
	if got := count(t, db, "users"); got == 0 {
		t.Error("expected users to be seeded")
	}
}

func TestSeedDataIdempotent(t *testing.T) {
	db := testDB(t)
	truncateData(t, db)
	ctx := context.Background()

	if err := seeds.SeedData(ctx, db); err != nil {
		t.Fatalf("first seed: %v", err)
	}

	before := map[string]int{
		"roles":       count(t, db, "roles"),
		"permissions": count(t, db, "permissions"),
		"users":       count(t, db, "users"),
	}

	if err := seeds.SeedData(ctx, db); err != nil {
		t.Fatalf("second seed: %v", err)
	}

	for table, n := range before {
		if got := count(t, db, table); got != n {
			t.Errorf("table %s changed on re-seed: %d -> %d", table, n, got)
		}
	}
}
