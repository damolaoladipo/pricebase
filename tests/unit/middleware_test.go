package unit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/damolaoladipo/pricebase/internal/middleware"
	"github.com/damolaoladipo/pricebase/internal/utils"
)

func errorStatus(w *httptest.ResponseRecorder, c *gin.Context) int {
	if !c.IsAborted() {
		return 0
	}
	if len(c.Errors) == 0 {
		return w.Code
	}
	if custom, ok := c.Errors.Last().Err.(*utils.ErrorResponse); ok {
		return custom.Status
	}
	return w.Code
}

func TestProtectNoHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/users", nil)

	middleware.Protect(nil, nil)(c)

	if !c.IsAborted() {
		t.Fatal("expected abort without Authorization header")
	}
	if got := errorStatus(w, c); got != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", got)
	}
}

func TestProtectNonBearer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/users", nil)
	c.Request.Header.Set("Authorization", "Basic dXNlcjpwYXNz")

	middleware.Protect(nil, nil)(c)

	if !c.IsAborted() {
		t.Fatal("expected abort for non-Bearer scheme")
	}
	if got := errorStatus(w, c); got != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", got)
	}
}

func TestProtectMalformedBearer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/users", nil)
	c.Request.Header.Set("Authorization", "Bearer")

	middleware.Protect(nil, nil)(c)

	if !c.IsAborted() {
		t.Fatal("expected abort for tokenless Bearer header")
	}
	if got := errorStatus(w, c); got != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", got)
	}
}

func TestAuthorizeNoUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/users", nil)

	middleware.Authorize(nil, middleware.AllRoles, nil)(c)

	if !c.IsAborted() {
		t.Fatal("expected abort without authenticated user")
	}
	if got := errorStatus(w, c); got != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", got)
	}
}

func TestAuthorizeWrongContextType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/users", nil)
	c.Set("user", "not-an-auth-user")

	middleware.Authorize(nil, middleware.AllRoles, nil)(c)

	if !c.IsAborted() {
		t.Fatal("expected abort for mistyped user value")
	}
	if got := errorStatus(w, c); got != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", got)
	}
}

func TestRoleLists(t *testing.T) {
	if len(middleware.SuperAdminRoles) != 1 || middleware.SuperAdminRoles[0] != string(utils.UserTypeSuper) {
		t.Errorf("superadmin list mismatch: %v", middleware.SuperAdminRoles)
	}
	if len(middleware.AllRoles) != 5 {
		t.Errorf("allRoles must hold five entries, got %v", middleware.AllRoles)
	}
	found := false
	for _, r := range middleware.AllRoles {
		if r == string(utils.UserTypeUser) {
			found = true
		}
	}
	if !found {
		t.Errorf("allRoles must include the user type: %v", middleware.AllRoles)
	}
}

func TestAdvancedNilPool(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/users?limit=2", nil)

	middleware.AdvancedResultsMiddleware(nil, utils.ResultsMiddleware{Table: "users"})(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 for nil pool, got %d", w.Code)
	}
}
