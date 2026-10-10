package mappers

import (
	"github.com/damolaoladipo/pricebase/internal/dtos"
	"github.com/damolaoladipo/pricebase/internal/models"
	"github.com/damolaoladipo/pricebase/internal/utils"
)

type UserMapper struct{}

func NewUserMapper() *UserMapper {
	return &UserMapper{}
}

func (m *UserMapper) MapRegister(user *models.User) dtos.RegisterUserResponse {

	return dtos.RegisterUserResponse{
		ID:           user.ID,
		Email:        user.Email,
		PhoneNumber:  user.PhoneNumber,
		PasswordType: user.PasswordType,
		UserType:     user.UserType,
	}
}

func (m *UserMapper) MapLoginUser(user *models.User, roles []models.Role, settings *dtos.SettingsResponse) dtos.LoginUserResponse {

	roleList := make([]dtos.RoleResponse, 0, len(roles))
	for _, role := range roles {
		roleList = append(roleList, dtos.RoleResponse{
			ID:   role.ID,
			Name: role.Name,
		})
	}

	country := dtos.CountryResponse{
		Code: user.Country.Code2,
		Name: user.Country.Name,
	}

	var mappedSettings *dtos.SettingsResponse
	if user.UserType != string(utils.UserTypeSuper) {
		mappedSettings = settings
	}

	return dtos.LoginUserResponse{
		ID:           user.ID,
		Email:        user.Email,
		PhoneNumber:  user.PhoneNumber,
		PhoneCode:    user.PhoneCode,
		UserType:     user.UserType,
		PasswordType: user.PasswordType,
		PasswordHash: user.SavedPassword,
		Login: dtos.LoginResponse{
			Last:   user.Login.Last,
			Method: user.Login.Method,
		},
		Onboard: dtos.OnboardResponse{
			Status: user.Onboard.Status,
			Stage:  user.Onboard.Stage,
		},
		Settings:        mappedSettings,
		Roles:           roleList,
		Country:         country,
		IsSuper:         user.IsSuper,
		IsAdmin:         user.IsAdmin,
		IsUser:          user.IsUser,
		IsOrganization:  user.IsBusiness,
		IsTeam:          user.IsTeam,
		IsPhoneVerified: user.IsPhoneVerified,
		IsActivated:     user.IsActivated,
		IsActive:        user.IsActive,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
	}
}
