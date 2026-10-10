package integration

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/damolaoladipo/pricebase/internal/config/seeds"
	"github.com/damolaoladipo/pricebase/internal/factory"
	"github.com/damolaoladipo/pricebase/internal/models"
	"github.com/damolaoladipo/pricebase/internal/repository"
	"github.com/jackc/pgx/v5"
)

func TestRoleRepositoryRoundTrip(t *testing.T) {
	db := testDB(t)
	truncateData(t, db)
	ctx := context.Background()
	repo := repository.NewRoleRepository(db)

	created, err := repo.CreateRole(ctx, models.Role{Name: "test-role", Description: "a test role"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.ID == "" {
		t.Error("expected generated id")
	}

	found, err := repo.FindByName(ctx, "test-role")
	if err != nil {
		t.Fatalf("find by name: %v", err)
	}
	if found.ID != created.ID {
		t.Errorf("id mismatch: %q vs %q", found.ID, created.ID)
	}

	// Re-creating the same name is a no-op returning the existing row.
	again, err := repo.CreateRole(ctx, models.Role{Name: "test-role", Description: "changed"})
	if err != nil {
		t.Fatalf("re-create: %v", err)
	}
	if again.ID != created.ID {
		t.Errorf("expected existing row, got %q", again.ID)
	}

	if _, err := repo.FindByName(ctx, "missing"); !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("expected ErrNoRows, got %v", err)
	}
}

func TestPermissionRepositoryRoundTrip(t *testing.T) {
	db := testDB(t)
	truncateData(t, db)
	ctx := context.Background()
	repo := repository.NewPermissionRepository(db)

	created, err := repo.CreatePermission(ctx, models.Permission{
		Name:  "Test Permissions",
		Code:  "PRMTEST01",
		Group: "testgroup",
		Slug:  "testgroup-permission",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.ID == "" {
		t.Error("expected generated id")
	}

	found, err := repo.FindByGroup(ctx, "testgroup")
	if err != nil {
		t.Fatalf("find by group: %v", err)
	}
	if found.ID != created.ID {
		t.Errorf("id mismatch: %q vs %q", found.ID, created.ID)
	}
}

func TestLinkPermissionIdempotent(t *testing.T) {
	db := testDB(t)
	truncateData(t, db)
	ctx := context.Background()
	roleRepo := repository.NewRoleRepository(db)
	permRepo := repository.NewPermissionRepository(db)

	role, err := roleRepo.CreateRole(ctx, models.Role{Name: "link-role"})
	if err != nil {
		t.Fatalf("create role: %v", err)
	}
	perm, err := permRepo.CreatePermission(ctx, models.Permission{Name: "P", Code: "PRMLINK01", Group: "linkgroup", Slug: "s"})
	if err != nil {
		t.Fatalf("create permission: %v", err)
	}

	if n, err := roleRepo.LinkPermission(ctx, role.ID, perm.ID); err != nil || n != 1 {
		t.Fatalf("first link: n=%d err=%v", n, err)
	}
	if n, err := roleRepo.LinkPermission(ctx, role.ID, perm.ID); err != nil || n != 0 {
		t.Fatalf("second link should be no-op: n=%d err=%v", n, err)
	}
}

func TestUserPasswordRoundTrip(t *testing.T) {
	db := testDB(t)
	truncateData(t, db)
	ctx := context.Background()

	if err := seeds.SeedData(ctx, db); err != nil {
		t.Fatalf("seed: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	factory.Factories(db)

	var email string
	if err := db.QueryRow(ctx, "SELECT email FROM users LIMIT 1").Scan(&email); err != nil {
		t.Fatalf("first email: %v", err)
	}
	user, err := userRepo.FindUserByEmail(ctx, email)
	if err != nil {
		t.Fatalf("find user: %v", err)
	}

	if factory.UserService.MatchPassword(&user, "MyS3cret!Pass") {
		t.Fatal("plaintext seed password must not match bcrypt")
	}

	user.Password = "MyS3cret!Pass"
	if err := factory.UserService.HashPassword(&user); err != nil {
		t.Fatalf("hash: %v", err)
	}
	if !factory.UserService.MatchPassword(&user, "MyS3cret!Pass") {
		t.Error("expected hashed password to match")
	}
}

func TestUserTokenPersistence(t *testing.T) {
	db := testDB(t)
	truncateData(t, db)
	ctx := context.Background()

	if err := seeds.SeedData(ctx, db); err != nil {
		t.Fatalf("seed: %v", err)
	}

	userRepo := repository.NewUserRepository(db)

	var email string
	if err := db.QueryRow(ctx, "SELECT email FROM users LIMIT 1").Scan(&email); err != nil {
		t.Fatalf("first email: %v", err)
	}
	user, err := userRepo.FindUserByEmail(ctx, email)
	if err != nil {
		t.Fatalf("find user: %v", err)
	}

	factory.Factories(db)

	plain, err := factory.AuthService.GetResetPasswordToken(ctx, user.ID)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	var stored string
	if err := db.QueryRow(ctx, "SELECT reset_password_token FROM users WHERE id = $1", user.ID).Scan(&stored); err != nil {
		t.Fatalf("read back: %v", err)
	}
	sum := sha256.Sum256([]byte(plain))
	if stored != hex.EncodeToString(sum[:]) {
		t.Error("stored hash does not match generated hash")
	}
}
