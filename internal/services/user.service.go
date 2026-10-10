package services

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"
	"unicode"

	"golang.org/x/crypto/bcrypt"

	"github.com/damolaoladipo/pricebase/internal/dtos"
	"github.com/damolaoladipo/pricebase/internal/mappers"
	"github.com/damolaoladipo/pricebase/internal/models"
	"github.com/damolaoladipo/pricebase/internal/repository"
	"github.com/damolaoladipo/pricebase/internal/utils"
	"github.com/jackc/pgx/v5"
)

type UserService struct {
	repo   *repository.UserRepository
	roles  *repository.RoleRepository
	system *SystemService
}

func NewUserService(repo *repository.UserRepository, system *SystemService, roles *repository.RoleRepository) *UserService {
	return &UserService{
		repo:   repo,
		system: system,
		roles:  roles,
	}
}

func (s *UserService) CreateUser(ctx context.Context, data *dtos.CreateUserDTO) utils.Result[models.User] {

	result := utils.Result[models.User]{
		Error:   false,
		Message: "",
		Code:    200,
		Data:    models.User{},
	}

	return result
}

func (s *UserService) EncryptUserPassword(ctx context.Context, user *models.User, password string) (bool, error) {
	
	encrypted, err := s.system.EncryptData(dtos.EncryptDataDTO{
		Payload:   password,
		Password:  user.Email,
		Separator: "-",
	})

	if err != nil {
		return false, err
	}

	if encrypted == "" {
		return false, nil
	}

	user.SavedPassword = encrypted

	if err := s.repo.UpdateSavedPassword(ctx, user.ID, encrypted); err != nil {
		return false, err
	}

	return true, nil
}

func (s *UserService) DecryptUserPassword(user *models.User) (string, error) {
	
	decrypted, err := s.system.DecryptData(dtos.DecryptDataDTO{
		Password:  user.Email,
		Payload:   user.SavedPassword,
		Separator: "-",
	})

	if err != nil {
		return "", err
	}

	if decrypted == nil {
		return "", nil
	}

	return fmt.Sprint(decrypted), nil
}

func (s *UserService) MatchEncryptedPassword(user *models.User, hash string) (bool, error) {
	
	if user.SavedPassword != hash {
		return false, nil
	}

	// Decrypt encrypted password.
	decrypted, err := s.DecryptUserPassword(user)
	if err != nil {
		return false, err
	}

	// Decrypt supplied hash.
	hashDecrypt, err := s.system.DecryptData(dtos.DecryptDataDTO{
		Password:  user.Email,
		Payload:   hash,
		Separator: "-",
	})

	if err != nil {
		return false, err
	}

	if hashDecrypt == nil {
		return false, nil
	}

	return decrypted == fmt.Sprint(hashDecrypt), nil
}

// ValidateUser checks the enum-constrained fields, mirroring the TS schema
// enum: entries. Empty values are skipped so partial updates pass.
func (s *UserService) ValidateUser(user *models.User) error {
	
	if user.UserType != "" && !utils.UserType(user.UserType).IsValid() {
		return fmt.Errorf("invalid userType %q, choose from %v", user.UserType, utils.UserTypeValues())
	}
	if user.OTPType != "" && !utils.OTPType(user.OTPType).IsValid() {
		return fmt.Errorf("invalid otpType %q, choose from %v", user.OTPType, utils.OTPTypeValues())
	}
	if user.Login.Method != "" && !utils.LoginMethod(user.Login.Method).IsValid() {
		return fmt.Errorf("invalid login method %q, choose from %v", user.Login.Method, utils.LoginMethodValues())
	}
	if user.Register.Method != "" && !utils.RegisterMethod(user.Register.Method).IsValid() {
		return fmt.Errorf("invalid register method %q, choose from %v", user.Register.Method, utils.RegisterMethodValues())
	}
	return nil
}

// HashPassword bcrypt-hashes user.Password in place.
// It mirrors the TS pre-save hook; call it at registration before CreateUser
// (there are no ORM hooks in Go).
func (s *UserService) HashPassword(user *models.User) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.Password = string(hash)
	return nil
}

// MatchPassword compares a plaintext password against the stored hash.
func (s *UserService) MatchPassword(user *models.User, password string) bool {
	if user.Password == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) == nil
}

