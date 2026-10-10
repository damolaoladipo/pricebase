package config

import (
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"

	"github.com/damolaoladipo/pricebase/internal/middleware"
	"github.com/damolaoladipo/pricebase/internal/routes"
	"github.com/damolaoladipo/pricebase/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// bodyLimit caps request bodies and multipart uploads at 200 MiB.
const bodyLimit = 200 * 1024 * 1024

func App(db *pgxpool.Pool) *gin.Engine {

	// laod ENV - reuse bootstrap singleton when available, else load strictly.
	env := utils.Current()
	if env == nil {
		var err error
		env, err = utils.NewENV()
		if err != nil {
			log.Fatal(err)
		}
	}

	if env.IsProd {
		gin.SetMode(gin.ReleaseMode)
	}

	// gin.New() starts with an empty chain.
	// So we can add our middlewares to the chain.
	app := gin.New()

	// Engine-level config, not middleware.
	// Trust X-Forwarded-* from any proxy so ClientIP() is correct behind a
	// load balancer. Narrow this to your LB's CIDR in production.
	_ = app.SetTrustedProxies([]string{"0.0.0.0/0", "::/0"})
	app.MaxMultipartMemory = bodyLimit

	// Recovery middleware
	// so a panic anywhere below is caught and turned
	// into a 500 instead of killing the process.
	app.Use(middleware.Recovery(env))

	// RequestID — assigns X-Request-ID early so error responses carry it too.
	app.Use(middleware.RequestID())

	// Error handling middleware
	// This will catch any errors and return a 500 response.
	app.Use(middleware.ErrorHandler())

	// CORS — early, so preflight OPTIONS is answered before the rest of the
	// chain does work, and so CORS headers are present on error responses too.
	app.Use(middleware.CORS(env))

	// Security — sets response hardening headers (HSTS, nosniff,
	// frame options). Header-only, no request parsing.
	app.Use(middleware.Security(env))

	// Body limit — wraps Request.Body in a MaxBytesReader so an oversized
	// upload fails at read time instead of buffering into memory. Must run
	// before anything that reads the body.
	app.Use(func(ctx *gin.Context) {
		ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, bodyLimit)
		ctx.Next()
	})

	// CookieParser — parses cookies into the context so downstream
	// middleware and handlers can read them via middleware.Cookie(ctx, name).
	app.Use(middleware.CookieParser())

	// Sessions — cookie-backed session store, like express-session.
	// JWT_SECRET signs and encrypts the session cookie. Handlers read and
	// write the session via sessions.Default(ctx): session.Set("token",
	// auth.Token), session.Options(...), session.Save(). Defaults below
	// mirror AuthService.GetAuthToken: 70-day lifetime, plain HTTP unless
	// production, SameSite none.
	sessionStore := cookie.NewStore([]byte(env.JWT_SECRET))
	sessionStore.Options(sessions.Options{
		Path:     "/",
		MaxAge:   70 * 24 * 60 * 60,
		HttpOnly: false,
		Secure:   env.IsProd,
		SameSite: http.SameSiteNoneMode,
	})
	app.Use(sessions.Sessions("pricebase_session", sessionStore))

	// Agent — parses User-Agent into a struct on the context for device
	// tracking and audit logs.
	app.Use(middleware.Agent())

	// Sanitize — strips unsafe keys and values from incoming payloads.
	// Runs after parsing middleware, before handlers see the data.

	app.Use(middleware.Sanitize())
	// Request logging — noisy, so dev and staging only. Production logging
	// goes through applogger instead.
	if env.IsDev || env.IsStaging {
		app.Use(gin.Logger())
	}

	// Static assets — PublicStatic serves matching files from ./public and
	// aborts; non-matching requests fall through to the router. The Static
	// call adds the explicit /public prefix route.
	app.Use(middleware.PublicStatic("./public"))
	app.Static("/public", "./public")

	// Application routes (plain-function handlers, see internal/handlers).
	routes.RegisterRoutes(app)

	return app
}
