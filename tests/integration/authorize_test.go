package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	commontypes "github.com/damolaoladipo/common-go/types"
	"github.com/damolaoladipo/pricebase/internal/config/seeds"
	"github.com/damolaoladipo/pricebase/internal/dtos"
	"github.com/damolaoladipo/pricebase/internal/factory"
	"github.com/damolaoladipo/pricebase/internal/middleware"
	"github.com/damolaoladipo/pricebase/internal/repository"
)

func TestAuthorizeLive(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testDB(t)
	truncateData(t, db)
	ctx := context.Background()

	if err := seeds.SeedData(ctx, db); err != nil {
		t.Fatalf("seed: %v", err)
	}
	factory.Factories(db)

	roleRepo := repository.NewRoleRepository(db)
	superRole, err := roleRepo.FindByName(ctx, "superadmin")
	if err != nil {
		t.Fatalf("find superadmin role: %v", err)
	}

	perms, err := factory.AuthService.ExtractPermissions(ctx, superRole.ID)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if len(perms) == 0 {
		t.Fatal("expected seeded permissions")
	}
	shared := make([]commontypes.UserPermission, 0, len(perms))
	for _, p := range perms {
		shared = append(shared, commontypes.UserPermission{Group: p.Group, Actions: p.Actions})
	}

	setUser := func(user dtos.AuthUser) gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set("user", user)
			c.Next()
		}
	}
	ok := func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) }

	t.Run("allowed role passes", func(t *testing.T) {
		engine := gin.New()
		engine.Use(middleware.ErrorHandler())
		engine.GET("/r",
			setUser(dtos.AuthUser{ID: "u1", Email: "a@b.c", UserType: "superadmin", RoleIDs: []string{superRole.ID}, Permissions: shared}),
			middleware.Authorize(db, middleware.AllRoles, nil),
			ok,
		)

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/r", nil))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("unknown role rejected", func(t *testing.T) {
		engine := gin.New()
		engine.Use(middleware.ErrorHandler())
		engine.GET("/r",
			setUser(dtos.AuthUser{ID: "u1", UserType: "user", RoleIDs: []string{"missing-role-id"}, Permissions: shared}),
			middleware.Authorize(db, middleware.AllRoles, nil),
			ok,
		)

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/r", nil))
		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("no user rejected", func(t *testing.T) {
		engine := gin.New()
		engine.Use(middleware.ErrorHandler())
		engine.GET("/r", middleware.Authorize(db, middleware.AllRoles, nil), ok)

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/r", nil))
		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("permission granted", func(t *testing.T) {
		engine := gin.New()
		engine.Use(middleware.ErrorHandler())
		engine.GET("/r",
			setUser(dtos.AuthUser{ID: "u1", UserType: "superadmin", RoleIDs: []string{superRole.ID}, Permissions: shared}),
			middleware.Authorize(db, middleware.AllRoles, &dtos.PermitParams{
				Group:   perms[0].Group,
				Actions: []string{perms[0].Actions[0]},
				Roles:   middleware.AllRoles,
			}),
			ok,
		)

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/r", nil))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("permission denied", func(t *testing.T) {
		engine := gin.New()
		engine.Use(middleware.ErrorHandler())
		engine.GET("/r",
			setUser(dtos.AuthUser{ID: "u1", UserType: "superadmin", RoleIDs: []string{superRole.ID}, Permissions: shared}),
			middleware.Authorize(db, middleware.AllRoles, &dtos.PermitParams{
				Group:   perms[0].Group,
				Actions: []string{"action-that-does-not-exist"},
				Roles:   middleware.AllRoles,
			}),
			ok,
		)

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/r", nil))
		if w.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d: %s", w.Code, w.Body.String())
		}
	})
}
