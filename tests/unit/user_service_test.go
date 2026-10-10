package unit

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/damolaoladipo/pricebase/internal/models"
	"github.com/damolaoladipo/pricebase/internal/services"
	"github.com/damolaoladipo/pricebase/internal/utils"
)

func testUserService() *services.UserService {
	return services.NewUserService(nil, nil, nil)
}

func testAuthService() *services.AuthService {
	return services.NewAuthService(nil, nil, nil, nil)
}

func TestValidateUserDefaults(t *testing.T) {
	svc := testUserService()
	// A zero User carries no enum values, so validation skips everything.
	user := models.User{}

	if err := svc.ValidateUser(&user); err != nil {
		t.Errorf("zero user should validate: %v", err)
	}

	user.UserType = string(utils.UserTypeUser)
	user.OTPType = string(utils.OTPTypeGeneral)

	if err := svc.ValidateUser(&user); err != nil {
		t.Errorf("user with valid enums should validate: %v", err)
	}
}

func TestValidateUserRejects(t *testing.T) {
	svc := testUserService()
	user := models.User{}
	user.UserType = "bogus"

	if err := svc.ValidateUser(&user); err == nil {
		t.Error("expected error for invalid user type")
	}

	user = models.User{}
	user.OTPType = "bogus"

	if err := svc.ValidateUser(&user); err == nil {
		t.Error("expected error for invalid otp type")
	}
}

func TestUserPasswordHashing(t *testing.T) {
	svc := testUserService()
	user := models.User{}
	user.Password = "MyS3cret!Pass"

	if err := svc.HashPassword(&user); err != nil {
		t.Fatalf("hash: %v", err)
	}
	if user.Password == "MyS3cret!Pass" {
		t.Fatal("expected password to be hashed")
	}
	if !svc.MatchPassword(&user, "MyS3cret!Pass") {
		t.Error("expected password to match")
	}
	if svc.MatchPassword(&user, "wrong") {
		t.Error("expected wrong password not to match")
	}

	empty := models.User{}
	if svc.MatchPassword(&empty, "anything") {
		t.Error("expected empty password not to match")
	}
}

func TestUserHasRole(t *testing.T) {
	svc := testUserService()
	user := models.User{}
	user.RoleIDs = []string{"role-a", "role-b"}

	if !svc.HasRole(&user, "role-a") {
		t.Error("expected HasRole to find role-a")
	}
	if svc.HasRole(&user, "role-c") {
		t.Error("expected HasRole to miss role-c")
	}
}

func TestCheckLockedStatus(t *testing.T) {
	svc := testUserService()
	user := models.User{}

	locked, err := svc.CheckLockedStatus(context.Background(), &user)
	if err != nil || locked {
		t.Errorf("expected unlocked by default: locked=%v err=%v", locked, err)
	}

	user.IsLocked = true
	locked, err = svc.CheckLockedStatus(context.Background(), &user)
	if err != nil || locked {
		t.Errorf("expected lock without expiry to read unlocked: locked=%v err=%v", locked, err)
	}

	future := time.Now().Add(time.Hour)
	user.LockedUntil = &future
	locked, err = svc.CheckLockedStatus(context.Background(), &user)
	if err != nil || !locked {
		t.Errorf("expected locked with live expiry: locked=%v err=%v", locked, err)
	}
}

func TestGetSignedJwtToken(t *testing.T) {
	svc := testAuthService()
	user := models.User{}
	user.ID = "abc123"
	user.Email = "super@pricebase.so"
	user.RoleIDs = []string{"role-a"}

	token, err := svc.GetSignedJwtToken(&user, "test-secret", "1h")
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if len(strings.Split(token, ".")) != 3 {
		t.Errorf("expected 3-part JWT, got %q", token)
	}

	if _, err := svc.GetSignedJwtToken(&user, "", "1h"); err == nil {
		t.Error("expected error for empty secret")
	}
	if _, err := svc.GetSignedJwtToken(&user, "s", "forever"); err == nil {
		t.Error("expected error for bad expiry")
	}
}
