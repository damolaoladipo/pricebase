package middleware

import (
	"github.com/damolaoladipo/pricebase/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/unrolled/secure"
)

// Security returns Gin middleware that applies HTTP security and hardening
// headers to every response using github.com/unrolled/secure.
//
// Mount Security near the top of the Gin pipeline so downstream handlers
// inherit hardened defaults before any response bytes are written. The
// middleware wraps the response writer and request through secure.Process,
// then adds supplemental headers and strips X-Powered-By before and after
// ctx.Next().
//
// The env parameter controls development versus production behavior. When env
// is nil or not production (IsProd false), IsDevelopment is enabled on the
// secure.Options, which relaxes Strict-Transport-Security enforcement
// (ForceSTSHeader false). In production, STS is forced with a one-year max-age
// and includeSubdomains. Other fixed options include Content-Type nosniff,
// SAMEORIGIN framing, a restrictive Content-Security-Policy, Permissions-Policy
// disabling sensitive device APIs, Referrer-Policy no-referrer, and
// X-Forwarded-Proto https SSL proxy header trust.
//
// Request and response interaction:
//   - Reads ctx.Request and writes security headers via ctx.Writer.
//   - Sets X-Download-Options and Origin-Agent-Cluster on the response.
//   - Deletes X-Powered-By from response headers before and after the chain.
//
// Failure behavior: if secure.Process returns an error, or if the writer
// status is a redirect (301–399) after processing, the middleware calls
// ctx.Abort() and does not invoke downstream handlers. On success it calls
// ctx.Next() so the remainder of the chain runs under the applied headers.
func Security(env *utils.ENV) gin.HandlerFunc {
	isDev := env == nil || !env.IsProd
	secureMiddleware := secure.New(secure.Options{
		IsDevelopment:                 isDev,
		SSLProxyHeaders:               map[string]string{"X-Forwarded-Proto": "https"},
		STSSeconds:                    31536000,
		STSIncludeSubdomains:          true,
		ForceSTSHeader:                !isDev,
		ContentTypeNosniff:            true,
		CustomFrameOptionsValue:       "SAMEORIGIN",
		CustomBrowserXssValue:         "0",
		ReferrerPolicy:                "no-referrer",
		CrossOriginOpenerPolicy:       "same-origin",
		CrossOriginResourcePolicy:     "cross-origin",
		XDNSPrefetchControl:           "off",
		XPermittedCrossDomainPolicies: "none",
		ContentSecurityPolicy:         "default-src 'none'; frame-ancestors 'self'; base-uri 'none'; form-action 'none'",
		PermissionsPolicy:             "accelerometer=(), camera=(), geolocation=(), gyroscope=(), magnetometer=(), microphone=(), payment=(), usb=()",
	})

	return func(ctx *gin.Context) {
		if err := secureMiddleware.Process(ctx.Writer, ctx.Request); err != nil {
			ctx.Abort()
			return
		}

		if status := ctx.Writer.Status(); status > 300 && status < 399 {
			ctx.Abort()
			return
		}

		ctx.Header("X-Download-Options", "noopen")
		ctx.Header("Origin-Agent-Cluster", "?1")
		ctx.Writer.Header().Del("X-Powered-By")
		ctx.Next()
		ctx.Writer.Header().Del("X-Powered-By")
	}
}
