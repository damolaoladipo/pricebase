package middleware

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// PublicStatic returns Gin middleware that serves static files from dir for
// matching GET and HEAD requests before the rest of the handler chain runs.
//
// The dir argument is cleaned once at middleware construction with
// filepath.Clean and used as the filesystem root. For each request, the URL path
// (without leading slash) is joined under that root. Paths that escape the root
// via directory traversal (for example "../") are rejected: the middleware
// calls ctx.Next() and does not serve a file. Requests for directories, missing
// files, or non-GET/HEAD methods also fall through to ctx.Next().
//
// Request and response interaction:
//   - Reads ctx.Request.Method and ctx.Request.URL.Path.
//   - On match, calls ctx.File(full) to stream the file and ctx.Abort() so
//     downstream handlers are not invoked.
//
// When a file is served, the response status and headers are set by Gin's
// ctx.File helper (including Content-Type where inferable). No environment
// variables or options struct are used; behavior is fully determined by dir
// and the incoming path. Failed stat checks silently pass through to the next
// handler rather than returning an error response from this middleware.
func PublicStatic(dir string) gin.HandlerFunc {
	cleanDir := filepath.Clean(dir)

	return func(ctx *gin.Context) {
		if ctx.Request.Method != http.MethodGet && ctx.Request.Method != http.MethodHead {
			ctx.Next()
			return
		}

		rel := strings.TrimPrefix(ctx.Request.URL.Path, "/")
		if rel == "" {
			ctx.Next()
			return
		}

		full := filepath.Clean(filepath.Join(cleanDir, rel))
		if full != cleanDir && !strings.HasPrefix(full, cleanDir+string(os.PathSeparator)) {
			ctx.Next()
			return
		}

		info, err := os.Stat(full)
		if err != nil || info.IsDir() {
			ctx.Next()
			return
		}

		ctx.File(full)
		ctx.Abort()
	}
}
