package services

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	commonutils "github.com/damolaoladipo/common-go/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/damolaoladipo/pricebase/internal/common/redis"
	"github.com/damolaoladipo/pricebase/internal/dtos"
	"github.com/damolaoladipo/pricebase/internal/mappers"
	"github.com/damolaoladipo/pricebase/internal/models"
	"github.com/damolaoladipo/pricebase/internal/repository"

	"github.com/damolaoladipo/pricebase/internal/utils"
)

// AuthService mirrors FieldOS auth.service.ts (extractActions,
// validateActions, extractPermissions).
//
// The role is loaded by ID, its permissions are loaded with actions, and
// names are extracted in memory via ExtractActions.
type AuthService struct {
	roles       *repository.RoleRepository
	permissions *repository.PermissionRepository
	user        *repository.UserRepository
	userMapper  *mappers.UserMapper
}

func NewAuthService(roles *repository.RoleRepository, permissions *repository.PermissionRepository, user *repository.UserRepository, userMapper *mappers.UserMapper) *AuthService {
	return &AuthService{roles: roles, permissions: permissions, user: user, userMapper: userMapper}
}

// ExtractActions returns action names from permission actions.
func (s *AuthService) ExtractActions(data []utils.PermissionAction) []string {

	result := make([]string, 0, len(data))
	for _, action := range data {
		result = append(result, action.Name)
	}
	return result
}

// ValidateActions returns true when every requested action exists in available
func (s *AuthService) ValidateActions(available []string, requested []string) bool {

	set := make(map[string]struct{}, len(available))
	for _, name := range available {
		set[name] = struct{}{}
	}

	for _, name := range requested {
		if _, ok := set[name]; !ok {
			return false
		}
	}

	return true
}

// ExtractPermissions builds the user-permission shape for a role.
// It mirrors extractPermissions: the role is loaded by ID, a missing role
// yields an empty result (like if (role)), and action names come from
// ExtractActions over each permission's actions.
func (s *AuthService) ExtractPermissions(ctx context.Context, id string) ([]utils.UserPermission, error) {

	result := make([]utils.UserPermission, 0)

	if s.roles == nil || s.permissions == nil {
		return result, nil
	}

	permissions, err := s.permissions.FindByRoleID(ctx, id)
	if err != nil {
		return nil, err
	}

	for _, permission := range permissions {
		result = append(result, utils.UserPermission{
			Group:   permission.Group,
			Actions: s.ExtractActions(permission.Actions),
		})
	}

	return result, nil
}

// JWTClaims are the session claims embedded in the signed token.
// RegisteredClaims.ID carries the jti; ExpiresAt is set from JWT_EXPIRES.

// GenerateJWT signs a session token and caches the session.
// Returns the signed token and the jti.
func (s *AuthService) GenerateJWT(ctx context.Context, data dtos.GenerateJWTDTO) (string, string, error) {

	env := utils.Current()
	secret := ""
	if env != nil {
		secret = env.JWT_SECRET
	}
	if secret == "" {
		return "", "", errors.New("JWT secret is not defined")
	}

	expiry := 24 * time.Hour
	if env != nil && env.JWT_EXPIRES != "" {
		dur, err := time.ParseDuration(env.JWT_EXPIRES)
		if err != nil {
			return "", "", fmt.Errorf("invalid JWT_EXPIRES %q: %w", env.JWT_EXPIRES, err)
		}
		expiry = dur
	}

	jti := uuid.NewString()
	now := time.Now()
	claims := dtos.JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
		},
		Email:     data.Email,
		ID:        data.ID,
		Roles:     data.Roles,
		RoleNames: data.RoleNames,
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign token: %w", err)
	}

	jData := redis.Data{
		Key:   utils.CacheKey(fmt.Sprintf("%s:%s", utils.CacheKeys.SessionJTI, jti)),
		Value: map[string]any{"userId": data.ID},
	}
	if err := redis.REDIS.KeepData(ctx, jData, 86400); err != nil {
		return "", "", fmt.Errorf("failed to cache session: %w", err)
	}

	uData := redis.Data{
		Key:   utils.CacheKey(fmt.Sprintf("%s:%s", utils.CacheKeys.UserSession, data.ID)),
		Value: map[string]any{"jti": jti},
	}
	if err := redis.REDIS.SKeep(ctx, uData); err != nil {
		return "", "", fmt.Errorf("failed to index user session: %w", err)
	}

	return signed, jti, nil
}

