package middleware

import (
	"github.com/gin-gonic/gin"
)

const ContextCookies = "cookies"

// CookieParser returns Gin middleware that snapshots all cookies from the
// incoming HTTP request into a name-to-value map on the Gin context.
//
// Register CookieParser before handlers or middleware that need repeated cookie
// lookups without calling ctx.Cookie on each access. The middleware reads
// ctx.Request.Cookies(), builds map[string]string (last value wins if the
// server exposes duplicate names), and stores it with ctx.Set(ContextCookies, cookies).
//
// CookieParser does not validate, sign, or decrypt cookies; it only copies raw
// name/value pairs from the request. It does not write Set-Cookie response
// headers and never aborts—the chain always continues via ctx.Next().
func CookieParser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		cookies := map[string]string{}
		for _, cookie := range ctx.Request.Cookies() {
			cookies[cookie.Name] = cookie.Value
		}

		ctx.Set(ContextCookies, cookies)
		ctx.Next()
	}
}

// Cookie returns the value of the named cookie from the current HTTP request.
//
// It delegates to ctx.Cookie(name), which reads directly from the request's
// Cookie header rather than the map built by CookieParser. Use this for a
// single named lookup when CookieParser middleware is not mounted or when you
// prefer Gin's native cookie API.
//
// When the cookie is absent or ctx.Cookie returns an error, Cookie returns an
// empty string. It does not panic and does not write to the response or context.
func Cookie(ctx *gin.Context, name string) string {
	value, err := ctx.Cookie(name)
	if err != nil {
		return ""
	}
	return value
}

// Cookies returns the cookie map previously stored on the Gin context by
// CookieParser middleware.
//
// Use this in handlers that need the full cookie set or multiple lookups after
// CookieParser has run. It reads the ContextCookies key and type-asserts to
// map[string]string.
//
// When CookieParser has not run, the key is missing, or the stored value has an
// unexpected type, Cookies returns a new empty map (not nil) so callers can
// range or index safely. It does not read the request Cookie header directly.
func Cookies(ctx *gin.Context) map[string]string {
	value, exists := ctx.Get(ContextCookies)
	if !exists {
		return map[string]string{}
	}

	cookies, ok := value.(map[string]string)
	if !ok {
		return map[string]string{}
	}

	return cookies
}
