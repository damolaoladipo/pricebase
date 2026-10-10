package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

	applogger "github.com/damolaoladipo/pricebase/internal/common/logger"
	"github.com/damolaoladipo/pricebase/internal/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery returns Gin middleware that catches panics from downstream handlers
// and converts them into a JSON error response instead of crashing the process.
//
// Mount Recovery early in the chain (often first or immediately after logging)
// so every handler and middleware benefits from the defer/recover boundary.
// The middleware calls ctx.Next() inside a deferred recover; when no panic
// occurs, the defer is a no-op and the response passes through unchanged.
//
// The env parameter controls diagnostic logging. When env is non-nil and either
// IsDev or IsStaging is true, the recovered value and full debug.Stack() are
// printed to stdout. Production environments omit stack traces from logs unless
// configured elsewhere.
//
// Panic value handling: non-error panic values are wrapped with fmt.Errorf("%v").
// If the resulting error unwraps to *utils.ErrorResponse, StatusCode, Message,
// Errors, and Data fields populate the JSON payload; zero or empty fields fall
// back to HTTP 500 and "Server Error". Data is included only when it type-asserts
// to map[string]any.
//
// Response interaction: if ctx.Writer.Written() is already true when a panic is
// recovered (partial response flushed), the middleware calls ctx.Abort() without
// writing additional JSON to avoid corrupting the client stream. Otherwise it
// responds with ctx.AbortWithStatusJSON using a envelope:
// error, errors, data, message, and status fields.
func Recovery(env *utils.ENV) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		defer func() {
			recovered := recover()
			if recovered == nil {
				return
			}

			err, ok := recovered.(error)
			if !ok {
				err = fmt.Errorf("%v", recovered)
			}

			if env != nil && (env.IsDev || env.IsStaging) {
				fmt.Println("the error:", err)
				fmt.Println(string(debug.Stack()))
			}

			applogger.SystemLogger.Error("unhandled panic",
				zap.Error(err),
				zap.String("path", ctx.Request.URL.Path),
				zap.String("method", ctx.Request.Method),
			)

			if ctx.Writer.Written() {
				ctx.Abort()
				return
			}

			status := http.StatusInternalServerError
			message := "Server Error"
			errorsList := []any{}
			data := map[string]any{}

			var custom *utils.ErrorResponse
			if errors.As(err, &custom) && custom != nil {
				if custom.Status != 0 {
					status = custom.Status
				}
				if custom.Message != "" {
					message = custom.Message
				}
				if custom.Errors != nil {
					errorsList = utils.ErrorsAsAny(custom.Errors)
				}
				if custom.Data != nil {
					if mapped, ok := custom.Data.(map[string]any); ok {
						data = mapped
					}
				}
			}

			ctx.AbortWithStatusJSON(status, gin.H{
				"error":   true,
				"errors":  errorsList,
				"data":    data,
				"message": message,
				"status":  status,
			})
		}()

		ctx.Next()
	}
}