// IncreaseLoginLimit increments the attempt counter, locks the account for
// 30 minutes at 5 attempts, and persists via UpdateLoginLimit.
// Returns the new attempt count.
func (s *UserService) IncreaseLoginLimit(ctx context.Context, user *models.User) (int, error) {

	user.LoginLimit++

	if user.LoginLimit >= 5 {
		user.IsLocked = true
		lockedUntil := time.Now().Add(30 * time.Minute)
		user.LockedUntil = &lockedUntil
	}

	if err := s.repo.UpdateLoginLimit(ctx, user.ID, user.LoginLimit, user.IsLocked, user.LockedUntil); err != nil {
		return 0, err
	}

	return user.LoginLimit, nil
}

// CheckLockedStatus reports whether the account is locked.
// A past lock expiry auto-unlocks: flags reset in memory and persisted,
// mirroring the TS check. Returns the lock state.
func (s *UserService) CheckLockedStatus(ctx context.Context, user *models.User) (bool, error) {
	
	if !user.IsLocked || user.LockedUntil == nil {
		return false, nil
	}

	if time.Now().After(*user.LockedUntil) {
		user.IsLocked = false
		user.LockedUntil = nil
		user.LoginLimit = 0
		if err := s.repo.UpdateAccountStatus(ctx, user.ID, user.IsActivated, user.IsActive, false, 0); err != nil {
			return false, err
		}
		return false, nil
	}

	return true, nil
}

// HasRole reports whether roleID is in the user's stored role snapshot.
func (s *UserService) HasRole(user *models.User, roleID string) bool {
	
	for _, id := range user.RoleIDs {
		if id == roleID {
			return true
		}
	}
	return false
}

// PhoneExists reports whether any user holds the number as primary or alt.
func (s *UserService) PhoneExists(ctx context.Context, phone string) (bool, error) {
	
	_, err := s.repo.FindByPhone(ctx, phone)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// UpdateUserType assigns a type in memory and persists it.
func (s *UserService) UpdateUserType(ctx context.Context, user *models.User, userType string) error {
	user.UserType = userType
	return s.repo.UpdateUserType(ctx, user.ID, userType)
}

// UpdateLastLogin stamps the login time and default method, then persists.
func (s *UserService) UpdateLastLogin(ctx context.Context, user *models.User) error {
	
	user.Login.Last = time.Now().UTC().Format(time.RFC3339)
	if user.Login.Method == "" {
		user.Login.Method = string(utils.LoginMethodEmail)
	}
	return s.repo.UpdateLogin(ctx, user.ID, user.Login)
}

// UpdateLoginInfo stamps the login time and default method, then persists.
// The TS version also parses the request user-agent; device tracking has no
// repository in Go, so only the login stamp is recorded here.
func (s *UserService) UpdateLoginInfo(ctx context.Context, user *models.User) error {
	
	user.Login.Last = time.Now().UTC().Format(time.RFC3339)
	if user.Login.Method == "" {
		user.Login.Method = string(utils.LoginMethodEmail)
	}
	return s.repo.UpdateLogin(ctx, user.ID, user.Login)
}

// ActivateAccount enables the account and clears lock state, then persists.
func (s *UserService) ActivateAccount(ctx context.Context, user *models.User) error {

	user.IsActivated = true
	user.IsActive = true
	user.IsLocked = false
	user.LoginLimit = 0
	return s.repo.UpdateAccountStatus(ctx, user.ID, true, true, false, 0)

}

// DeactivateAccount disables and locks the account, then persists.
// The TS version also sets isDeactivated, which has no model counterpart;
// IsLocked carries the locked state instead.
func (s *UserService) DeactivateAccount(ctx context.Context, user *models.User) error {
	
	user.IsActive = false
	user.IsActivated = false
	user.IsLocked = true
	return s.repo.UpdateAccountStatus(ctx, user.ID, false, false, true, user.LoginLimit)
}

// GenerateOTPCode creates a 6-digit code with 15-minute expiry and persists it.
// Returns the plaintext code for delivery.
func (s *UserService) GenerateOTPCode(ctx context.Context, user *models.User, otpType string) (string, error) {
	
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "", fmt.Errorf("failed to generate otp: %w", err)
	}
	code := fmt.Sprintf("%06d", n.Int64()+100000)
	expiry := time.Now().Add(15 * time.Minute)

	user.OTPCode = &code
	user.OTPCodeExpire = &expiry
	user.OTPType = otpType

	if err := s.repo.UpdateOTP(ctx, user.ID, code, &expiry, otpType); err != nil {
		return "", err
	}

	return code, nil
}

