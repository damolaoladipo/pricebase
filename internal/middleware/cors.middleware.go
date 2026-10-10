package middleware

import (
	"net/http"
	"time"

	"github.com/damolaoladipo/pricebase/internal/utils"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS returns Gin middleware that applies Cross-Origin Resource Sharing headers
// using github.com/gin-contrib/cors.
//
// Register CORS before route handlers so simple and preflight requests receive
// consistent Access-Control-* headers. Allowed origins come from env.CORS_ORIGINS
// (comma-separated CORS_ORIGINS). Matching origins are reflected via AllowOriginFunc
// so credentialed browser requests work (Access-Control-Allow-Credentials cannot
// be combined with Access-Control-Allow-Origin: *).
//
// Allowed methods, headers, and exposed headers match the previous custom CORS
// policy. Preflight responses use 204 No Content and are cached for 24 hours.
func CORS(env *utils.ENV) gin.HandlerFunc {
	origins := []string{"http://localhost:3000"}
	if env != nil && len(env.CORS_ORIGINS) > 0 {
		origins = env.CORS_ORIGINS
	}

	allowed := make(map[string]struct{}, len(origins))
	for _, origin := range origins {
		allowed[origin] = struct{}{}
	}

	return cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			_, ok := allowed[origin]
			return ok
		},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodHead,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			"Accept",
			"Accept-Language",
			"Authorization",
			"Content-Language",
			"Content-Type",
			"Cookie",
			"Origin",
			"Referer",
			"User-Agent",
			"X-Requested-With",
			"X-Request-ID",
			"X-Correlation-ID",
			"x-access-token",
			"x-api-key",
			"x-client-id",
			"x-device-id",
		},
		ExposeHeaders: []string{
			"Authorization",
			"Content-Disposition",
			"Content-Length",
			"Content-Type",
			"X-Request-ID",
			"X-Correlation-ID",
			"x-access-token",
		},
		AllowCredentials:          true,
		AllowPrivateNetwork:       true,
		AllowWebSockets:           true,
		MaxAge:                    24 * time.Hour,
		OptionsResponseStatusCode: http.StatusNoContent,
	})
}
