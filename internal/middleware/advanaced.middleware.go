package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	r "github.com/damolaoladipo/pricebase/internal/common/redis"
	"github.com/damolaoladipo/pricebase/internal/dtos"
	"github.com/damolaoladipo/pricebase/internal/utils"
)

// AdvancedResultsMiddleware populates gin.Context with paginated results.
//
// Mirrors the TS advancedResults  middleware:
//
//	cache === true:  fetch computeKey(ENV.APP_ENV, key); on hit set
//	  c("advancedResults") and return; on miss run Advanced (no db in the
//	  query — the pool is passed explicitly as its own argument),
//	  store the response for 1800s, then set it.
//	cache === false: run Advanced and set c("advancedResults").
//
// pool is the Postgres pool, passed explicitly (Go has no model object to
// carry the connection like the TS mongoose model does). Redis comes from
// the shared redis.REDIS singleton, mirroring the TS
// `import REDIS from '../common/services/redis.service'`. The payload stored
// under "advancedResults" is always dtos.ResponseDTO[any]. opts.Table is the
// TS `model`: the SQL table to query. opts.Data maps to the TS `query`/`q`
// extra filters, opts.CacheKey to `key`, opts.SortRef to the ref column
// (filtered by the `:id` URL param when present).
func AdvancedResultsMiddleware(pool *pgxpool.Pool, opts utils.ResultsMiddleware) gin.HandlerFunc {

	return func(c *gin.Context) {

		if pool == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, dtos.ResponseDTO[any]{
				Error:   true,
				Errors:  []any{},
				Message: "database not initialized",
				Status:  http.StatusInternalServerError,
			})
			return
		}

		// if cache is enabled and cache key is not empty, fetch data from cache
		if opts.Cache && opts.CacheKey != "" {
			// fetch data from cache
			cached, err := r.REDIS.FetchData(c.Request.Context(), utils.CacheKey(opts.CacheKey))
			if err == nil && cached != nil {
				var hit dtos.ResponseDTO[any]
				if raw, mErr := json.Marshal(cached); mErr == nil {
					if uErr := json.Unmarshal(raw, &hit); uErr == nil && hit.Data != nil {
						c.Set("advancedResults", hit)
						c.Next()
						return
					}
				}
			}

			// make query parameters
			queryParam := make(map[string][]string, len(c.Request.URL.Query()))
			for k, v := range c.Request.URL.Query() {
				queryParam[k] = v
			}

			// get reference value — only when the route has an `:id` param.
			// Otherwise Value stays nil and Advanced applies no ref filter
			// (c.Param returns "" when absent, and "" is non-nil, so it
			// must not be passed through as Value).
			var value any
			if id := c.Param("id"); id != "" {
				value = id
			}

			// prepare advanced query
			query := utils.AdvancedQuery{
				Table:      opts.Table,
				Columns:    opts.Columns,
				Fields:     opts.Fields,
				Ref:        opts.SortRef,
				Value:      value,
				Data:       opts.Data,
				Operator:   opts.Operator,
				Query:      opts.Query,
				QueryParam: queryParam,
				Paginate:   opts.Paginate,
			}

			// run advanced query
			result, err := utils.Advanced(c.Request.Context(), pool, query)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, dtos.ResponseDTO[any]{
					Error:   true,
					Errors:  []any{},
					Message: err.Error(),
					Status:  http.StatusInternalServerError,
				})

				return
			}

			count := int64(result.Count)
			total := result.Total

			var pagination *dtos.PaginationDTO
			if result.Pagination != nil && (result.Pagination.Next != nil || result.Pagination.Prev != nil) {
				pagination = &dtos.PaginationDTO{}
				if result.Pagination.Next != nil {
					pagination.Next = dtos.PaginationPageDTO{
						Page:  result.Pagination.Next.Page,
						Limit: result.Pagination.Next.Limit,
					}
				}
				if result.Pagination.Prev != nil {
					pagination.Prev = dtos.PaginationPageDTO{
						Page:  result.Pagination.Prev.Page,
						Limit: result.Pagination.Prev.Limit,
					}
				}
			}

			response := dtos.ResponseDTO[any]{
				Error:      false,
				Errors:     []any{},
				Count:      &count,
				Total:      &total,
				Pagination: pagination,
				Data:       result.Data,
				Message:    "successful",
				Status:     http.StatusOK,
			}

			_ = r.REDIS.KeepData(c.Request.Context(), r.Data{
				Key:   utils.CacheKey(opts.CacheKey),
				Value: response,
			}, 1800)

			c.Set("advancedResults", response)
			c.Next()
			return
		}

		// if cache is disabled or cache key is empty, run advanced query
		if !opts.Cache || opts.CacheKey == "" {

			queryParam := make(map[string][]string, len(c.Request.URL.Query()))
			for k, v := range c.Request.URL.Query() {
				queryParam[k] = v
			}

			var value any
			if id := c.Param("id"); id != "" {
				value = id
			}

			// prepare advanced query
			query := utils.AdvancedQuery{
				Table:      opts.Table,
				Columns:    opts.Columns,
				Fields:     opts.Fields,
				Ref:        opts.SortRef,
				Value:      value,
				Data:       opts.Data,
				Operator:   opts.Operator,
				Query:      opts.Query,
				QueryParam: queryParam,
				Paginate:   opts.Paginate,
			}

			// run advanced query
			result, err := utils.Advanced(c.Request.Context(), pool, query)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, dtos.ResponseDTO[any]{
					Error:   true,
					Errors:  []any{},
					Message: err.Error(),
					Status:  http.StatusInternalServerError,
				})
				return
			}

			count := int64(result.Count)
			total := result.Total

			var pagination *dtos.PaginationDTO
			if result.Pagination != nil && (result.Pagination.Next != nil || result.Pagination.Prev != nil) {
				pagination = &dtos.PaginationDTO{}
				if result.Pagination.Next != nil {
					pagination.Next = dtos.PaginationPageDTO{
						Page:  result.Pagination.Next.Page,
						Limit: result.Pagination.Next.Limit,
					}
				}
				if result.Pagination.Prev != nil {
					pagination.Prev = dtos.PaginationPageDTO{
						Page:  result.Pagination.Prev.Page,
						Limit: result.Pagination.Prev.Limit,
					}
				}
			}

			// set advanced results to the context
			c.Set("advancedResults", dtos.ResponseDTO[any]{
				Error:      false,
				Errors:     []any{},
				Count:      &count,
				Total:      &total,
				Pagination: pagination,
				Data:       result.Data,
				Message:    "successful",
				Status:     http.StatusOK,
			})
			c.Next()

		}

	}
}
