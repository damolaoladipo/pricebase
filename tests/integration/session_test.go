package integration

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/damolaoladipo/pricebase/internal/common/redis"
	"github.com/damolaoladipo/pricebase/internal/config/seeds"
	"github.com/damolaoladipo/pricebase/internal/dtos"
	"github.com/damolaoladipo/pricebase/internal/factory"
	"github.com/damolaoladipo/pricebase/internal/models"
	"github.com/damolaoladipo/pricebase/internal/repository"
	"github.com/damolaoladipo/pricebase/internal/utils"
)

// initIntegrationEnv loads the utils singleton for tests that sign tokens.
// Redis-backed tests need JWT_SECRET/JWT_EXPIRES/APP_ENV at minimum; the full
// required set is provided so NewENV succeeds.
func initIntegrationEnv(t *testing.T) {
	t.Helper()
	vars := map[string]string{
		"GO_ENV":            "test",
		"APP_ENV":           "test",
		"APP_PORT":          "8080",
		"API_ROUTE":         "/api",
		"APP_CHANNELS":      "email",
		"AUTH_TYPE":         "development",
		"JWT_SECRET":        "integration-test-secret",
		"JWT_EXPIRES":       "1h",
		"JWT_COOKIE_EXPIRE": "1h",
		"POSTGRES_URL":      "postgres://localhost:5432/pricebase_test?sslmode=disable",
		"POSTGRES_TEST_URL": "postgres://localhost:5432/pricebase_test?sslmode=disable",
		"CORS_ORIGINS":      "http://localhost:3000",
		"SUPERADMIN_EMAIL":  "admin@example.com",
	}
	for k, v := range vars {
		t.Setenv(k, v)
	}
	if _, err := utils.NewENV(); err != nil {
		t.Fatalf("NewENV: %v", err)
	}
}

// testRedis connects the shared wrapper to local Redis, skipping when down.
// JTIs are unique per test so no flush is needed.
func testRedis(t *testing.T) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redis.REDIS.Connect(ctx, redis.Options{
		Host:     "127.0.0.1",
		Port:     6379,
		Platform: redis.Provisioned,
	}); err != nil {
		t.Skipf("redis unreachable (%v), skipping", err)
	}
}

func TestSessionLifecycle(t *testing.T) {
	db := testDB(t)
	truncateData(t, db)
	initIntegrationEnv(t)
	testRedis(t)
	factory.Factories(db)
	ctx := context.Background()

	token, jti, err := factory.AuthService.GenerateJWT(ctx, dtos.GenerateJWTDTO{
		Email: "sess@example.com",
		ID:    "user-sess-1",
	})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if token == "" || jti == "" {
		t.Fatal("expected token and jti")
	}

	parsed, err := jwt.Parse(token, func(*jwt.Token) (any, error) {
		return []byte("integration-test-secret"), nil
	})
	if err != nil || !parsed.Valid {
		t.Fatalf("token must verify with the test secret: %v", err)
	}

	ok, err := factory.AuthService.VerifySession(ctx, dtos.VerifySessionDTO{JTI: jti, UserID: "user-sess-1"})
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if !ok {
		t.Error("expected fresh session to verify")
	}

	ok, _ = factory.AuthService.VerifySession(ctx, dtos.VerifySessionDTO{JTI: jti, UserID: "someone-else"})
	if ok {
		t.Error("expected wrong user to fail verification")
	}

	if err := factory.AuthService.RevokeSession(ctx, jti); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	ok, _ = factory.AuthService.VerifySession(ctx, dtos.VerifySessionDTO{JTI: jti, UserID: "user-sess-1"})
	if ok {
		t.Error("expected revoked session to fail verification")
	}
}

