package integration

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/damolaoladipo/pricebase/internal/config/seeds"
	"github.com/damolaoladipo/pricebase/internal/dtos"
	"github.com/damolaoladipo/pricebase/internal/factory"
	"github.com/damolaoladipo/pricebase/internal/models"
	"github.com/damolaoladipo/pricebase/internal/repository"
	"github.com/damolaoladipo/pricebase/internal/utils"
)

func createFlowUser(t *testing.T, ctx context.Context, repo *repository.UserRepository, email, userType string) models.User {
	t.Helper()
	created, err := repo.CreateUser(ctx, models.User{
		Email:       email,
		UserType:    userType,
		FirstName:   "Flow",
		LastName:    "Test",
		PhoneNumber: "08030000001",
		IsActive:    true,
		IsActivated: true,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	return created
}

func TestUserAccountFlows(t *testing.T) {
	db := testDB(t)
	truncateData(t, db)
	ctx := context.Background()
	factory.Factories(db)
	userRepo := repository.NewUserRepository(db)

	user := createFlowUser(t, ctx, userRepo, "flows@example.com", string(utils.UserTypeUser))

	t.Run("phone exists", func(t *testing.T) {
		ok, err := factory.UserService.PhoneExists(ctx, "08030000001")
		if err != nil || !ok {
			t.Errorf("expected phone to exist: ok=%v err=%v", ok, err)
		}
		ok, err = factory.UserService.PhoneExists(ctx, "09999999999")
		if err != nil || ok {
			t.Errorf("expected unknown phone to miss: ok=%v err=%v", ok, err)
		}
	})

	t.Run("update user type", func(t *testing.T) {
		if err := factory.UserService.UpdateUserType(ctx, &user, string(utils.UserTypeOrganization)); err != nil {
			t.Fatalf("update type: %v", err)
		}
		reloaded, err := userRepo.FindUserByID(ctx, user.ID)
		if err != nil {
			t.Fatalf("reload: %v", err)
		}
		if reloaded.UserType != string(utils.UserTypeOrganization) {
			t.Errorf("expected persisted type, got %q", reloaded.UserType)
		}
		user = reloaded
	})

	t.Run("last login stamp", func(t *testing.T) {
		if err := factory.UserService.UpdateLastLogin(ctx, &user); err != nil {
			t.Fatalf("last login: %v", err)
		}
		if user.Login.Last == "" || user.Login.Method != string(utils.LoginMethodEmail) {
			t.Errorf("expected stamped login, got %+v", user.Login)
		}
		reloaded, err := userRepo.FindUserByID(ctx, user.ID)
		if err != nil {
			t.Fatalf("reload: %v", err)
		}
		if reloaded.Login.Last == "" {
			t.Error("expected persisted login stamp")
		}
	})

	t.Run("deactivate then activate", func(t *testing.T) {
		if err := factory.UserService.DeactivateAccount(ctx, &user); err != nil {
			t.Fatalf("deactivate: %v", err)
		}
		if user.IsActive || user.IsActivated || !user.IsLocked {
			t.Errorf("expected disabled+locked, got %+v", user)
		}
		if err := factory.UserService.ActivateAccount(ctx, &user); err != nil {
			t.Fatalf("activate: %v", err)
		}
		reloaded, err := userRepo.FindUserByID(ctx, user.ID)
		if err != nil {
			t.Fatalf("reload: %v", err)
		}
		if !reloaded.IsActive || !reloaded.IsActivated || reloaded.IsLocked || reloaded.LoginLimit != 0 {
			t.Errorf("expected enabled+unlocked+reset, got %+v", reloaded)
		}
		user = reloaded
	})

	t.Run("expired lock auto-unlocks", func(t *testing.T) {
		past := time.Now().Add(-time.Minute)
		if _, err := db.Exec(ctx, "UPDATE users SET is_locked = true, locked_until = $1, login_limit = 5 WHERE id = $2", past, user.ID); err != nil {
			t.Fatalf("force lock: %v", err)
		}
		locked, err := factory.UserService.CheckLockedStatus(ctx, &user)
		if err != nil {
			t.Fatalf("status: %v", err)
		}
		_ = locked
		stale, err := userRepo.FindUserByID(ctx, user.ID)
		if err != nil {
			t.Fatalf("reload: %v", err)
		}
		locked, err = factory.UserService.CheckLockedStatus(ctx, &stale)
		if err != nil || locked {
			t.Errorf("expected auto-unlock: locked=%v err=%v", locked, err)
		}
		reloaded, err := userRepo.FindUserByID(ctx, user.ID)
		if err != nil {
			t.Fatalf("reload: %v", err)
		}
		if reloaded.IsLocked || reloaded.LoginLimit != 0 {
			t.Errorf("expected cleared lock state, got locked=%v limit=%d", reloaded.IsLocked, reloaded.LoginLimit)
		}
	})
}

func TestOTPFlows(t *testing.T) {
	db := testDB(t)
	truncateData(t, db)
	ctx := context.Background()
	factory.Factories(db)
	userRepo := repository.NewUserRepository(db)

	user := createFlowUser(t, ctx, userRepo, "otp-flows@example.com", string(utils.UserTypeUser))

	code, err := factory.UserService.GenerateOTPCode(ctx, &user, string(utils.OTPTypeGeneral))
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(code) != 6 {
		t.Errorf("expected 6-digit code, got %q", code)
	}

	found, err := factory.UserService.VerifyOTPCode(ctx, code)
	if err != nil || found == nil {
		t.Errorf("expected code holder: found=%v err=%v", found, err)
	}
	found, err = factory.UserService.VerifyOTPCode(ctx, "000000")
	if err != nil || found != nil {
		t.Errorf("expected miss for wrong code: found=%v err=%v", found, err)
	}

	res, err := factory.UserService.VerifyOTP(ctx, dtos.VerifyOTPDTO{Email: user.Email, Code: code, OTPType: string(utils.OTPTypeGeneral)})
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if res.Error || res.Code != 200 {
		t.Errorf("expected success, got %+v", res)
	}

	res, _ = factory.UserService.VerifyOTP(ctx, dtos.VerifyOTPDTO{Email: user.Email, Code: code, OTPType: string(utils.OTPTypeGeneral)})
	if !res.Error || res.Code != 400 {
		t.Errorf("expected consumed code to fail, got %+v", res)
	}

	code2, err := factory.UserService.GenerateOTPCode(ctx, &user, string(utils.OTPTypeGeneral))
	if err != nil {
		t.Fatalf("generate 2: %v", err)
	}
	past := time.Now().Add(-time.Minute)
	if _, err := db.Exec(ctx, "UPDATE users SET otp_code_expire = $1 WHERE id = $2", past, user.ID); err != nil {
		t.Fatalf("expire code: %v", err)
	}
	res, _ = factory.UserService.VerifyOTP(ctx, dtos.VerifyOTPDTO{Email: user.Email, Code: code2, OTPType: string(utils.OTPTypeGeneral)})
	if !res.Error || res.Code != 400 {
		t.Errorf("expected expired code to fail, got %+v", res)
	}
}

func TestLoggedInUserAndAttachRole(t *testing.T) {
	db := testDB(t)
	truncateData(t, db)
	ctx := context.Background()

	if err := seeds.SeedData(ctx, db); err != nil {
		t.Fatalf("seed: %v", err)
	}
	factory.Factories(db)
	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)

	user := createFlowUser(t, ctx, userRepo, "attach-me@example.com", string(utils.UserTypeUser))

	mapped, err := factory.UserService.GetLoggedInUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("logged in user: %v", err)
	}
	if mapped.Email != user.Email {
		t.Errorf("mapped mismatch: %+v", mapped)
	}

	if _, err := factory.UserService.GetLoggedInUser(ctx, "missing-id"); err == nil {
		t.Error("expected error for missing user")
	} else if custom, ok := err.(*utils.ErrorResponse); !ok || custom.Status != http.StatusUnauthorized {
		t.Errorf("expected 401 error, got %#v", err)
	}

	if err := factory.UserService.AttachRole(ctx, &user, "admin"); err != nil {
		t.Fatalf("attach: %v", err)
	}
	admin, err := roleRepo.FindByName(ctx, "admin")
	if err != nil {
		t.Fatalf("find admin role: %v", err)
	}
	has, err := userRepo.HasRole(ctx, user.ID, admin.ID)
	if err != nil || !has {
		t.Errorf("expected role link: has=%v err=%v", has, err)
	}

	if err := factory.UserService.AttachRole(ctx, &user, "no-such-role"); err == nil {
		t.Error("expected error for missing role")
	} else if custom, ok := err.(*utils.ErrorResponse); !ok || custom.Status != http.StatusBadRequest {
		t.Errorf("expected 400 error, got %#v", err)
	}
}
