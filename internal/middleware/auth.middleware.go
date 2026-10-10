package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	commonauth "github.com/damolaoladipo/common-go/auth"
	pgroles "github.com/damolaoladipo/common-go/postgres"
	commontypes "github.com/damolaoladipo/common-go/types"

	"github.com/damolaoladipo/pricebase/internal/dtos"
	"github.com/damolaoladipo/pricebase/internal/repository"
	"github.com/damolaoladipo/pricebase/internal/services"
	"github.com/damolaoladipo/pricebase/internal/utils"
)

// SuperAdminRoles, Roles, OrgRoles, TeamRoles and AllRoles mirror the TS
// superadmin / roles / orgRoles / teamRoles / allRoles lists.
var SuperAdminRoles = []string{string(utils.UserTypeSuper)}
var Roles = []string{string(utils.UserTypeSuper), string(utils.UserTypeAdmin)}
var OrgRoles = []string{string(utils.UserTypeSuper), string(utils.UserTypeAdmin), string(utils.UserTypeOrganization)}
var TeamRoles = []string{string(utils.UserTypeSuper), string(utils.UserTypeAdmin), string(utils.UserTypeOrganization), string(utils.UserTypeTeam)}
var AllRoles = []string{string(utils.UserTypeSuper), string(utils.UserTypeAdmin), string(utils.UserTypeOrganization), string(utils.UserTypeTeam), string(utils.UserTypeUser)}

// Protect authenticates the bearer token and loads the caller.
// It mirrors the TS protect middleware: a decoded object payload goes through
// the session check then the user lookup; a string payload goes through the
// secret API-key lookup restricted to superadmin and organization callers.
// Failures abort through the ErrorHandler via utils.ErrorResponse, mirroring
// next(new ErrorResponse(...)).
func Protect(AuthService *services.AuthService, userRepo *repository.UserRepository) gin.HandlerFunc {

	return func(c *gin.Context) {

		authorization := c.GetHeader("Authorization")
		if authorization == "" || !strings.HasPrefix(authorization, "Bearer") {
			c.Error(utils.NewErrorResponse("Error!", http.StatusUnauthorized, []string{"user not authorized to access this route"}, nil))
			c.Abort()
			return
		}

		parts := strings.SplitN(authorization, " ", 2)
		if len(parts) != 2 || strings.TrimSpace(parts[1]) == "" {
			c.Error(utils.NewErrorResponse("unauthenticated!", http.StatusUnauthorized, []string{"user not authorized to access this route"}, nil))
			c.Abort()
			return
		}
		token := parts[1]

		env := utils.Current()
		secret := ""
		if env != nil {
			secret = env.JWT_SECRET
		}
		if secret == "" {
			c.Error(utils.NewErrorResponse("Error!", http.StatusInternalServerError, []string{"authentication is not configured"}, nil))
			c.Abort()
			return
		}

		verify := commonauth.Protect(token, secret)
		if verify.Error || verify.Data == nil {
			c.Error(utils.NewErrorResponse("unauthenticated!", http.StatusUnauthorized, []string{"user not authorized to access this route"}, nil))
			c.Abort()
			return
		}

		if claims, ok := verify.Data.(jwt.MapClaims); ok {
			id, _ := claims["id"].(string)
			jti, _ := claims["jti"].(string)

			sessionOk, err := AuthService.VerifySession(c.Request.Context(), dtos.VerifySessionDTO{JTI: jti, UserID: id})
			if err != nil {
				c.Error(err)
				c.Abort()
				return
			}
			if !sessionOk {
				c.Error(utils.NewErrorResponse("unauthenticated!", http.StatusUnauthorized, []string{"user not authorized to access this route"}, nil))
				c.Abort()
				return
			}

			user, err := userRepo.FindUserByID(c.Request.Context(), id)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					c.Error(utils.NewErrorResponse("unauthenticated!", http.StatusUnauthorized, []string{"user not authorized to access this route"}, nil))
					c.Abort()
					return
				}
				c.Error(err)
				c.Abort()
				return
			}

			if !user.IsActive {
				c.Error(utils.NewErrorResponse("Error!", http.StatusForbidden, []string{"account currently deactivated. please contact support"}, nil))
				c.Abort()
				return
			}

			c.Set("userType", user.UserType)
			permissions := make([]commontypes.UserPermission, 0, len(user.RolePermissions))
			for _, p := range user.RolePermissions {
				permissions = append(permissions, commontypes.UserPermission{Group: p.Group, Actions: p.Actions})
			}
			c.Set("user", dtos.AuthUser{
				ID:          user.ID,
				Email:       user.Email,
				UserType:    user.UserType,
				RoleIDs:     user.RoleIDs,
				Permissions: permissions,
			})
			c.Set("auth", dtos.AuthSession{Token: token, JTI: jti})
			c.Next()
			return
		}

		if key, ok := verify.Data.(string); ok {

			found, err := AuthService.DecodeAPIKey(c.Request.Context(), &dtos.DecodeAPIKeyDTO{APIKey: key, Type: dtos.APIKeyTypeSecret})
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					c.Error(utils.NewErrorResponse("unauthenticated!", http.StatusUnauthorized, []string{"user not authorized to access this route"}, nil))
					c.Abort()
					return
				}
				c.Error(err)
				c.Abort()
				return
			}

			if !found.IsActive {
				c.Error(utils.NewErrorResponse("Error!", http.StatusForbidden, []string{"account currently deactivated. please contact support"}, nil))
				c.Abort()
				return
			}

			if found.UserType != string(utils.UserTypeSuper) && found.UserType != string(utils.UserTypeOrganization) {
				c.Error(utils.NewErrorResponse("Forbidden!", http.StatusForbidden, []string{"user is not authorized to use api keys"}, nil))
				c.Abort()
				return
			}

			c.Set("userType", found.UserType)
			permissions := make([]commontypes.UserPermission, 0, len(found.RolePermissions))
			for _, p := range found.RolePermissions {
				permissions = append(permissions, commontypes.UserPermission{Group: p.Group, Actions: p.Actions})
			}
			c.Set("user", dtos.AuthUser{
				ID:          found.ID,
				Email:       found.Email,
				UserType:    found.UserType,
				RoleIDs:     found.RoleIDs,
				Permissions: permissions,
			})
			c.Set("auth", dtos.AuthSession{Token: token})
			c.Next()
			return
		}

		c.Error(utils.NewErrorResponse("unauthenticated!", http.StatusUnauthorized, []string{"user not authorized to access this route"}, nil))
		c.Abort()
	}
}