func TestRevokeAllSessions(t *testing.T) {
	db := testDB(t)
	truncateData(t, db)
	initIntegrationEnv(t)
	testRedis(t)
	factory.Factories(db)
	ctx := context.Background()

	var jtis []string
	for i := 0; i < 2; i++ {
		_, jti, err := factory.AuthService.GenerateJWT(ctx, dtos.GenerateJWTDTO{
			Email: "multi@example.com",
			ID:    "user-multi-1",
		})
		if err != nil {
			t.Fatalf("generate %d: %v", i, err)
		}
		jtis = append(jtis, jti)
	}

	if err := factory.AuthService.RevokeAllSessions(ctx, "user-multi-1"); err != nil {
		t.Fatalf("revoke all: %v", err)
	}
	for _, jti := range jtis {
		ok, _ := factory.AuthService.VerifySession(ctx, dtos.VerifySessionDTO{JTI: jti, UserID: "user-multi-1"})
		if ok {
			t.Errorf("expected jti %s to be revoked", jti)
		}
	}
}

func TestGetActivationTokenLive(t *testing.T) {
	db := testDB(t)
	truncateData(t, db)
	initIntegrationEnv(t)
	factory.Factories(db)
	ctx := context.Background()
	userRepo := repository.NewUserRepository(db)

	created, err := userRepo.CreateUser(ctx, models.User{
		Email:       "activate-me@example.com",
		UserType:    "user",
		FirstName:   "Act",
		LastName:    "Me",
		IsActive:    true,
		IsActivated: false,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	plain, err := factory.AuthService.GetActivationToken(ctx, created.ID)
	if err != nil {
		t.Fatalf("activation token: %v", err)
	}
	if plain == "" {
		t.Fatal("expected plaintext token")
	}

	stored, err := userRepo.FindUserByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("reload user: %v", err)
	}
	_ = stored

	// FindUserByID does not select token columns, so read them directly.
	var hash *string
	var expires *time.Time
	err = db.QueryRow(ctx, "SELECT activation_token, activation_token_expire FROM users WHERE id = $1", created.ID).Scan(&hash, &expires)
	if err != nil {
		t.Fatalf("read token columns: %v", err)
	}
	if hash == nil {
		t.Fatal("expected stored hash")
	}
	sum := sha256.Sum256([]byte(plain))
	if *hash != hex.EncodeToString(sum[:]) {
		t.Error("stored hash must be sha256 of the plaintext token")
	}
	if expires == nil || time.Until(*expires) > 10*time.Minute {
		t.Error("expected ~10 minute expiry")
	}

	empty, err := factory.AuthService.GetActivationToken(ctx, "missing-user-id")
	if err != nil || empty != "" {
		t.Errorf("expected empty token and nil error for missing user, got %q, %v", empty, err)
	}
}

func TestGetAuthTokenLive(t *testing.T) {
	db := testDB(t)
	truncateData(t, db)
	ctx := context.Background()

	if err := seeds.SeedData(ctx, db); err != nil {
		t.Fatalf("seed: %v", err)
	}
	initIntegrationEnv(t)
	testRedis(t)
	factory.Factories(db)
	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)

	superRole, err := roleRepo.FindByName(ctx, "superadmin")
	if err != nil {
		t.Fatalf("find superadmin role: %v", err)
	}

	created, err := userRepo.CreateUser(ctx, models.User{
		Email:       "login-me@example.com",
		UserType:    "superadmin",
		FirstName:   "Log",
		LastName:    "In",
		IsActive:    true,
		IsActivated: true,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	perms, err := factory.AuthService.ExtractPermissions(ctx, superRole.ID)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if err := userRepo.UpdateRolesPermissions(ctx, created.ID, []string{superRole.ID}, perms); err != nil {
		t.Fatalf("attach role snapshot: %v", err)
	}

	auth, err := factory.AuthService.GetAuthToken(ctx, created.ID)
	if err != nil {
		t.Fatalf("auth token: %v", err)
	}
	if auth.Token == "" {
		t.Error("expected signed token")
	}
	if auth.User.Email != "login-me@example.com" {
		t.Errorf("mapped user mismatch: %+v", auth.User)
	}
	if len(auth.User.Roles) != 1 || auth.User.Roles[0].Name != "superadmin" {
		t.Errorf("mapped roles mismatch: %+v", auth.User.Roles)
	}
	if auth.User.PasswordHash != "" {
		t.Error("mapped user must not leak the password hash")
	}
	if auth.Cookie.MaxAge != 70*24*60*60 || auth.Cookie.Path != "/" || auth.Cookie.Secure {
		t.Errorf("cookie mismatch: %+v", auth.Cookie)
	}
}

func TestLoginLockoutLive(t *testing.T) {
	db := testDB(t)
	truncateData(t, db)
	ctx := context.Background()
	factory.Factories(db)
	userRepo := repository.NewUserRepository(db)

	created, err := userRepo.CreateUser(ctx, models.User{
		Email:     "lock-me@example.com",
		UserType:  "user",
		FirstName: "Lock",
		LastName:  "Me",
		IsActive:  true,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	user, err := userRepo.FindUserByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("load user: %v", err)
	}

	for i := 1; i <= 4; i++ {
		got, err := factory.UserService.IncreaseLoginLimit(ctx, &user)
		if err != nil {
			t.Fatalf("attempt %d: %v", i, err)
		}
		if got != i {
			t.Errorf("expected limit %d, got %d", i, got)
		}
		locked, err := factory.UserService.CheckLockedStatus(ctx, &user)
		if err != nil {
			t.Fatalf("lock status %d: %v", i, err)
		}
		if locked {
			t.Errorf("expected unlocked after %d attempts", i)
		}
	}

	got, err := factory.UserService.IncreaseLoginLimit(ctx, &user)
	if err != nil {
		t.Fatalf("attempt 5: %v", err)
	}
	if got != 5 {
		t.Errorf("expected limit 5, got %d", got)
	}
	locked, err := factory.UserService.CheckLockedStatus(ctx, &user)
	if err != nil {
		t.Fatalf("lock status: %v", err)
	}
	if !locked {
		t.Fatal("expected locked after 5 attempts")
	}
	if user.LockedUntil == nil || time.Until(*user.LockedUntil) > 30*time.Minute {
		t.Error("expected ~30 minute lock expiry")
	}

	reloaded, err := userRepo.FindUserByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if !reloaded.IsLocked || reloaded.LoginLimit != 5 || reloaded.LockedUntil == nil {
		t.Errorf("expected persisted lock state, got %+v", reloaded)
	}
}

func TestResetInviteTokensLive(t *testing.T) {
	db := testDB(t)
	truncateData(t, db)
	ctx := context.Background()
	factory.Factories(db)
	userRepo := repository.NewUserRepository(db)

	created, err := userRepo.CreateUser(ctx, models.User{
		Email:     "tokens-live@example.com",
		UserType:  "user",
		FirstName: "Tok",
		LastName:  "Ens",
		IsActive:  true,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	reset, err := factory.AuthService.GetResetPasswordToken(ctx, created.ID)
	if err != nil || reset == "" {
		t.Fatalf("reset token: %v", err)
	}
	invite, err := factory.AuthService.GetInviteToken(ctx, created.ID)
	if err != nil || invite == "" {
		t.Fatalf("invite token: %v", err)
	}

	var resetHash, inviteHash *string
	var resetExp, inviteExp *time.Time
	err = db.QueryRow(ctx, "SELECT reset_password_token, reset_password_token_expire, invite_token, invite_token_expire FROM users WHERE id = $1", created.ID).Scan(&resetHash, &resetExp, &inviteHash, &inviteExp)
	if err != nil {
		t.Fatalf("read token columns: %v", err)
	}

	cases := map[string]struct {
		plain string
		hash  *string
		exp   *time.Time
		ttl   time.Duration
	}{
		"reset":  {plain: reset, hash: resetHash, exp: resetExp, ttl: 10 * time.Minute},
		"invite": {plain: invite, hash: inviteHash, exp: inviteExp, ttl: 24 * time.Hour},
	}
	for name, tc := range cases {
		if tc.hash == nil {
			t.Errorf("expected stored %s hash", name)
			continue
		}
		sum := sha256.Sum256([]byte(tc.plain))
		if *tc.hash != hex.EncodeToString(sum[:]) {
			t.Errorf("stored %s hash must be sha256 of plaintext", name)
		}
		if tc.exp == nil || time.Until(*tc.exp) > tc.ttl {
			t.Errorf("expected ~%v %s expiry", tc.ttl, name)
		}
	}

	empty, err := factory.AuthService.GetResetPasswordToken(ctx, "missing-user-id")
	if err != nil || empty != "" {
		t.Errorf("expected empty token and nil error for missing user, got %q, %v", empty, err)
	}
}