// VerifySession returns true when the cached jti maps to userID.
func (s *AuthService) VerifySession(ctx context.Context, data dtos.VerifySessionDTO) (bool, error) {

	cached, err := redis.REDIS.FetchData(ctx, utils.CacheKey(fmt.Sprintf("%s:%s", utils.CacheKeys.SessionJTI, data.JTI)))
	if err != nil {
		return false, err
	}
	if cached == nil {
		return false, nil
	}

	entry, ok := cached.(map[string]any)
	if !ok {
		return false, nil
	}
	cachedUserID, _ := entry["userId"].(string)
	if cachedUserID == "" {
		return false, nil
	}

	return cachedUserID == data.UserID, nil
}

// RevokeSession deletes one session: the jti key plus its member in the
// user-session set. A missing key or a cached entry without userId is a
// no-op, mirroring the TS guards.
func (s *AuthService) RevokeSession(ctx context.Context, jti string) error {

	key := utils.CacheKey(fmt.Sprintf("%s:%s", utils.CacheKeys.SessionJTI, jti))
	cached, err := redis.REDIS.FetchData(ctx, key)
	if err != nil {
		return err
	}
	if cached == nil {
		return nil
	}

	entry, ok := cached.(map[string]any)
	if !ok {
		return nil
	}
	cachedUserID, _ := entry["userId"].(string)
	if cachedUserID == "" {
		return nil
	}

	if err := redis.REDIS.DeleteData(ctx, key); err != nil {
		return err
	}

	return redis.REDIS.SDelete(ctx, utils.CacheKey(fmt.Sprintf("%s:%s", utils.CacheKeys.UserSession, cachedUserID)), map[string]any{"jti": jti})
}

// RevokeAllSessions deletes every session of a user: each jti key found in
// the user-session set, then the set key itself.
func (s *AuthService) RevokeAllSessions(ctx context.Context, userID string) error {

	key := utils.CacheKey(fmt.Sprintf("%s:%s", utils.CacheKeys.UserSession, userID))
	members, err := redis.REDIS.FetchMembers(ctx, key)
	if err != nil {
		return err
	}

	for _, member := range members {
		entry, ok := member.(map[string]any)
		if !ok {
			continue
		}
		jti, _ := entry["jti"].(string)
		if jti == "" {
			continue
		}
		if err := redis.REDIS.DeleteData(ctx, utils.CacheKey(fmt.Sprintf("%s:%s", utils.CacheKeys.SessionJTI, jti))); err != nil {
			return err
		}
	}

	return redis.REDIS.DeleteData(ctx, key)
}

// GetActivationToken creates an activation token for a user.
// It mirrors getActivationToken: a missing user yields an empty token and no
// error (TS null). The plaintext token is returned for email while its sha256
// hash and expiry persist through the repository.
func (s *AuthService) GetActivationToken(ctx context.Context, id string) (string, error) {

	_, err := s.user.FindUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", err
	}

	raw := make([]byte, 20)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	plain := hex.EncodeToString(raw)
	sum := sha256.Sum256([]byte(plain))
	hash := hex.EncodeToString(sum[:])
	expiry := time.Now().Add(10 * time.Minute)

	if err := s.user.UpdateActivationToken(ctx, id, hash, &expiry); err != nil {
		return "", err
	}

	return plain, nil
}

// GetResetPasswordToken creates a reset-password token for a user.
// A missing user yields an empty token and no error. The plaintext token is
// returned for email while its sha256 hash and 10-minute expiry persist.
func (s *AuthService) GetResetPasswordToken(ctx context.Context, id string) (string, error) {

	_, err := s.user.FindUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", err
	}

	raw := make([]byte, 20)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	plain := hex.EncodeToString(raw)
	sum := sha256.Sum256([]byte(plain))
	hash := hex.EncodeToString(sum[:])
	expiry := time.Now().Add(10 * time.Minute)

	if err := s.user.UpdateResetToken(ctx, id, hash, &expiry); err != nil {
		return "", err
	}

	return plain, nil
}

// GetInviteToken creates an invite token for a user.
// A missing user yields an empty token and no error. The plaintext token is
// returned for email while its sha256 hash and 24-hour expiry persist.
func (s *AuthService) GetInviteToken(ctx context.Context, id string) (string, error) {

	_, err := s.user.FindUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", err
	}

	raw := make([]byte, 20)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	plain := hex.EncodeToString(raw)
	sum := sha256.Sum256([]byte(plain))
	hash := hex.EncodeToString(sum[:])
	expiry := time.Now().Add(24 * time.Hour)

	if err := s.user.UpdateInviteToken(ctx, id, hash, &expiry); err != nil {
		return "", err
	}

	return plain, nil
}