// Authorize gates a route by role, then optionally by permission.
// It mirrors the TS authorize middleware: the caller must come from Protect
// (the "user" context key), role membership is checked against the database,
// and a supplied dtos.PermitParams is enforced through the shared permit
// check. A nil perm skips the permission stage.
func Authorize(pool *pgxpool.Pool, roles []string, perm *dtos.PermitParams) gin.HandlerFunc {

	return func(c *gin.Context) {

		raw, ok := c.Get("user")
		if !ok {
			c.Error(utils.NewErrorResponse("unauthorized!", http.StatusUnauthorized, []string{"user is not signed in"}, nil))
			c.Abort()
			return
		}

		user, ok := raw.(dtos.AuthUser)
		if !ok {
			c.Error(utils.NewErrorResponse("unauthorized!", http.StatusUnauthorized, []string{"user is not signed in"}, nil))
			c.Abort()
			return
		}

		if !pgroles.Authorize(c.Request.Context(), roles, user.RoleIDs, pool) {
			c.Error(utils.NewErrorResponse("unauthorized!", http.StatusUnauthorized, []string{"user is not authorized to access this route"}, nil))
			c.Abort()
			return
		}

		if perm != nil {
			userPermissions := make([]commontypes.UserPermission, 0, len(user.Permissions))
			for _, p := range user.Permissions {
				userPermissions = append(userPermissions, commontypes.UserPermission{Group: p.Group, Actions: p.Actions})
			}

			check := commonauth.Permit(commontypes.PermitOptions{
				Group:           perm.Group,
				Actions:         perm.Actions,
				Roles:           perm.Roles,
				UserType:        user.UserType,
				UserPermissions: userPermissions,
			})
			if !check {
				c.Error(utils.NewErrorResponse("Not Permitted!", http.StatusForbidden, []string{"user is not permitted to access resource"}, nil))
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
