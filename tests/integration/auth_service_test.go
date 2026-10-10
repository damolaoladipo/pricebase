package integration

import (
	"context"
	"testing"

	"github.com/damolaoladipo/pricebase/internal/config/seeds"
	"github.com/damolaoladipo/pricebase/internal/factory"
	"github.com/damolaoladipo/pricebase/internal/repository"
)

func TestExtractPermissionsLive(t *testing.T) {
	db := testDB(t)
	truncateData(t, db)
	ctx := context.Background()

	if err := seeds.SeedData(ctx, db); err != nil {
		t.Fatalf("seed: %v", err)
	}

	factory.Factories(db)

	roleRepo := repository.NewRoleRepository(db)
	role, err := roleRepo.FindByName(ctx, "superadmin")
	if err != nil {
		t.Fatalf("find superadmin role: %v", err)
	}

	perms, err := factory.AuthService.ExtractPermissions(ctx, role.ID)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if len(perms) == 0 {
		t.Fatal("expected permission groups for superadmin")
	}

	for _, p := range perms {
		if p.Group == "" {
			t.Error("permission group must not be empty")
		}
		if len(p.Actions) == 0 {
			t.Errorf("group %q should have actions", p.Group)
		}
	}

	// Missing role mirrors the TS guard: empty result, no error.
	empty, err := factory.AuthService.ExtractPermissions(ctx, "000000000000000000000000")
	if err != nil {
		t.Fatalf("extract missing: %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("expected empty result for missing role, got %v", empty)
	}
}
