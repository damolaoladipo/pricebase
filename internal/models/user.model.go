package models

import (
	"time"

	"github.com/damolaoladipo/pricebase/internal/utils"
)

type User struct {
	Avatar string `json:"avatar" db:"avatar"`
	Logo   string `json:"logo" db:"logo"`

	FirstName  string `json:"firstName" db:"first_name"`
	LastName   string `json:"lastName" db:"last_name"`
	MiddleName string `json:"middleName" db:"middle_name"`

	PhoneNumber  string `json:"phoneNumber" db:"phone_number"`
	PhoneCode    string `json:"phoneCode" db:"phone_code"`
	CountryPhone string `json:"countryPhone" db:"country_phone"`
	AltPhone     string `json:"altPhone" db:"alt_phone"`

	Email         string `json:"email" db:"email"`
	Password      string `json:"password" db:"password"`
	PasswordType  string `json:"passwordType" db:"password_type"`
	SavedPassword string `json:"savedPassword" db:"saved_password"`
	Passcode      string `json:"passcode" db:"passcode"`

	UserType     string `json:"userType" db:"user_type"`
	BusinessName string `json:"businessName" db:"business_name"`

	Login    utils.UserLogin    `json:"login" db:"login"`
	Onboard  utils.UserOnboard  `json:"onboard" db:"onboard"`
	Register utils.UserRegister `json:"register" db:"register"`
	Status   utils.UserStatus   `json:"status" db:"status"`

	ActivationToken       *string    `json:"activationToken,omitempty" db:"activation_token"`
	ActivationTokenExpire *time.Time `json:"activationTokenExpire,omitempty" db:"activation_token_expire"`

	ResetPasswordToken       *string    `json:"resetPasswordToken,omitempty" db:"reset_password_token"`
	ResetPasswordTokenExpire *time.Time `json:"resetPasswordTokenExpire,omitempty" db:"reset_password_token_expire"`

	OTPCode       *string    `json:"otpCode,omitempty" db:"otp_code"`
	OTPCodeExpire *time.Time `json:"otpCodeExpire,omitempty" db:"otp_code_expire"`
	OTPType       string     `json:"otpType" db:"otp_type"`

	InviteToken       *string    `json:"inviteToken,omitempty" db:"invite_token"`
	InviteTokenExpire *time.Time `json:"inviteTokenExpire,omitempty" db:"invite_token_expire"`
	InviteStatus      string     `json:"inviteStatus" db:"invite_status"`

	OAuth []utils.OAuthProvider `json:"oauth" db:"oauth"`

	APIKey utils.APIKey   `json:"apiKey" db:"-"`
	Keys   []utils.APIKey `json:"keys" db:"keys"`

	IsSuper    bool `json:"isSuper" db:"is_super"`
	IsAdmin    bool `json:"isAdmin" db:"is_admin"`
	IsBusiness bool `json:"isbusiness" db:"is_business"`
	IsUser     bool `json:"isUser" db:"is_user"`
	IsTeam     bool `json:"isTeam" db:"is_team"`

	IsActivated     bool       `json:"isActivated" db:"is_activated"`
	IsActive        bool       `json:"isActive" db:"is_active"`
	IsLocked        bool       `json:"isLocked" db:"is_locked"`
	LockedUntil     *time.Time `json:"lockedUntil,omitempty" db:"locked_until"`
	IsPhoneVerified bool       `json:"isPhoneVerified" db:"is_phone_verified"`
	IsWelcomed      bool       `json:"isWelcomed" db:"is_welcomed"`
	LoginLimit      int        `json:"loginLimit" db:"login_limit"`

	Country utils.Country `json:"country" db:"country"`

	// Relationships
	BusinessID      string   `json:"businessId,omitempty" db:"business_id"`
	TeamRoleID      string   `json:"teamRoleId,omitempty" db:"team_role_id"`
	SettingsID      string   `json:"settingsId,omitempty" db:"settings_id"`
	NotificationsID []string `json:"notificationsId,omitempty" db:"notifications_id"`
	DevicesID       []string `json:"devicesId,omitempty" db:"devices_id"`

	RoleIDs         []string               `json:"roleIds" db:"role_ids"`
	RolePermissions []utils.UserPermission `json:"rolePermissions" db:"permissions"`

	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
	Version   int64     `json:"version" db:"version"`
	ID        string    `json:"id" db:"id"`
}