// GetSignedJwtToken signs a JWT carrying the user's id, email, and roles.
// It mirrors the TS getSignedJwtToken explicitly with stdlib HMAC-SHA256
// (no JWT dependency): base64url(header).base64url(payload).base64url(sig).
// secret is the HMAC key (ENV.JWT_SECRET); expiry is a Go duration string
// like "1h" (ENV.JWT_EXPIRES).
func (s *AuthService) GetSignedJwtToken(user *models.User, secret, expiry string) (string, error) {
	if secret == "" {
		return "", fmt.Errorf("secret is required")
	}
	ttl, err := time.ParseDuration(expiry)
	if err != nil {
		return "", fmt.Errorf("invalid expiry %q: %w", expiry, err)
	}

	now := time.Now()
	roles := make([]string, 0, len(user.RoleIDs))
	for _, id := range user.RoleIDs {
		roles = append(roles, id)
	}

	payloadJSON, err := commonutils.ToJSON(map[string]any{
		"id":    user.ID,
		"email": user.Email,
		"roles": roles,
		"iat":   now.Unix(),
		"exp":   now.Add(ttl).Unix(),
	})
	if err != nil {
		return "", fmt.Errorf("failed to encode claims: %w", err)
	}

	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString(payloadJSON)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(header + "." + payload))

	return header + "." + payload + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

// GetAuthToken builds the login payload for a user.
// It mirrors getAuthToken: roles are loaded for the token claims, a JWT is
// signed via GenerateJWT, cookie options default to insecure with secure
// flipped on in production, and the user is shaped by UserMapper.MapLoginUser.
func (s *AuthService) GetAuthToken(ctx context.Context, userID string) (*dtos.AuthTokenResult, error) {

	authUser, err := s.user.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	roleIDs := make([]string, 0, len(authUser.RoleIDs))
	roleNames := make([]string, 0, len(authUser.RoleIDs))
	roles := make([]models.Role, 0, len(authUser.RoleIDs))
	for _, roleID := range authUser.RoleIDs {
		role, err := s.roles.FindByID(ctx, roleID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			return nil, err
		}
		roleIDs = append(roleIDs, role.ID)
		roleNames = append(roleNames, role.Name)
		roles = append(roles, role)
	}

	signed, _, err := s.GenerateJWT(ctx, dtos.GenerateJWTDTO{
		Email:     authUser.Email,
		ID:        authUser.ID,
		Roles:     roleIDs,
		RoleNames: roleNames,
	})
	if err != nil {
		return nil, err
	}

	secure := false
	if env := utils.Current(); env != nil && env.IsProd {
		secure = true
	}

	mapped := s.userMapper.MapLoginUser(&authUser, roles, nil)

	return &dtos.AuthTokenResult{
		Token: signed,
		User:  mapped,
		Cookie: dtos.AuthCookieOptions{
			Expires:  time.Now().Add(70 * 24 * time.Hour),
			MaxAge:   70 * 24 * 60 * 60,
			Path:     "/",
			HTTPOnly: false,
			Secure:   secure,
			SameSite: "none",
		},
	}, nil
}

// GenerateAPIKey
// GenerateAPIKey generates and stores API keys for the user.
//
// It generates a secret and public key, combines them with the domain,
// and creates a SHA-256 hash for each. The result is an APIKey struct
// with the combined key, tokens, and domain. The APIKey is then added
// to the user's keys slice and persisted to the database.
//
// Returns the updated user and any errors that occur.
func (s *AuthService) GenerateAPIKey(ctx context.Context, user *models.User) (*models.User, error) {

	// Generate key secrets.
	secretBytes := make([]byte, 25)
	if _, err := rand.Read(secretBytes); err != nil {
		return nil, err
	}

	pubkeyBytes := make([]byte, 25)
	if _, err := rand.Read(pubkeyBytes); err != nil {
		return nil, err
	}

	secret := hex.EncodeToString(secretBytes)
	pubkey := hex.EncodeToString(pubkeyBytes)

	// Configure domain.
	domain := utils.DomainTest

	// check if the environment is production
	//ENV
	if current := utils.Current(); current != nil && current.IsProd {
		domain = utils.DomainLive
	}

	// Define keys.
	secretCombined := fmt.Sprintf("%s_%s_%s", utils.SECRETKEY, domain, secret)
	publicCombined := fmt.Sprintf("%s_%s_%s", utils.PUBLICKEY, domain, pubkey)

	// Generate tokens.
	secretToken := sha256.Sum256([]byte(secretCombined))
	publicToken := sha256.Sum256([]byte(publicCombined))

	apiKey := utils.APIKey{
		Secret:      secretCombined,
		Public:      publicCombined,
		SecretToken: hex.EncodeToString(secretToken[:]),
		PublicToken: hex.EncodeToString(publicToken[:]),
		Domain:      string(domain),
		IsActive:    true,
		UpdatedAt:   time.Now(),
	}

	// Update in-memory user.
	user.APIKey = apiKey
	user.Keys = append(user.Keys, apiKey)

	// Persist through PostgreSQL repository.
	if err := s.user.UpdateAPIKeys(
		ctx,
		user.ID,
		apiKey,
		user.Keys,
	); err != nil {
		return nil, err
	}

	return user, nil
}

