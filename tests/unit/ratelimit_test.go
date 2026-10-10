package unit

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/damolaoladipo/pricebase/internal/middleware"
	"github.com/damolaoladipo/pricebase/internal/utils"
	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	middleware.RateLimiters()
	os.Exit(m.Run())
}

func serveRateLimited(t *testing.T, limiter gin.HandlerFunc, remoteAddr string) *httptest.ResponseRecorder {
	t.Helper()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(limiter)
	r.GET("/ping", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.RemoteAddr = remoteAddr
	r.ServeHTTP(w, req)
	return w
}

func TestRateLimiterBlocksOverLimit(t *testing.T) {
	env := &utils.ENV{RL_GLOBAL_WINDOW_MS: "60000", RL_GLOBAL_MAX: "2"}

	first := serveRateLimited(t, middleware.RateLimiter(env), "10.0.11.1:1234")
	second := serveRateLimited(t, middleware.RateLimiter(env), "10.0.11.1:1234")
	third := serveRateLimited(t, middleware.RateLimiter(env), "10.0.11.1:1234")

	if first.Code != 200 || second.Code != 200 || third.Code != 429 {
		t.Fatalf("want [200 200 429], got [%d %d %d]", first.Code, second.Code, third.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(third.Body.Bytes(), &body); err != nil {
		t.Fatalf("429 body is not JSON: %v", err)
	}
	if body["message"] != "rate limit exceeded" || body["status"] != float64(429) {
		t.Fatalf("unexpected 429 body: %v", body)
	}
}

func TestLoginRateLimiterEnvMax(t *testing.T) {
	env := &utils.ENV{RL_LOGIN_WINDOW_MS: "60000", RL_LOGIN_MAX: "1"}

	first := serveRateLimited(t, middleware.LoginRateLimiter(env), "10.0.11.2:1234")
	second := serveRateLimited(t, middleware.LoginRateLimiter(env), "10.0.11.2:1234")

	if first.Code != 200 || second.Code != 429 {
		t.Fatalf("want [200 429], got [%d %d]", first.Code, second.Code)
	}
}

func TestAuthAndProtectedLimitersAllow(t *testing.T) {
	authEnv := &utils.ENV{RL_AUTH_WINDOW_MS: "60000", RL_AUTH_MAX: "100"}
	if w := serveRateLimited(t, middleware.AuthRateLimiter(authEnv), "10.0.11.3:1234"); w.Code != 200 {
		t.Fatalf("auth limiter: want 200, got %d", w.Code)
	}

	protectedEnv := &utils.ENV{RL_PROTECTED_WINDOW_MS: "60000", RL_PROTECTED_MAX: "300"}
	if w := serveRateLimited(t, middleware.ProtectedRateLimiter(protectedEnv), "10.0.11.4:1234"); w.Code != 200 {
		t.Fatalf("protected limiter: want 200, got %d", w.Code)
	}

	if w := serveRateLimited(t, middleware.LimitRequests(), "10.0.11.5:1234"); w.Code != 200 {
		t.Fatalf("basic limiter: want 200, got %d", w.Code)
	}
}
