package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/damolaoladipo/pricebase/internal/config/seeds"
	"github.com/damolaoladipo/pricebase/internal/dtos"
	"github.com/damolaoladipo/pricebase/internal/factory"
	"github.com/damolaoladipo/pricebase/internal/middleware"
	"github.com/damolaoladipo/pricebase/internal/models"
	"github.com/damolaoladipo/pricebase/internal/repository"
	"github.com/damolaoladipo/pricebase/internal/utils"
)

func TestProtectAPIKeyLive(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testDB(t)
	truncateData(t, db)
	ctx := context.Background()

	if err := seeds.SeedData(ctx, db); err != nil {
		t.Fatalf("seed: %v", err)
	}
	initIntegrationEnv(t)
	factory.Factories(db)
	userRepo := repository.NewUserRepository(db)

	newUser := func(email, userType string, active bool) models.User {
		created, err := userRepo.CreateUser(ctx, models.User{
			Email:       email,
			UserType:    userType,
			FirstName:   "Test",
			LastName:    "User",
			IsActive:    true,
			IsActivated: true,
		})
		if err != nil {
			t.Fatalf("create user: %v", err)
		}
		updated, err := factory.AuthService.GenerateAPIKey(ctx, &created)
		if err != nil {
			t.Fatalf("generate api key: %v", err)
		}
		if !active {
			if _, err := db.Exec(ctx, "UPDATE users SET is_active = false WHERE id = $1", created.ID); err != nil {
				t.Fatalf("deactivate: %v", err)
			}
		}
		return *updated
	}

	engine := gin.New()
	engine.Use(middleware.ErrorHandler())
	engine.GET("/me",
		middleware.Protect(factory.AuthService, userRepo),
		func(c *gin.Context) {
			raw, _ := c.Get("user")
			user := raw.(dtos.AuthUser)
			rawAuth, _ := c.Get("auth")
			auth := rawAuth.(dtos.AuthSession)
			c.JSON(http.StatusOK, gin.H{"email": user.Email, "type": user.UserType, "jti": auth.JTI})
		},
	)
	call := func(secret string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		req.Header.Set("Authorization", "Bearer "+secret)
		engine.ServeHTTP(w, req)
		return w
	}

	t.Run("superadmin api key passes", func(t *testing.T) {
		user := newUser("super-key@example.com", string(utils.UserTypeSuper), true)
		w := call(user.APIKey.Secret)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("organization api key passes", func(t *testing.T) {
		user := newUser("org-key@example.com", string(utils.UserTypeOrganization), true)
		w := call(user.APIKey.Secret)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("plain user api key forbidden", func(t *testing.T) {
		user := newUser("plain-key@example.com", string(utils.UserTypeUser), true)
		w := call(user.APIKey.Secret)
		if w.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("deactivated user forbidden", func(t *testing.T) {
		user := newUser("off-key@example.com", string(utils.UserTypeSuper), false)
		w := call(user.APIKey.Secret)
		if w.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("unknown key unauthenticated", func(t *testing.T) {
		w := call("sk_test_not-a-real-key")
		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestProtectJWTSessionLive(t *testing.T) {
	gin.SetMode(gin.TestMode)
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

	created, err := userRepo.CreateUser(ctx, models.User{
		Email:       "jwt-user@example.com",
		UserType:    "user",
		FirstName:   "Jwt",
		LastName:    "User",
		IsActive:    true,
		IsActivated: true,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	token, jti, err := factory.AuthService.GenerateJWT(ctx, dtos.GenerateJWTDTO{
		Email: created.Email,
		ID:    created.ID,
	})
	if err != nil {
		t.Fatalf("mint session: %v", err)
	}

	engine := gin.New()
	engine.Use(middleware.ErrorHandler())
	engine.GET("/me",
		middleware.Protect(factory.AuthService, userRepo),
		func(c *gin.Context) {
			raw, _ := c.Get("user")
			user := raw.(dtos.AuthUser)
			rawAuth, _ := c.Get("auth")
			auth := rawAuth.(dtos.AuthSession)
			c.JSON(http.StatusOK, gin.H{"email": user.Email, "jti": auth.JTI})
		},
	)
	call := func(token string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		engine.ServeHTTP(w, req)
		return w
	}

	t.Run("live session passes with jti", func(t *testing.T) {
		w := call(token)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), jti) {
			t.Errorf("expected jti in response: %s", w.Body.String())
		}
	})

	t.Run("tampered token rejected", func(t *testing.T) {
		w := call(token[:len(token)-2] + "xx")
		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("revoked session rejected", func(t *testing.T) {
		if err := factory.AuthService.RevokeSession(ctx, jti); err != nil {
			t.Fatalf("revoke: %v", err)
		}
		w := call(token)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d: %s", w.Code, w.Body.String())
		}
	})
}