// DecodeAPIKey
//
// finds the user associated with an API key.
//
// It creates a SHA-256 hash of the API key and token, and then searches
// the database for the user associated with the key and token.
//
// Returns the user and any errors that occur.
func (s *AuthService) DecodeAPIKey(ctx context.Context, data *dtos.DecodeAPIKeyDTO) (*models.User, error) {

	token := sha256.Sum256([]byte(data.APIKey))
	tokenHash := hex.EncodeToString(token[:])

	switch data.Type {
	case "secret":
		user, err := s.user.FindBySecretAPIKey(
			ctx,
			data.APIKey,
			tokenHash,
		)
		if err != nil {
			return nil, err
		}

		return &user, nil

	case "public":
		user, err := s.user.FindByPublicAPIKey(
			ctx,
			data.APIKey,
			tokenHash,
		)
		if err != nil {
			return nil, err
		}

		return &user, nil

	default:
		return nil, errors.New("invalid API key type")
	}
}

// package services

// import (
// 	"context"

// 	"github.com/damolaoladipo/pricebase/internal/dtos"
// 	"github.com/damolaoladipo/pricebase/internal/models"
// 	"github.com/damolaoladipo/pricebase/internal/repositories"
// 	"github.com/damolaoladipo/pricebase/internal/utils"
// )

// type AuthService struct {
// 	repo *repositories.UserRepository
// }

// func NewAuthService(repo *repositories.UserRepository) *AuthService {
// 	return &AuthService{
// 		repo: repo,
// 	}
// }

// func (s *AuthService) Register(ctx context.Context, data *dtos.RegisterUserDTO) utils.Result[models.User] {
// 	return utils.Result[models.User]{}
// }

// func (s *AuthService) ValidateRegister(data *dtos.RegisterUserDTO) utils.Result[models.User] {
// 	allowedUsers := []models.UserType{
// 		models.UserTypeListener,
// 		models.UserTypeCreator,
// 		models.UserTypeMinister,
// 	}

// 	result := Result{
// 		Error:   false,
// 		Message: "",
// 		Code:    200,
// 		Data:    map[string]any{},
// 	}

// 	switch {
// 	case strings.TrimSpace(data.Email) == "":
// 		result.Error = true
// 		result.Message = "Email is required"

// 	case strings.TrimSpace(data.FirstName) == "":
// 		result.Error = true
// 		result.Message = "First name is required"

// 	case strings.TrimSpace(data.LastName) == "":
// 		result.Error = true
// 		result.Message = "Last name is required"

// 	case data.Password == "":
// 		result.Error = true
// 		result.Message = "Password is required"

// 	case !containsUserType(allowedUsers, data.UserType):
// 		result.Error = true
// 		result.Message = fmt.Sprintf(
// 			"Invalid user type value. choose from %s",
// 			formatUserTypes(allowedUsers),
// 		)
// 	}

// 	return result
// }

// func (s *AuthService) ValidateLogin(data LoginDTO) Result {
// 	result := Result{
// 		Error:   false,
// 		Message: "",
// 		Code:    200,
// 		Data:    nil,
// 	}

// 	if strings.TrimSpace(data.Email) == "" {
// 		result.Error = true
// 		result.Message = "email is required"
// 		return result
// 	}

// 	if data.Password == "" {
// 		result.Error = true
// 		result.Message = "password is required"
// 		return result
// 	}

// 	if !s.CheckEmail(data.Email) {
// 		result.Error = true
// 		result.Message = "a valid email is required"
// 		return result
// 	}

// 	return result
// }

