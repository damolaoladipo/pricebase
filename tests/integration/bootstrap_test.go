package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/damolaoladipo/pricebase/internal/config"
	"github.com/damolaoladipo/pricebase/internal/config/seeds"
	"github.com/damolaoladipo/pricebase/internal/models"
	"github.com/damolaoladipo/pricebase/internal/repository"
)

func testPostgresURL(t *testing.T) string {
	t.Helper()
	url := os.Getenv("POSTGRES_URL")
	if url == "" {
		url = "postgres://localhost:5432/pricebase_test?sslmode=disable"
	}
	return url
}

func TestConnectDB(t *testing.T) {
	if _, err := config.ConnectDB(""); err == nil {
		t.Error("expected error for empty database URL")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	bad, err := config.ConnectDB("postgres://127.0.0.1:1/nope?sslmode=disable")
	if err == nil {
		bad.Close()
		t.Error("expected error for unreachable host")
	}

	pool, err := config.ConnectDB(testPostgresURL(t))
	if err != nil {
		t.Skipf("postgres unreachable (%v), skipping", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		t.Errorf("ping: %v", err)
	}
}

func TestRunMigrationsIdempotent(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	if err := config.RunMigrations(db); err != nil {
		t.Fatalf("re-run migrations: %v", err)
	}

	files, err := filepath.Glob("../../migrations/*.up.sql")
	if err != nil {
		t.Fatalf("glob migrations: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("expected migration files")
	}

	var applied int
	if err := db.QueryRow(ctx, "SELECT COUNT(*) FROM migrations").Scan(&applied); err != nil {
		t.Fatalf("count applied: %v", err)
	}
	if applied != len(files) {
		t.Errorf("expected %d applied migrations, got %d", len(files), applied)
	}
}

func TestAttachSuperAdminRoleLive(t *testing.T) {
	db := testDB(t)
	truncateData(t, db)
	ctx := context.Background()
	initIntegrationEnv(t)

	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)

	created, err := userRepo.CreateUser(ctx, models.User{
		Email:         "admin@example.com",
		UserType:      "superadmin",
		FirstName:     "Super",
		LastName:      "Admin",
		Password:      "Temp123!",
		SavedPassword: "Temp123!",
		IsActive:      true,
		IsActivated:   true,
	})
	if err != nil {
		t.Fatalf("create superadmin user: %v", err)
	}

	// SeedData ends with AttachSuperAdminRole, so this exercises the full
	// chain: roles, permissions, links, then the superadmin attach.
	if err := seeds.SeedData(ctx, db); err != nil {
		t.Fatalf("seed: %v", err)
	}

	userRole, err := roleRepo.FindByName(ctx, "user")
	if err != nil {
		t.Fatalf("find user role: %v", err)
	}
	superRole, err := roleRepo.FindByName(ctx, "superadmin")
	if err != nil {
		t.Fatalf("find superadmin role: %v", err)
	}

	for _, roleID := range []string{userRole.ID, superRole.ID} {
		has, err := userRepo.HasRole(ctx, created.ID, roleID)
		if err != nil {
			t.Fatalf("has role %s: %v", roleID, err)
		}
		if !has {
			t.Errorf("expected superadmin to hold role %s", roleID)
		}
	}

	reloaded, err := userRepo.FindUserByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("reload superadmin: %v", err)
	}
	for _, roleID := range []string{userRole.ID, superRole.ID} {
		found := false
		for _, id := range reloaded.RoleIDs {
			if id == roleID {
				found = true
			}
		}
		if !found {
			t.Errorf("expected role snapshot to contain %s, got %v", roleID, reloaded.RoleIDs)
		}
	}
	if len(reloaded.RolePermissions) == 0 {
		t.Error("expected permission snapshot on superadmin")
	}

	// FindUserByID does not select api key columns, so read the secret directly.
	var secret *string
	err = db.QueryRow(ctx, "SELECT api_key->>'secret' FROM users WHERE id = $1", created.ID).Scan(&secret)
	if err != nil {
		t.Fatalf("read api key: %v", err)
	}
	if secret == nil || *secret == "" {
		t.Error("expected API key generated for superadmin")
	}

	// Re-seed must be a no-op for the attach (guarded by existing links).
	if err := seeds.SeedData(ctx, db); err != nil {
		t.Fatalf("re-seed: %v", err)
	}
}
