package dtos

import (
	"time"

	"github.com/damolaoladipo/common-go/types"
	"github.com/damolaoladipo/pricebase/internal/utils"
	"github.com/golang-jwt/jwt/v5"
)

type RegisterDTO struct {
}

type SeedRoleDTO struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type SeedPermDTO struct {
	Group   string                   `json:"group" binding:"required"`
	Actions []utils.PermissionAction `json:"actions" binding:"required"`
}

type SeedRolePermission struct {
	Group   string   `json:"group"`
	Actions []string `json:"actions"`
}

type SeedRolePermDTO struct {
	Role        string               `json:"role" binding:"required"`
	Permissions []SeedRolePermission `json:"permissions" binding:"required"`
}

type LoginDTO struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RegisterUserDTO struct {
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required"`
	UserType  string `json:"userType" binding:"required"`
}

// GenerateJWTDTO mirrors the TS type: the claims embedded in the signed
// session token by AuthService.GenerateJWT.
type GenerateJWTDTO struct {
	Email     string   `json:"email"`
	ID        string   `json:"id"`
	Roles     []string `json:"roles"`
	RoleNames []string `json:"roleNames"`
}

// VerifySessionDTO mirrors the TS type: the identifiers checked by
// AuthService.VerifySession against the cached session.
type VerifySessionDTO struct {
	JTI    string `json:"jti"`
	UserID string `json:"userId"`
}

// VerifyOTPDTO mirrors the TS verifyOtpDTO: the triple checked and consumed
// by UserService.VerifyOTP.
type VerifyOTPDTO struct {
	Email   string `json:"email" binding:"required,email"`
	Code    string `json:"code" binding:"required"`
	OTPType string `json:"otpType" binding:"required"`
}

type JWTClaims struct {
	jwt.RegisteredClaims
	Email     string   `json:"email"`
	ID        string   `json:"id"`
	Roles     []string `json:"roles"`
	RoleNames []string `json:"roleNames"`
}

// AuthCookieOptions mirrors the TS CookieOptions built in getAuthToken:
// 70-day expiry, readable by client scripts, plain HTTP unless production.
// MaxAge and Path map straight onto sessions.Options in the login handler:
//
//	session := sessions.Default(ctx)
//	session.Set("token", auth.Token)
//	session.Options(sessions.Options{
//		Path: auth.Cookie.Path, MaxAge: auth.Cookie.MaxAge,
//		HttpOnly: auth.Cookie.HTTPOnly, Secure: auth.Cookie.Secure,
//		SameSite: http.SameSiteNoneMode,
//	})
//	session.Save()
type AuthCookieOptions struct {
	Expires  time.Time `json:"expires"`
	MaxAge   int       `json:"maxAge"`
	Path     string    `json:"path"`
	HTTPOnly bool      `json:"httpOnly"`
	Secure   bool      `json:"secure"`
	SameSite string    `json:"sameSite"`
}

// AuthTokenResult is the getAuthToken payload: signed token, mapped login
// user, and the cookie options the handler applies to the response.
type AuthTokenResult struct {
	Token  string            `json:"token"`
	User   LoginUserResponse `json:"user"`
	Cookie AuthCookieOptions `json:"cookie"`
}

// AuthUser is the authenticated caller carried on the request context.
// It mirrors the TS req.user shape: identity, type, role references and the
// permission snapshot used by Authorize. Built by the protect middleware from
// the user row; RoleIDs feed the role check, Permissions the permit check.
type AuthUser struct {
	ID          string                 `json:"id"`
	Email       string                 `json:"email"`
	UserType    string                 `json:"userType"`
	RoleIDs     []string               `json:"roles"`
	Permissions []types.UserPermission `json:"permissions"`
}

// AuthSession is carried on the request context alongside AuthUser.
// It mirrors the TS req.auth shape: the bearer token and its jti.
// API-key callers have an empty JTI.
type AuthSession struct {
	Token string `json:"token"`
	JTI   string `json:"jti,omitempty"`
}

// PermitParams mirrors the TS IPermit shape: the optional permission stage
// of Authorize. A nil *PermitParams skips the permission check.
type PermitParams struct {
	Group   string
	Actions []string
	Roles   []string
}