// VerifyOTPCode finds the holder of a live code.
// A missing or expired code yields nil with no error, mirroring TS null.
func (s *UserService) VerifyOTPCode(ctx context.Context, code string) (*models.User, error) {
	
	user, err := s.repo.FindByOTPCode(ctx, code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// VerifyOTP checks an email/code/type triple and consumes the code.
// An unknown triple returns 400 Invalid OTP code; an expired code is cleared
// first and returns 400 OTP expired. A valid code returns the user.
func (s *UserService) VerifyOTP(ctx context.Context, data dtos.VerifyOTPDTO) (utils.Result[any], error) {
	
	result := utils.Result[any]{Error: false, Message: "", Code: 200, Data: map[string]any{}}

	user, err := s.repo.FindByEmailOTP(ctx, data.Email, data.Code, data.OTPType)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			result.Error = true
			result.Message = "Invalid OTP code"
			result.Code = 400
			return result, nil
		}
		return result, err
	}

	if user.OTPCodeExpire == nil || time.Now().After(*user.OTPCodeExpire) {
		if err := s.repo.UpdateOTP(ctx, user.ID, "", nil, user.OTPType); err != nil {
			return result, err
		}
		result.Error = true
		result.Message = "OTP has expired. Please request a new one"
		result.Code = 400
		return result, nil
	}

	if err := s.repo.UpdateOTP(ctx, user.ID, "", nil, user.OTPType); err != nil {
		return result, err
	}

	result.Data = user
	result.Message = "OTP verified successfully"
	return result, nil
}

// GetLoggedInUser loads the caller by ID for authenticated routes.
// A missing user is a 401 error, mirroring the TS unauthorized result.
func (s *UserService) GetLoggedInUser(ctx context.Context, userID string) (dtos.LoginUserResponse, error) {
	
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dtos.LoginUserResponse{}, utils.NewErrorResponse("authorized - user details not found", 401, []string{"authorized - user details not found"}, nil)
		}
		return dtos.LoginUserResponse{}, err
	}

	roles := make([]models.Role, 0, len(user.RoleIDs))
	for _, roleID := range user.RoleIDs {
		role, err := s.roles.FindByID(ctx, roleID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			return dtos.LoginUserResponse{}, err
		}
		roles = append(roles, role)
	}

	return mappers.NewUserMapper().MapLoginUser(&user, roles, nil), nil
}

// AttachRole links a role by name and records it on the user snapshot.
// A missing role is a 400 error, mirroring the TS throw.
func (s *UserService) AttachRole(ctx context.Context, user *models.User, roleName string) error {
	role, err := s.roles.FindByName(ctx, roleName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return utils.NewErrorResponse(fmt.Sprintf("Role %s does not exist.", roleName), 400, []string{fmt.Sprintf("Role %s does not exist.", roleName)}, nil)
		}
		return err
	}

	if _, err := s.repo.AttachRole(ctx, user.ID, role.ID); err != nil {
		return err
	}

	roleIDs := user.RoleIDs
	found := false
	for _, id := range roleIDs {
		if id == role.ID {
			found = true
		}
	}
	if !found {
		roleIDs = append(roleIDs, role.ID)
	}
	user.RoleIDs = roleIDs

	return s.repo.UpdateRolesPermissions(ctx, user.ID, roleIDs, user.RolePermissions)
}

// emailPattern mirrors the TS checkEmail regex.
var emailPattern = regexp.MustCompile(`^\w+([\.-]?\w+)*@\w+([\.-]?\w+)*(\.\w{2,3})+$`)

