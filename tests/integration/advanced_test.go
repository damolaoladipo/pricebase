package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/damolaoladipo/pricebase/internal/config/seeds"
	"github.com/damolaoladipo/pricebase/internal/dtos"
	"github.com/damolaoladipo/pricebase/internal/middleware"
	"github.com/damolaoladipo/pricebase/internal/utils"
)

// advancedResponse mirrors the JSON of dtos.ResponseDTO[any].
type advancedResponse struct {
	Error      bool             `json:"error"`
	Count      int              `json:"count"`
	Total      int64            `json:"total"`
	Message    string           `json:"message"`
	Status     int              `json:"status"`
	Data       []map[string]any `json:"data"`
	Pagination *struct {
		Next *struct {
			Page  int `json:"page"`
			Limit int `json:"limit"`
		} `json:"next"`
		Prev *struct {
			Page  int `json:"page"`
			Limit int `json:"limit"`
		} `json:"prev"`
	} `json:"pagination"`
}

func TestAdvancedMiddlewareLive(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testDB(t)
	truncateData(t, db)
	ctx := context.Background()

	if err := seeds.SeedData(ctx, db); err != nil {
		t.Fatalf("seed: %v", err)
	}

	engine := gin.New()
	engine.Use(middleware.ErrorHandler())
	engine.GET("/users", middleware.AdvancedResultsMiddleware(db, utils.ResultsMiddleware{Table: "users"}), func(c *gin.Context) {
		raw, ok := c.Get("advancedResults")
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "missing advancedResults"})
			return
		}
		res, ok := raw.(dtos.ResponseDTO[any])
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "wrong advancedResults type"})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/users?limit=2", nil)
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var res advancedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if res.Error {
		t.Error("expected error=false")
	}
	if res.Message != "successful" || res.Status != 200 {
		t.Errorf("envelope mismatch: %+v", res)
	}
	if res.Total == 0 {
		t.Error("expected seeded users, total=0")
	}
	if len(res.Data) > 2 {
		t.Errorf("expected at most 2 rows with limit=2, got %d", len(res.Data))
	}
	if res.Pagination == nil && res.Total > int64(len(res.Data)) {
		t.Error("expected pagination object when more rows exist")
	}
	if res.Pagination != nil && res.Pagination.Next == nil && res.Total > int64(len(res.Data)) {
		t.Error("expected next page when more rows exist")
	}
}

func TestAdvancedMiddlewareEmptyTable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testDB(t)
	truncateData(t, db)

	engine := gin.New()
	engine.Use(middleware.ErrorHandler())
	engine.GET("/users", middleware.AdvancedResultsMiddleware(db, utils.ResultsMiddleware{Table: "users"}), func(c *gin.Context) {
		raw, _ := c.Get("advancedResults")
		res, ok := raw.(dtos.ResponseDTO[any])
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "wrong type"})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on empty table, got %d: %s", w.Code, w.Body.String())
	}

	var res advancedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if res.Total != 0 || res.Count != 0 {
		t.Errorf("expected zero totals on empty table, got %+v", res)
	}
}

func TestAdvancedMiddlewareCacheHit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testDB(t)
	truncateData(t, db)
	ctx := context.Background()

	if err := seeds.SeedData(ctx, db); err != nil {
		t.Fatalf("seed: %v", err)
	}
	initIntegrationEnv(t)
	testRedis(t)

	cacheKey := "test.advanced.users." + utils.GenerateID()
	engine := gin.New()
	engine.Use(middleware.ErrorHandler())
	engine.GET("/users", middleware.AdvancedResultsMiddleware(db, utils.ResultsMiddleware{
		Table:    "users",
		Cache:    true,
		CacheKey: cacheKey,
	}), func(c *gin.Context) {
		raw, _ := c.Get("advancedResults")
		res, ok := raw.(dtos.ResponseDTO[any])
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "wrong type"})
			return
		}
		c.JSON(http.StatusOK, res)
	})
	call := func() advancedResponse {
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/users", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var res advancedResponse
		if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
			t.Fatalf("decode: %v", err)
		}
		return res
	}

	first := call()
	if first.Total == 0 {
		t.Fatal("expected seeded users on miss")
	}

	// Empty the table: a live query would now return zero, so a stale total
	// on the second call proves the response came from cache.
	truncateData(t, db)
	second := call()
	if second.Total != first.Total {
		t.Errorf("expected cached total %d, got %d", first.Total, second.Total)
	}
}
