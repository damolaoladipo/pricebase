package unit

import (
	"testing"

	"github.com/damolaoladipo/pricebase/internal/dtos"
	"github.com/damolaoladipo/pricebase/internal/mappers"
	"github.com/damolaoladipo/pricebase/internal/models"
	"github.com/damolaoladipo/pricebase/internal/utils"
)

func testMapperUser() *models.User {
	return &models.User{
		ID:              "user-id-1",
		Email:           "ada@example.com",
		PhoneNumber:     "08030000000",
		PhoneCode:       "+234",
		UserType:        string(utils.UserTypeOrganization),
		PasswordType:    "bcrypt",
		SavedPassword:   "hashed",
		Login:           utils.UserLogin{Last: "2026-01-01", Method: "email"},
		Onboard:         utils.UserOnboard{Stage: "done", Status: "complete"},
		Country:         utils.Country{Name: "Nigeria", Code2: "NG"},
		IsSuper:         false,
		IsAdmin:         true,
		IsUser:          true,
		IsBusiness:      true,
		IsTeam:          false,
		IsPhoneVerified: true,
		IsActivated:     true,
		IsActive:        true,
	}
}

func testMapperRoles() []models.Role {
	return []models.Role{
		{ID: "role-id-1", Name: "admin"},
	}
}

func TestMapRegister(t *testing.T) {
	m := mappers.NewUserMapper()
	got := m.MapRegister(testMapperUser())

	if got.ID != "user-id-1" || got.Email != "ada@example.com" {
		t.Errorf("identity mismatch: %+v", got)
	}
	if got.PhoneNumber != "08030000000" || got.PasswordType != "bcrypt" {
		t.Errorf("contact/secret mismatch: %+v", got)
	}
	if got.UserType != string(utils.UserTypeOrganization) {
		t.Errorf("userType mismatch: %q", got.UserType)
	}
}

func TestMapLoginUserNonSuperKeepsSettings(t *testing.T) {
	m := mappers.NewUserMapper()
	settings := &dtos.SettingsResponse{SMS: true, Email: false}
	got := m.MapLoginUser(testMapperUser(), testMapperRoles(), settings)

	if got.Settings == nil {
		t.Fatal("expected settings for non-superadmin user")
	}
	if !got.Settings.SMS || got.Settings.Email {
		t.Errorf("settings mismatch: %+v", got.Settings)
	}
	if len(got.Roles) != 1 || got.Roles[0].ID != "role-id-1" || got.Roles[0].Name != "admin" {
		t.Errorf("roles mismatch: %+v", got.Roles)
	}
	if got.Country.Code != "NG" || got.Country.Name != "Nigeria" {
		t.Errorf("country mismatch: %+v", got.Country)
	}
	if got.Login.Method != "email" || got.Onboard.Stage != "done" {
		t.Errorf("login/onboard mismatch: %+v / %+v", got.Login, got.Onboard)
	}
	if got.PasswordHash != "hashed" {
		t.Errorf("password hash must carry savedPassword, got %q", got.PasswordHash)
	}
	if !got.IsOrganization {
		t.Error("IsOrganization must mirror IsBusiness")
	}
}

func TestMapLoginUserSuperDropsSettings(t *testing.T) {
	m := mappers.NewUserMapper()
	user := testMapperUser()
	user.UserType = string(utils.UserTypeSuper)

	got := m.MapLoginUser(user, testMapperRoles(), &dtos.SettingsResponse{SMS: true, Email: true})
	if got.Settings != nil {
		t.Errorf("expected nil settings for superadmin, got %+v", got.Settings)
	}
}

func TestMapLoginUserEmptyRoles(t *testing.T) {
	m := mappers.NewUserMapper()
	user := testMapperUser()
	got := m.MapLoginUser(user, nil, nil)
	if got.Roles == nil || len(got.Roles) != 0 {
		t.Errorf("expected empty roles slice, got %+v", got.Roles)
	}
}
