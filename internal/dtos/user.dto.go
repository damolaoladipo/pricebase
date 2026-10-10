package dtos

type SeedUserDTO struct {
	FirstName       string `json:"firstName"`
	LastName        string `json:"lastName"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	SavedPassword   string `json:"savedPassword"`
	UserType        string `json:"userType"`
	IsActive        bool   `json:"isActive"`
	IsSuper         bool   `json:"isSuper"`
	IsAdmin         bool   `json:"isAdmin"`
	IsBusiness      bool   `json:"isBusiness"`
	IsTeam          bool   `json:"isTeam"`
	IsUser          bool   `json:"isUser"`
	IsWelcomed      bool   `json:"isWelcomed"`
	IsActivated     bool   `json:"isActivated"`
	IsPhoneVerified bool   `json:"isPhoneVerified"`
}

type CreateUserDTO struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type APIKeyType string

const (
	APIKeyTypeSecret APIKeyType = "secret"
	APIKeyTypePublic APIKeyType = "public"
)

type DecodeAPIKeyDTO struct {
	APIKey string     `json:"apikey" binding:"required"`
	Type   APIKeyType `json:"type" binding:"required,oneof=secret public"`
}

// RegisterUserResponse mirrors TS UserMapper.mapRegister.
type RegisterUserResponse struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	PhoneNumber  string `json:"phoneNumber"`
	PasswordType string `json:"passwordType"`
	UserType     string `json:"userType"`
}

// RoleResponse is one entry of the mapped roles list.
type RoleResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// CountryResponse mirrors the TS country shape in mapLoginUser.
type CountryResponse struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// SettingsResponse mirrors the TS settings shape in mapLoginUser.
// A nil *SettingsResponse serializes the field as null.
type SettingsResponse struct {
	SMS   bool `json:"sms"`
	Email bool `json:"email"`
}

// LoginResponse mirrors the TS login shape in mapLoginUser.
type LoginResponse struct {
	Last   string `json:"last"`
	Method string `json:"method"`
}

// OnboardResponse mirrors the TS onboard shape in mapLoginUser.
type OnboardResponse struct {
	Status string `json:"status"`
	Stage  string `json:"stage"`
}

// LoginUserResponse mirrors TS UserMapper.mapLoginUser.
type LoginUserResponse struct {
	ID              string            `json:"id"`
	Email           string            `json:"email"`
	PhoneNumber     string            `json:"phoneNumber"`
	PhoneCode       string            `json:"phoneCode"`
	UserType        string            `json:"userType"`
	PasswordType    string            `json:"passwordType"`
	PasswordHash    string            `json:"passwordHash"`
	Login           LoginResponse     `json:"login"`
	Onboard         OnboardResponse   `json:"onboard"`
	Settings        *SettingsResponse `json:"settings"`
	Roles           []RoleResponse    `json:"roles"`
	Country         CountryResponse   `json:"country"`
	IsSuper         bool              `json:"isSuper"`
	IsAdmin         bool              `json:"isAdmin"`
	IsUser          bool              `json:"isUser"`
	IsOrganization  bool              `json:"isOrganization"`
	IsTeam          bool              `json:"isTeam"`
	IsPhoneVerified bool              `json:"isPhoneVerified"`
	IsActivated     bool              `json:"isActivated"`
	IsActive        bool              `json:"isActive"`
	CreatedAt       any               `json:"createdAt"`
	UpdatedAt       any               `json:"updatedAt"`
}