// ValidateRegister validates the registration payload before user creation.
// Only non-privileged types may self-register, mirroring the TS listener /
// creator / minister allowlist.
func (s *UserService) ValidateRegister(data dtos.RegisterUserDTO) utils.Result[any] {
	
	allowed := []string{
		string(utils.UserTypeOrganization),
		string(utils.UserTypeTeam),
		string(utils.UserTypeUser),
	}

	result := utils.Result[any]{Error: false, Message: "", Code: 200, Data: map[string]any{}}

	switch {
	case strings.TrimSpace(data.Email) == "":
		result.Error = true
		result.Message = "Email is required"
	case strings.TrimSpace(data.FirstName) == "":
		result.Error = true
		result.Message = "First name is required"
	case strings.TrimSpace(data.LastName) == "":
		result.Error = true
		result.Message = "Last name is required"
	case data.Password == "":
		result.Error = true
		result.Message = "Password is required"
	case !s.ValidateUserType(data.UserType, allowed):
		result.Error = true
		result.Message = fmt.Sprintf("Invalid user type value. choose from %s", strings.Join(allowed, ","))
	}

	return result
}

// ValidateLogin validates the login payload: required fields plus email format.
func (s *UserService) ValidateLogin(data dtos.LoginDTO) utils.Result[any] {
	
	result := utils.Result[any]{Error: false, Message: "", Code: 200}

	switch {
	case strings.TrimSpace(data.Email) == "":
		result.Error = true
		result.Message = "email is required"
	case data.Password == "":
		result.Error = true
		result.Message = "password is required"
	case !s.CheckEmail(data.Email):
		result.Error = true
		result.Message = "a valid email is required"
	}

	return result
}

// ValidatePhoneNumber mirrors the TS check: 0-leading with two digits after.
func (s *UserService) ValidatePhoneNumber(phone string) bool {
	
	if len(phone) < 3 {
		return false
	}
	if phone[0] != '0' {
		return false
	}
	return isDigit(phone[1]) && isDigit(phone[2])
}

func isDigit(value byte) bool {
	return value >= '0' && value <= '9'
}

// CheckEmail validates the format, bypassing .africa domains like the TS version.
func (s *UserService) CheckEmail(email string) bool {
	
	if strings.Contains(email, ".africa") {
		return true
	}
	return emailPattern.MatchString(email)
}

// CheckPassword requires 8+ chars with upper, lower, digit and special.
// Go's regexp has no lookahead, so the TS lookahead pattern is checked manually.
func (s *UserService) CheckPassword(password string) bool {
	
	if len(password) < 8 {
		return false
	}

	var upper, lower, digit, special bool
	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			upper = true
		case unicode.IsLower(c):
			lower = true
		case unicode.IsDigit(c):
			digit = true
		case !unicode.IsSpace(c):
			special = true
		}
	}

	return upper && lower && digit && special
}

// ValidateLoginCredentials checks the login payload presence.
func (s *UserService) ValidateLoginCredentials(data dtos.LoginDTO) utils.Result[any] {
	
	result := utils.Result[any]{Error: false, Message: "", Code: 200, Data: map[string]any{}}

	switch {
	case strings.TrimSpace(data.Email) == "":
		result.Error = true
		result.Message = "email is required"
	case data.Password == "":
		result.Error = true
		result.Message = "password is required"
	}

	return result
}

// ValidateUserType reports whether userType is in the allowed list.
// An empty allowlist rejects everything; pass utils.UserTypeValues for all.
func (s *UserService) ValidateUserType(userType string, allowed []string) bool {
	
	for _, allowedType := range allowed {
		if userType == allowedType {
			return true
		}
	}
	return false
}

// AttachPhoneCode joins a dial code and a local number.
// A dashed code drops its first 3 characters and gains a plus prefix,
// mirroring the TS substring handling exactly.
func (s *UserService) AttachPhoneCode(code string, phone string) string {
	
	if code == "" || phone == "" {
		return ""
	}

	codeStr := code
	if strings.Contains(code, "-") {
		rest := ""
		if len(code) > 3 {
			rest = code[3:]
		}
		codeStr = "+" + rest
	} else if !strings.Contains(code, "+") {
		codeStr = code
	}

	local := ""
	if len(phone) > 1 {
		local = phone[1:]
	}

	return codeStr + local
}

// CheckPhoneCode normalizes a phone against its dial code.
// Long local numbers shed their first 3 characters; already-prefixed
// numbers pass through untouched.
func (s *UserService) CheckPhoneCode(code string, phone string) string {
	if code == "" || phone == "" {
		return ""
	}

	if !strings.Contains(phone, "+") && len(phone) > 10 {
		rest := ""
		if len(phone) > 3 {
			rest = phone[3:]
		}
		return code + rest
	}

	if strings.Contains(phone, "+") {
		return phone
	}

	return ""
}