// func (s *AuthService) ValidatePhoneNumber(phone string) bool {
// 	if len(phone) < 3 {
// 		return false
// 	}

// 	if phone[0] != '0' {
// 		return false
// 	}

// 	return isDigit(phone[1]) && isDigit(phone[2])
// }

// func isDigit(value byte) bool {
// 	return value >= '0' && value <= '9'
// }

// func (s *AuthService) CheckEmail(email string) bool {
// 	email = strings.TrimSpace(email)

// 	if strings.Contains(email, ".africa") {
// 		return true
// 	}

// 	_, err := mail.ParseAddress(email)

// 	return err == nil
// }

// func (s *AuthService) CheckPassword(password string) bool {
// 	if len(password) < 8 {
// 		return false
// 	}

// 	var (
// 		hasUpper   bool
// 		hasLower   bool
// 		hasNumber  bool
// 		hasSpecial bool
// 	)

// 	for _, char := range password {
// 		switch {
// 		case char >= 'A' && char <= 'Z':
// 			hasUpper = true

// 		case char >= 'a' && char <= 'z':
// 			hasLower = true

// 		case char >= '0' && char <= '9':
// 			hasNumber = true

// 		default:
// 			hasSpecial = true
// 		}
// 	}

// 	return hasUpper &&
// 		hasLower &&
// 		hasNumber &&
// 		hasSpecial
// }

// func (s *AuthService) AttachPhoneCode(code, phone string) string {
// 	if code == "" || phone == "" {
// 		return ""
// 	}

// 	codeStr := code

// 	if strings.Contains(code, "-") {
// 		if len(code) >= 3 {
// 			codeStr = "+" + code[3:]
// 		}
// 	}

// 	return codeStr + phone[1:]
// }

// func (s *AuthService) CheckPhoneCode(code, phone string) string {
// 	if code == "" || phone == "" {
// 		return ""
// 	}

// 	if !strings.Contains(phone, "+") && len(phone) > 10 {
// 		return code + phone[3:]
// 	}

// 	if strings.Contains(phone, "+") {
// 		return phone
// 	}

// 	return ""
// }

// func (s *AuthService) PhoneExists(
// 	ctx context.Context,
// 	phone string,
// ) (bool, error) {
// 	user, err := s.UserRepository.FindByPhoneOrAltPhone(ctx, phone)

// 	if err != nil {
// 		return false, err
// 	}

// 	return user != nil, nil
// }

// func (r *UserRepository) UpdateUserType(
// 	ctx context.Context,
// 	userID primitive.ObjectID,
// 	userType models.UserType,
// ) error {
// 	_, err := r.Collection.UpdateOne(
// 		ctx,
// 		bson.M{"_id": userID},
// 		bson.M{
// 			"$set": bson.M{
// 				"userType": userType,
// 			},
// 		},
// 	)

// 	return err
// }
// func (s *AuthService) UpdateLastLogin(
// 	ctx context.Context,
// 	user *models.User,
// ) error {
// 	now := time.Now()

// 	method := user.Login.Method

// 	if method == "" {
// 		method = models.LoginMethodEmail
// 	}

// 	return s.UserRepository.UpdateLogin(
// 		ctx,
// 		user.ID,
// 		models.Login{
// 			Last:   now,
// 			Method: method,
// 		},
// 	)
// }

// func (s *AuthService) UpdateLoginInfo(
// 	ctx context.Context,
// 	user *models.User,
// 	c *gin.Context,
// ) error {
// 	userAgent := c.GetHeader("User-Agent")

// 	// Parse userAgent here if you actually need
// 	// device, OS and browser information.

// 	method := user.Login.Method

// 	if method == "" {
// 		method = models.LoginMethodEmail
// 	}

// 	return s.UserRepository.UpdateLogin(
// 		ctx,
// 		user.ID,
// 		models.Login{
// 			Last:   time.Now(),
// 			Method: method,
// 		},
// 	)
// }

// func (s *AuthService) GenerateOTPCode(
// 	ctx context.Context,
// 	user *models.User,
// 	otpType models.OTPType,
// ) (string, error) {
// 	code, err := generateOTP(6)

// 	if err != nil {
// 		return "", err
// 	}

// 	expiry := time.Now().Add(15 * time.Minute)

// 	err = s.UserRepository.UpdateOTP(
// 		ctx,
// 		user.ID,
// 		code,
// 		expiry,
// 		otpType,
// 	)

// 	if err != nil {
// 		return "", err
// 	}

// 	return code, nil
// }
