package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/microcosm-cc/bluemonday"
)

var htmlPolicy = bluemonday.StrictPolicy()

// Sanitize returns Gin middleware that strips MongoDB operator keys and HTML/XSS
// from incoming query parameters and JSON request bodies before handlers execute.
//
// MongoDB-style keys prefixed with "$" or containing "." are removed from query
// strings and JSON objects. String values are cleaned with bluemonday.StrictPolicy,
// which strips HTML and script payloads (the Go equivalent of xss-clean).
// Register Sanitize before handlers that bind query or JSON input.
//
// JSON body handling runs when Content-Type contains "application/json" and the
// body is non-nil. On read, unmarshal, or re-marshal failure the original bytes
// are restored so handlers can return their own validation errors.
//
// The middleware never aborts the chain; it always calls ctx.Next().
func Sanitize() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		query := ctx.Request.URL.Query()
		queryChanged := false
		for key, values := range query {
			if isUnsafeKey(key) {
				query.Del(key)
				queryChanged = true
				continue
			}
			for i, value := range values {
				cleaned := htmlPolicy.Sanitize(value)
				if cleaned != value {
					values[i] = cleaned
					queryChanged = true
				}
			}
		}
		if queryChanged {
			ctx.Request.URL.RawQuery = query.Encode()
		}

		contentType := ctx.GetHeader("Content-Type")
		if !strings.Contains(contentType, "application/json") || ctx.Request.Body == nil {
			ctx.Next()
			return
		}

		raw, err := io.ReadAll(ctx.Request.Body)
		_ = ctx.Request.Body.Close()
		if err != nil || len(raw) == 0 {
			ctx.Request.Body = io.NopCloser(bytes.NewReader(raw))
			ctx.Next()
			return
		}

		var payload any
		if err := json.Unmarshal(raw, &payload); err != nil {
			ctx.Request.Body = io.NopCloser(bytes.NewReader(raw))
			ctx.Next()
			return
		}

		encoded, err := json.Marshal(sanitizeValue(payload))
		if err != nil {
			ctx.Request.Body = io.NopCloser(bytes.NewReader(raw))
			ctx.Next()
			return
		}

		ctx.Request.Body = io.NopCloser(bytes.NewReader(encoded))
		ctx.Request.ContentLength = int64(len(encoded))
		ctx.Request.Header.Set("Content-Length", strconv.Itoa(len(encoded)))

		ctx.Next()
	}
}

// isUnsafeKey reports whether key should be stripped because it resembles a
// MongoDB operator or dotted field path.
func isUnsafeKey(key string) bool {
	return strings.HasPrefix(key, "$") || strings.Contains(key, ".")
}

// sanitizeValue recursively removes unsafe map keys and HTML-sanitizes strings.
func sanitizeValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		clean := map[string]any{}
		for key, item := range typed {
			if isUnsafeKey(key) {
				continue
			}
			clean[key] = sanitizeValue(item)
		}
		return clean
	case []any:
		clean := make([]any, 0, len(typed))
		for _, item := range typed {
			clean = append(clean, sanitizeValue(item))
		}
		return clean
	case string:
		return htmlPolicy.Sanitize(typed)
	default:
		return value
	}
}
