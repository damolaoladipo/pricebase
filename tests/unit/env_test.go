package unit

import (
	"os"
	"strings"
	"testing"

	"github.com/damolaoladipo/pricebase/internal/utils"
)

// testEnvVars are the required keys for utils.NewENV.
var testEnvVars = map[string]string{
	"GO_ENV":            "test",
	"APP_ENV":           "development",
	"APP_PORT":          "8080",
	"API_ROUTE":         "/api",
	"APP_CHANNELS":      "email",
	"AUTH_TYPE":         "development",
	"JWT_SECRET":        "test-secret",
	"JWT_EXPIRES":       "1h",
	"JWT_COOKIE_EXPIRE": "1h",
	"POSTGRES_URL":      "postgres://localhost:5432/pricebase_test?sslmode=disable",
	"POSTGRES_TEST_URL": "postgres://localhost:5432/pricebase_test?sslmode=disable",
	"CORS_ORIGINS":      "http://localhost:3000,http://localhost:3001",
	"SUPERADMIN_EMAIL":  "admin@example.com",
}

// initTestEnv sets every required env var (overridable per key) and loads the
// utils singleton. Callers must not depend on a nil singleton afterwards.
func initTestEnv(t *testing.T, overrides map[string]string) *utils.ENV {
	t.Helper()
	for k, v := range testEnvVars {
		if ov, ok := overrides[k]; ok {
			v = ov
		}
		t.Setenv(k, v)
	}
	env, err := utils.NewENV()
	if err != nil {
		t.Fatalf("NewENV: %v", err)
	}
	return env
}

func TestNewENVMissingKey(t *testing.T) {
	for k, v := range testEnvVars {
		t.Setenv(k, v)
	}
	if err := os.Unsetenv("JWT_SECRET"); err != nil {
		t.Fatal(err)
	}

	env, err := utils.NewENV()
	if err == nil {
		t.Fatal("expected error for missing JWT_SECRET")
	}
	if !strings.Contains(err.Error(), "JWT_SECRET") {
		t.Errorf("error should name the missing key, got %v", err)
	}
	if env != nil {
		t.Error("expected nil env on failure")
	}
}

func TestNewENVBadPort(t *testing.T) {
	for k, v := range testEnvVars {
		t.Setenv(k, v)
	}
	t.Setenv("APP_PORT", "not-a-port")
	if _, err := utils.NewENV(); err == nil {
		t.Error("expected error for non-integer APP_PORT")
	} else if !strings.Contains(err.Error(), "APP_PORT") {
		t.Errorf("error should name APP_PORT, got %v", err)
	}
}

func TestNewENVBadEmailFlag(t *testing.T) {
	for k, v := range testEnvVars {
		t.Setenv(k, v)
	}
	t.Setenv("EMAIL_ENABLED", "maybe")
	if _, err := utils.NewENV(); err == nil {
		t.Error("expected error for non-boolean EMAIL_ENABLED")
	}
}

func TestNewENVSuccessFlags(t *testing.T) {
	env := initTestEnv(t, map[string]string{
		"APP_ENV":      "production",
		"APP_PORT":     "9000",
		"CORS_ORIGINS": "https://a.example, https://b.example ",
	})

	if !env.IsProd || env.IsDev || env.IsStaging {
		t.Errorf("expected only IsProd, got %+v", env)
	}
	if env.APP_PORT != 9000 {
		t.Errorf("expected port 9000, got %d", env.APP_PORT)
	}
	if len(env.CORS_ORIGINS) != 2 || env.CORS_ORIGINS[0] != "https://a.example" {
		t.Errorf("CORS origins not split/trimmed: %v", env.CORS_ORIGINS)
	}
	if env.JWT_SECRET != "test-secret" {
		t.Errorf("secret mismatch: %q", env.JWT_SECRET)
	}
}

func TestComputeKey(t *testing.T) {
	if got := utils.ComputeKey("production", "pricebase.cache.jti"); got != "pricebase.cache.jti.prod" {
		t.Errorf("production key must carry .prod suffix, got %q", got)
	}
	if got := utils.ComputeKey("development", "k"); got != "k" {
		t.Errorf("development key must pass through, got %q", got)
	}
	if got := utils.ComputeKey("staging", "k"); got != "k" {
		t.Errorf("staging key must pass through, got %q", got)
	}
}

func TestCacheKeyUsesActiveEnv(t *testing.T) {
	initTestEnv(t, map[string]string{"APP_ENV": "production"})
	if got := utils.CacheKey("pricebase.cache.jti"); got != "pricebase.cache.jti.prod" {
		t.Errorf("CacheKey must namespace production, got %q", got)
	}

	initTestEnv(t, map[string]string{"APP_ENV": "development"})
	if got := utils.CacheKey("pricebase.cache.jti"); got != "pricebase.cache.jti" {
		t.Errorf("CacheKey must pass development through, got %q", got)
	}

	if got := utils.CacheEnv(); got != "development" {
		t.Errorf("CacheEnv must report active env, got %q", got)
	}
}
