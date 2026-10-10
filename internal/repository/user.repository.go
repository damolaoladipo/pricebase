package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	commonutils "github.com/damolaoladipo/common-go/utils"
	"github.com/damolaoladipo/pricebase/internal/models"
	"github.com/damolaoladipo/pricebase/internal/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// CreateUser =>
// This function is used to create a new user in the database.
// It returns the created user and an error if the user creation fails.
func (r *UserRepository) CreateUser(ctx context.Context, data models.User) (models.User, error) {

	query := utils.USER_CREATE
	var user models.User

	// IDs use utils.GenerateID (24-char hex), never database defaults.
	if data.ID == "" {
		data.ID = utils.GenerateID()
	}

	loginJSON, err := commonutils.ToJSON(data.Login)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to marshal user login: %w", err)
	}
	onboardJSON, err := commonutils.ToJSON(data.Onboard)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to marshal user onboard: %w", err)
	}
	registerJSON, err := commonutils.ToJSON(data.Register)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to marshal user register: %w", err)
	}
	statusJSON, err := commonutils.ToJSON(data.Status)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to marshal user status: %w", err)
	}
	oauth := data.OAuth
	if oauth == nil {
		oauth = []utils.OAuthProvider{}
	}
	oauthJSON, err := commonutils.ToJSON(oauth)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to marshal user oauth: %w", err)
	}
	countryJSON, err := commonutils.ToJSON(data.Country)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to marshal user country: %w", err)
	}
	notifications := data.NotificationsID
	if notifications == nil {
		notifications = []string{}
	}
	notificationsJSON, err := commonutils.ToJSON(notifications)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to marshal user notifications: %w", err)
	}
	devices := data.DevicesID
	if devices == nil {
		devices = []string{}
	}
	devicesJSON, err := commonutils.ToJSON(devices)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to marshal user devices: %w", err)
	}

	err = r.db.QueryRow(
		ctx,
		query,
		data.ID,
		data.UserType,
		data.FirstName,
		data.LastName,
		data.Email,
		data.Password,
		data.SavedPassword,
		data.IsSuper,
		data.IsAdmin,
		data.IsBusiness,
		data.IsUser,
		data.IsTeam,
		data.IsActivated,
		data.IsActive,
		data.IsWelcomed,
		data.BusinessName,
		data.BusinessID,
		loginJSON,
		onboardJSON,
		registerJSON,
		statusJSON,
		oauthJSON,
		countryJSON,
		notificationsJSON,
		devicesJSON,
		data.MiddleName,
		data.Avatar,
		data.Logo,
		data.PhoneNumber,
		data.PhoneCode,
		data.CountryPhone,
		data.AltPhone,
		data.PasswordType,
		data.Passcode,
		data.TeamRoleID,
		data.SettingsID,
	).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.FirstName,
		&user.LastName,
		&user.SavedPassword,
		&user.UserType,
		&user.IsSuper,
		&user.IsAdmin,
		&user.IsBusiness,
		&user.IsUser,
		&user.IsTeam,
		&user.IsActivated,
		&user.IsActive,
		&user.IsWelcomed,
		&user.IsLocked,
		&user.LoginLimit,
		&user.LockedUntil,
		&user.BusinessName,
		&user.BusinessID,
		&loginJSON,
		&onboardJSON,
		&registerJSON,
		&statusJSON,
		&oauthJSON,
		&countryJSON,
		&notificationsJSON,
		&devicesJSON,
		&user.MiddleName,
		&user.Avatar,
		&user.Logo,
		&user.PhoneNumber,
		&user.PhoneCode,
		&user.CountryPhone,
		&user.AltPhone,
		&user.PasswordType,
		&user.Passcode,
		&user.TeamRoleID,
		&user.SettingsID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Version,
	)

	if err != nil {

		return models.User{}, fmt.Errorf(
			"failed to create user: %w",
			err,
		)

	}

	if len(loginJSON) > 0 {
		if err := commonutils.FromJSON(loginJSON, &user.Login); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user login: %w", err)
		}
	}
	if len(onboardJSON) > 0 {
		if err := commonutils.FromJSON(onboardJSON, &user.Onboard); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user onboard: %w", err)
		}
	}
	if len(registerJSON) > 0 {
		if err := commonutils.FromJSON(registerJSON, &user.Register); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user register: %w", err)
		}
	}
	if len(statusJSON) > 0 {
		if err := commonutils.FromJSON(statusJSON, &user.Status); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user status: %w", err)
		}
	}
	if len(oauthJSON) > 0 {
		if err := commonutils.FromJSON(oauthJSON, &user.OAuth); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user oauth: %w", err)
		}
	}
	if len(countryJSON) > 0 {
		if err := commonutils.FromJSON(countryJSON, &user.Country); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user country: %w", err)
		}
	}
	if len(notificationsJSON) > 0 {
		if err := commonutils.FromJSON(notificationsJSON, &user.NotificationsID); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user notifications: %w", err)
		}
	}
	if len(devicesJSON) > 0 {
		if err := commonutils.FromJSON(devicesJSON, &user.DevicesID); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user devices: %w", err)
		}
	}

	return user, nil

}

// find by Email
func (r *UserRepository) FindUserByEmail(ctx context.Context, email string) (models.User, error) {

	query := utils.USER_FIND_BY_EMAIL

	var user models.User
	var loginJSON, onboardJSON, registerJSON, statusJSON, oauthJSON, countryJSON, notificationsJSON, devicesJSON []byte
	var roleIDsJSON, permissionsJSON []byte

	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.SavedPassword,
		&user.UserType,
		&roleIDsJSON,
		&permissionsJSON,
		&user.IsSuper,
		&user.IsAdmin,
		&user.IsBusiness,
		&user.IsUser,
		&user.IsTeam,
		&user.IsActivated,
		&user.IsActive,
		&user.IsWelcomed,
		&user.IsLocked,
		&user.LoginLimit,
		&user.LockedUntil,
		&user.BusinessName,
		&user.BusinessID,
		&loginJSON,
		&onboardJSON,
		&registerJSON,
		&statusJSON,
		&oauthJSON,
		&countryJSON,
		&notificationsJSON,
		&devicesJSON,
		&user.MiddleName,
		&user.Avatar,
		&user.Logo,
		&user.PhoneNumber,
		&user.PhoneCode,
		&user.CountryPhone,
		&user.AltPhone,
		&user.PasswordType,
		&user.Passcode,
		&user.TeamRoleID,
		&user.SettingsID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, pgx.ErrNoRows
		}

		return models.User{}, fmt.Errorf(
			"failed to find user by Email: %w",
			err,
		)

	}

	if len(roleIDsJSON) > 0 {
		if err := commonutils.FromJSON(roleIDsJSON, &user.RoleIDs); err != nil {
			return models.User{}, fmt.Errorf(
				"failed to decode role ids: %w",
				err,
			)
		}
	}

	if len(permissionsJSON) > 0 {
		if err := commonutils.FromJSON(permissionsJSON, &user.RolePermissions); err != nil {
			return models.User{}, fmt.Errorf(
				"failed to decode role permissions: %w",
				err,
			)
		}
	}

	if len(loginJSON) > 0 {
		if err := commonutils.FromJSON(loginJSON, &user.Login); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user login: %w", err)
		}
	}
	if len(onboardJSON) > 0 {
		if err := commonutils.FromJSON(onboardJSON, &user.Onboard); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user onboard: %w", err)
		}
	}
	if len(registerJSON) > 0 {
		if err := commonutils.FromJSON(registerJSON, &user.Register); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user register: %w", err)
		}
	}
	if len(statusJSON) > 0 {
		if err := commonutils.FromJSON(statusJSON, &user.Status); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user status: %w", err)
		}
	}
	if len(oauthJSON) > 0 {
		if err := commonutils.FromJSON(oauthJSON, &user.OAuth); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user oauth: %w", err)
		}
	}
	if len(countryJSON) > 0 {
		if err := commonutils.FromJSON(countryJSON, &user.Country); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user country: %w", err)
		}
	}
	if len(notificationsJSON) > 0 {
		if err := commonutils.FromJSON(notificationsJSON, &user.NotificationsID); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user notifications: %w", err)
		}
	}
	if len(devicesJSON) > 0 {
		if err := commonutils.FromJSON(devicesJSON, &user.DevicesID); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user devices: %w", err)
		}
	}

	return user, nil

}

// finds the user by ID.
//
// It wraps utils.USER_FIND_BY_ID and returns pgx.ErrNoRows when no user
// matches id.
//
// Returns the user and any errors that occur.
func (r *UserRepository) FindUserByID(ctx context.Context, id string) (models.User, error) {

	query := utils.USER_FIND_BY_ID
	var user models.User
	var loginJSON, onboardJSON, registerJSON, statusJSON, oauthJSON, countryJSON, notificationsJSON, devicesJSON []byte
	var roleIDsJSON, permissionsJSON []byte

	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.SavedPassword,
		&user.UserType,
		&roleIDsJSON,
		&permissionsJSON,
		&user.IsSuper,
		&user.IsAdmin,
		&user.IsBusiness,
		&user.IsUser,
		&user.IsTeam,
		&user.IsActivated,
		&user.IsActive,
		&user.IsWelcomed,
		&user.IsLocked,
		&user.LoginLimit,
		&user.LockedUntil,
		&user.BusinessName,
		&user.BusinessID,
		&loginJSON,
		&onboardJSON,
		&registerJSON,
		&statusJSON,
		&oauthJSON,
		&countryJSON,
		&notificationsJSON,
		&devicesJSON,
		&user.MiddleName,
		&user.Avatar,
		&user.Logo,
		&user.PhoneNumber,
		&user.PhoneCode,
		&user.CountryPhone,
		&user.AltPhone,
		&user.PasswordType,
		&user.Passcode,
		&user.TeamRoleID,
		&user.SettingsID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, pgx.ErrNoRows
		}

		return models.User{}, fmt.Errorf(
			"failed to find user by ID: %w",
			err,
		)

	}

	if len(roleIDsJSON) > 0 {
		if err := commonutils.FromJSON(roleIDsJSON, &user.RoleIDs); err != nil {
			return models.User{}, fmt.Errorf(
				"failed to decode role ids: %w",
				err,
			)
		}
	}

	if len(permissionsJSON) > 0 {
		if err := commonutils.FromJSON(permissionsJSON, &user.RolePermissions); err != nil {
			return models.User{}, fmt.Errorf(
				"failed to decode role permissions: %w",
				err,
			)
		}
	}

	if len(loginJSON) > 0 {
		if err := commonutils.FromJSON(loginJSON, &user.Login); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user login: %w", err)
		}
	}
	if len(onboardJSON) > 0 {
		if err := commonutils.FromJSON(onboardJSON, &user.Onboard); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user onboard: %w", err)
		}
	}
	if len(registerJSON) > 0 {
		if err := commonutils.FromJSON(registerJSON, &user.Register); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user register: %w", err)
		}
	}
	if len(statusJSON) > 0 {
		if err := commonutils.FromJSON(statusJSON, &user.Status); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user status: %w", err)
		}
	}
	if len(oauthJSON) > 0 {
		if err := commonutils.FromJSON(oauthJSON, &user.OAuth); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user oauth: %w", err)
		}
	}
	if len(countryJSON) > 0 {
		if err := commonutils.FromJSON(countryJSON, &user.Country); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user country: %w", err)
		}
	}
	if len(notificationsJSON) > 0 {
		if err := commonutils.FromJSON(notificationsJSON, &user.NotificationsID); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user notifications: %w", err)
		}
	}
	if len(devicesJSON) > 0 {
		if err := commonutils.FromJSON(devicesJSON, &user.DevicesID); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user devices: %w", err)
		}
	}

	return user, nil

}

// UpdateSavedPassword persists a user's encrypted password.
// It wraps utils.USER_UPDATE_SAVED_PASSWORD and returns pgx.ErrNoRows when
// no user matches id.
func (r *UserRepository) UpdateSavedPassword(ctx context.Context, id string, savedPassword string) error {

	res, err := r.db.Exec(ctx, utils.USER_UPDATE_SAVED_PASSWORD, id, savedPassword)
	if err != nil {
		return fmt.Errorf(
			"failed to update user password: %w",
			err,
		)
	}

	if res.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// UpdateAPIKeys persists a user's active API key and key history as JSON.
// It wraps utils.USER_UPDATE_API_KEYS and returns pgx.ErrNoRows when no
// user matches id.
func (r *UserRepository) UpdateAPIKeys(ctx context.Context, id string, apiKey utils.APIKey, keys []utils.APIKey) error {

	keyJSON, err := commonutils.ToJSON(apiKey)
	if err != nil {
		return fmt.Errorf(
			"failed to marshal api key: %w",
			err,
		)
	}

	keysJSON, err := commonutils.ToJSONArray(keys)
	if err != nil {
		return fmt.Errorf(
			"failed to marshal api keys: %w",
			err,
		)
	}

	res, err := r.db.Exec(ctx, utils.USER_UPDATE_API_KEYS, id, keyJSON, keysJSON)
	if err != nil {
		return fmt.Errorf(
			"failed to update user api keys: %w",
			err,
		)
	}

	if res.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// FindBySecretAPIKey
//
// finds the user holding a secret key/token pair.
// It wraps utils.USER_FIND_BY_SECRET_API_KEY.
func (r *UserRepository) FindBySecretAPIKey(ctx context.Context, key string, token string) (models.User, error) {

	var user models.User
	var loginJSON, onboardJSON, registerJSON, statusJSON, oauthJSON, countryJSON, notificationsJSON, devicesJSON []byte
	var apiKeyJSON []byte
	var apiKeysJSON []byte
	var roleIDsJSON, permissionsJSON []byte

	err := r.db.QueryRow(ctx, utils.USER_FIND_BY_SECRET_API_KEY, key, token).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.SavedPassword,
		&user.UserType,
		&apiKeyJSON,
		&apiKeysJSON,
		&roleIDsJSON,
		&permissionsJSON,
		&user.IsSuper,
		&user.IsAdmin,
		&user.IsBusiness,
		&user.IsUser,
		&user.IsTeam,
		&user.IsActivated,
		&user.IsActive,
		&user.IsWelcomed,
		&user.IsLocked,
		&user.LoginLimit,
		&user.LockedUntil,
		&user.BusinessName,
		&user.BusinessID,
		&loginJSON,
		&onboardJSON,
		&registerJSON,
		&statusJSON,
		&oauthJSON,
		&countryJSON,
		&notificationsJSON,
		&devicesJSON,
		&user.MiddleName,
		&user.Avatar,
		&user.Logo,
		&user.PhoneNumber,
		&user.PhoneCode,
		&user.CountryPhone,
		&user.AltPhone,
		&user.PasswordType,
		&user.Passcode,
		&user.TeamRoleID,
		&user.SettingsID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, pgx.ErrNoRows
		}

		return models.User{}, fmt.Errorf(
			"failed to find user by secret api key: %w",
			err,
		)
	}

	if len(apiKeyJSON) > 0 {
		if err := commonutils.FromJSON(apiKeyJSON, &user.APIKey); err != nil {
			return models.User{}, fmt.Errorf(
				"failed to decode api key: %w",
				err,
			)
		}
	}

	if len(apiKeysJSON) > 0 {
		if err := commonutils.FromJSON(apiKeysJSON, &user.Keys); err != nil {
			return models.User{}, fmt.Errorf(
				"failed to decode api keys: %w",
				err,
			)
		}
	}

	if len(roleIDsJSON) > 0 {
		if err := commonutils.FromJSON(roleIDsJSON, &user.RoleIDs); err != nil {
			return models.User{}, fmt.Errorf(
				"failed to decode role ids: %w",
				err,
			)
		}
	}

	if len(permissionsJSON) > 0 {
		if err := commonutils.FromJSON(permissionsJSON, &user.RolePermissions); err != nil {
			return models.User{}, fmt.Errorf(
				"failed to decode role permissions: %w",
				err,
			)
		}
	}

	if len(loginJSON) > 0 {
		if err := commonutils.FromJSON(loginJSON, &user.Login); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user login: %w", err)
		}
	}
	if len(onboardJSON) > 0 {
		if err := commonutils.FromJSON(onboardJSON, &user.Onboard); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user onboard: %w", err)
		}
	}
	if len(registerJSON) > 0 {
		if err := commonutils.FromJSON(registerJSON, &user.Register); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user register: %w", err)
		}
	}
	if len(statusJSON) > 0 {
		if err := commonutils.FromJSON(statusJSON, &user.Status); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user status: %w", err)
		}
	}
	if len(oauthJSON) > 0 {
		if err := commonutils.FromJSON(oauthJSON, &user.OAuth); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user oauth: %w", err)
		}
	}
	if len(countryJSON) > 0 {
		if err := commonutils.FromJSON(countryJSON, &user.Country); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user country: %w", err)
		}
	}
	if len(notificationsJSON) > 0 {
		if err := commonutils.FromJSON(notificationsJSON, &user.NotificationsID); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user notifications: %w", err)
		}
	}
	if len(devicesJSON) > 0 {
		if err := commonutils.FromJSON(devicesJSON, &user.DevicesID); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user devices: %w", err)
		}
	}

	return user, nil
}

// FindByPublicAPIKey
//
// finds the user holding a public key/token pair.
// It wraps utils.USER_FIND_BY_PUBLIC_API_KEY.
func (r *UserRepository) FindByPublicAPIKey(ctx context.Context, key string, token string) (models.User, error) {

	var user models.User
	var loginJSON, onboardJSON, registerJSON, statusJSON, oauthJSON, countryJSON, notificationsJSON, devicesJSON []byte
	var apiKeyJSON []byte
	var apiKeysJSON []byte
	var roleIDsJSON, permissionsJSON []byte

	err := r.db.QueryRow(ctx, utils.USER_FIND_BY_PUBLIC_API_KEY, key, token).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.SavedPassword,
		&user.UserType,
		&apiKeyJSON,
		&apiKeysJSON,
		&roleIDsJSON,
		&permissionsJSON,
		&user.IsSuper,
		&user.IsAdmin,
		&user.IsBusiness,
		&user.IsUser,
		&user.IsTeam,
		&user.IsActivated,
		&user.IsActive,
		&user.IsWelcomed,
		&user.IsLocked,
		&user.LoginLimit,
		&user.LockedUntil,
		&user.BusinessName,
		&user.BusinessID,
		&loginJSON,
		&onboardJSON,
		&registerJSON,
		&statusJSON,
		&oauthJSON,
		&countryJSON,
		&notificationsJSON,
		&devicesJSON,
		&user.MiddleName,
		&user.Avatar,
		&user.Logo,
		&user.PhoneNumber,
		&user.PhoneCode,
		&user.CountryPhone,
		&user.AltPhone,
		&user.PasswordType,
		&user.Passcode,
		&user.TeamRoleID,
		&user.SettingsID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, pgx.ErrNoRows
		}

		return models.User{}, fmt.Errorf(
			"failed to find user by public api key: %w",
			err,
		)
	}

	if len(apiKeyJSON) > 0 {
		if err := commonutils.FromJSON(apiKeyJSON, &user.APIKey); err != nil {
			return models.User{}, fmt.Errorf(
				"failed to decode api key: %w",
				err,
			)
		}
	}

	if len(apiKeysJSON) > 0 {
		if err := commonutils.FromJSON(apiKeysJSON, &user.Keys); err != nil {
			return models.User{}, fmt.Errorf(
				"failed to decode api keys: %w",
				err,
			)
		}
	}

	if len(roleIDsJSON) > 0 {
		if err := commonutils.FromJSON(roleIDsJSON, &user.RoleIDs); err != nil {
			return models.User{}, fmt.Errorf(
				"failed to decode role ids: %w",
				err,
			)
		}
	}

	if len(permissionsJSON) > 0 {
		if err := commonutils.FromJSON(permissionsJSON, &user.RolePermissions); err != nil {
			return models.User{}, fmt.Errorf(
				"failed to decode role permissions: %w",
				err,
			)
		}
	}

	if len(loginJSON) > 0 {
		if err := commonutils.FromJSON(loginJSON, &user.Login); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user login: %w", err)
		}
	}
	if len(onboardJSON) > 0 {
		if err := commonutils.FromJSON(onboardJSON, &user.Onboard); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user onboard: %w", err)
		}
	}
	if len(registerJSON) > 0 {
		if err := commonutils.FromJSON(registerJSON, &user.Register); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user register: %w", err)
		}
	}
	if len(statusJSON) > 0 {
		if err := commonutils.FromJSON(statusJSON, &user.Status); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user status: %w", err)
		}
	}
	if len(oauthJSON) > 0 {
		if err := commonutils.FromJSON(oauthJSON, &user.OAuth); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user oauth: %w", err)
		}
	}
	if len(countryJSON) > 0 {
		if err := commonutils.FromJSON(countryJSON, &user.Country); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user country: %w", err)
		}
	}
	if len(notificationsJSON) > 0 {
		if err := commonutils.FromJSON(notificationsJSON, &user.NotificationsID); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user notifications: %w", err)
		}
	}
	if len(devicesJSON) > 0 {
		if err := commonutils.FromJSON(devicesJSON, &user.DevicesID); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user devices: %w", err)
		}
	}

	return user, nil
}

// HasRole reports whether a user already holds a role.
// It wraps utils.USER_ROLE_EXISTS.
func (r *UserRepository) HasRole(ctx context.Context, userID string, roleID string) (bool, error) {

	var exists bool
	err := r.db.QueryRow(ctx, utils.USER_ROLE_EXISTS, userID, roleID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf(
			"failed to check user role: %w",
			err,
		)
	}

	return exists, nil
}

// AttachRole links a user to a role via utils.USER_ROLE_INSERT.
// It returns rows inserted (0 when the link already existed).
func (r *UserRepository) AttachRole(ctx context.Context, userID string, roleID string) (int64, error) {

	res, err := r.db.Exec(ctx, utils.USER_ROLE_INSERT, userID, roleID)
	if err != nil {
		return 0, fmt.Errorf(
			"failed to attach role to user: %w",
			err,
		)
	}

	return res.RowsAffected(), nil
}

// UpdateRolesPermissions persists a user's role snapshot: the linked role
// IDs and the derived permission set. It wraps
// utils.USER_UPDATE_ROLES_PERMISSIONS and returns pgx.ErrNoRows when no
// user matches id. Only attachSuperRole calls this; snapshots are
// recomputed every boot so they never rot.
func (r *UserRepository) UpdateRolesPermissions(ctx context.Context, id string, roleIDs []string, permissions []utils.UserPermission) error {

	roleIDsJSON, err := commonutils.ToJSONArray(roleIDs)
	if err != nil {
		return fmt.Errorf(
			"failed to marshal role ids: %w",
			err,
		)
	}

	permissionsJSON, err := commonutils.ToJSONArray(permissions)
	if err != nil {
		return fmt.Errorf(
			"failed to marshal role permissions: %w",
			err,
		)
	}

	res, err := r.db.Exec(
		ctx,
		utils.USER_UPDATE_ROLES_PERMISSIONS,
		id,
		roleIDsJSON,
		permissionsJSON,
	)

	if err != nil {
		return fmt.Errorf(
			"failed to update user roles permissions: %w",
			err,
		)
	}

	if res.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// UpdateLoginLimit persists a user's login limit, lock flag and lock expiry.
// It wraps utils.USER_UPDATE_LOGIN_LIMIT and returns pgx.ErrNoRows when no
// user matches id. Pair with User.IncreaseLoginLimit, which computes the
// new values in memory.
func (r *UserRepository) UpdateLoginLimit(ctx context.Context, id string, limit int, locked bool, lockedUntil *time.Time) error {

	res, err := r.db.Exec(
		ctx,
		utils.USER_UPDATE_LOGIN_LIMIT,
		id,
		limit,
		locked,
		lockedUntil,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to update user login limit: %w",
			err,
		)
	}

	if res.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// UpdateResetToken persists a reset-password token hash and expiry.
// It wraps utils.USER_UPDATE_RESET_TOKEN. Pair with
// User.GetResetPasswordToken, which generates them.
func (r *UserRepository) UpdateResetToken(ctx context.Context, id, token string, expire *time.Time) error {

	res, err := r.db.Exec(
		ctx,
		utils.USER_UPDATE_RESET_TOKEN,
		id,
		token,
		expire,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to update user reset password token: %w",
			err,
		)
	}

	if res.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// UpdateActivationToken persists an activation token hash and expiry.
// It wraps utils.USER_UPDATE_ACTIVATION_TOKEN. Pair with
// User.GetActivationToken, which generates them.
func (r *UserRepository) UpdateActivationToken(ctx context.Context, id, token string, expire *time.Time) error {

	res, err := r.db.Exec(
		ctx,
		utils.USER_UPDATE_ACTIVATION_TOKEN,
		id,
		token,
		expire,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to update user activation token: %w",
			err,
		)
	}

	if res.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// UpdateInviteToken persists an invite token hash and expiry.
// It wraps utils.USER_UPDATE_INVITE_TOKEN. Pair with
// User.GetInviteToken, which generates them.
func (r *UserRepository) UpdateInviteToken(ctx context.Context, id, token string, expire *time.Time) error {

	res, err := r.db.Exec(
		ctx,
		utils.USER_UPDATE_INVITE_TOKEN,
		id,
		token,
		expire,
	)

	if err != nil {
		return fmt.Errorf(
			"failed to update user invite token: %w",
			err,
		)
	}

	if res.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (r *UserRepository) FindByPhone(ctx context.Context, phone string) (models.User, error) {

	query := utils.USER_FIND_BY_PHONE

	var user models.User
	var loginJSON, onboardJSON, registerJSON, statusJSON, oauthJSON, countryJSON, notificationsJSON, devicesJSON []byte
	var roleIDsJSON, permissionsJSON []byte

	err := r.db.QueryRow(ctx, query, phone).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.SavedPassword,
		&user.UserType,
		&roleIDsJSON,
		&permissionsJSON,
		&user.IsSuper,
		&user.IsAdmin,
		&user.IsBusiness,
		&user.IsUser,
		&user.IsTeam,
		&user.IsActivated,
		&user.IsActive,
		&user.IsWelcomed,
		&user.IsLocked,
		&user.LoginLimit,
		&user.LockedUntil,
		&user.BusinessName,
		&user.BusinessID,
		&loginJSON,
		&onboardJSON,
		&registerJSON,
		&statusJSON,
		&oauthJSON,
		&countryJSON,
		&notificationsJSON,
		&devicesJSON,
		&user.MiddleName,
		&user.Avatar,
		&user.Logo,
		&user.PhoneNumber,
		&user.PhoneCode,
		&user.CountryPhone,
		&user.AltPhone,
		&user.PasswordType,
		&user.Passcode,
		&user.TeamRoleID,
		&user.SettingsID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, pgx.ErrNoRows
		}

		return models.User{}, fmt.Errorf(
			"failed to find user by phone: %w",
			err,
		)

	}

	if len(roleIDsJSON) > 0 {
		if err := commonutils.FromJSON(roleIDsJSON, &user.RoleIDs); err != nil {
			return models.User{}, fmt.Errorf(
				"failed to decode role ids: %w",
				err,
			)
		}
	}

	if len(permissionsJSON) > 0 {
		if err := commonutils.FromJSON(permissionsJSON, &user.RolePermissions); err != nil {
			return models.User{}, fmt.Errorf(
				"failed to decode role permissions: %w",
				err,
			)
		}
	}

	if len(loginJSON) > 0 {
		if err := commonutils.FromJSON(loginJSON, &user.Login); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user login: %w", err)
		}
	}
	if len(onboardJSON) > 0 {
		if err := commonutils.FromJSON(onboardJSON, &user.Onboard); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user onboard: %w", err)
		}
	}
	if len(registerJSON) > 0 {
		if err := commonutils.FromJSON(registerJSON, &user.Register); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user register: %w", err)
		}
	}
	if len(statusJSON) > 0 {
		if err := commonutils.FromJSON(statusJSON, &user.Status); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user status: %w", err)
		}
	}
	if len(oauthJSON) > 0 {
		if err := commonutils.FromJSON(oauthJSON, &user.OAuth); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user oauth: %w", err)
		}
	}
	if len(countryJSON) > 0 {
		if err := commonutils.FromJSON(countryJSON, &user.Country); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user country: %w", err)
		}
	}
	if len(notificationsJSON) > 0 {
		if err := commonutils.FromJSON(notificationsJSON, &user.NotificationsID); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user notifications: %w", err)
		}
	}
	if len(devicesJSON) > 0 {
		if err := commonutils.FromJSON(devicesJSON, &user.DevicesID); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user devices: %w", err)
		}
	}

	return user, nil

}

func (r *UserRepository) FindByOTPCode(ctx context.Context, code string) (models.User, error) {

	query := utils.USER_FIND_BY_OTP

	var user models.User
	var loginJSON, onboardJSON, registerJSON, statusJSON, oauthJSON, countryJSON, notificationsJSON, devicesJSON []byte
	var roleIDsJSON, permissionsJSON []byte

	err := r.db.QueryRow(ctx, query, code).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.SavedPassword,
		&user.UserType,
		&roleIDsJSON,
		&permissionsJSON,
		&user.IsSuper,
		&user.IsAdmin,
		&user.IsBusiness,
		&user.IsUser,
		&user.IsTeam,
		&user.IsActivated,
		&user.IsActive,
		&user.IsWelcomed,
		&user.IsLocked,
		&user.LoginLimit,
		&user.LockedUntil,
		&user.BusinessName,
		&user.BusinessID,
		&loginJSON,
		&onboardJSON,
		&registerJSON,
		&statusJSON,
		&oauthJSON,
		&countryJSON,
		&notificationsJSON,
		&devicesJSON,
		&user.MiddleName,
		&user.Avatar,
		&user.Logo,
		&user.PhoneNumber,
		&user.PhoneCode,
		&user.CountryPhone,
		&user.AltPhone,
		&user.PasswordType,
		&user.Passcode,
		&user.TeamRoleID,
		&user.SettingsID,
		&user.OTPCode,
		&user.OTPCodeExpire,
		&user.OTPType,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, pgx.ErrNoRows
		}

		return models.User{}, fmt.Errorf(
			"failed to find user by otp: %w",
			err,
		)

	}

	if len(roleIDsJSON) > 0 {
		if err := commonutils.FromJSON(roleIDsJSON, &user.RoleIDs); err != nil {
			return models.User{}, fmt.Errorf(
				"failed to decode role ids: %w",
				err,
			)
		}
	}

	if len(permissionsJSON) > 0 {
		if err := commonutils.FromJSON(permissionsJSON, &user.RolePermissions); err != nil {
			return models.User{}, fmt.Errorf(
				"failed to decode role permissions: %w",
				err,
			)
		}
	}

	if len(loginJSON) > 0 {
		if err := commonutils.FromJSON(loginJSON, &user.Login); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user login: %w", err)
		}
	}
	if len(onboardJSON) > 0 {
		if err := commonutils.FromJSON(onboardJSON, &user.Onboard); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user onboard: %w", err)
		}
	}
	if len(registerJSON) > 0 {
		if err := commonutils.FromJSON(registerJSON, &user.Register); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user register: %w", err)
		}
	}
	if len(statusJSON) > 0 {
		if err := commonutils.FromJSON(statusJSON, &user.Status); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user status: %w", err)
		}
	}
	if len(oauthJSON) > 0 {
		if err := commonutils.FromJSON(oauthJSON, &user.OAuth); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user oauth: %w", err)
		}
	}
	if len(countryJSON) > 0 {
		if err := commonutils.FromJSON(countryJSON, &user.Country); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user country: %w", err)
		}
	}
	if len(notificationsJSON) > 0 {
		if err := commonutils.FromJSON(notificationsJSON, &user.NotificationsID); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user notifications: %w", err)
		}
	}
	if len(devicesJSON) > 0 {
		if err := commonutils.FromJSON(devicesJSON, &user.DevicesID); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user devices: %w", err)
		}
	}

	return user, nil

}

func (r *UserRepository) FindByEmailOTP(ctx context.Context, email, code, otpType string) (models.User, error) {

	query := utils.USER_FIND_BY_EMAIL_OTP

	var user models.User
	var loginJSON, onboardJSON, registerJSON, statusJSON, oauthJSON, countryJSON, notificationsJSON, devicesJSON []byte
	var roleIDsJSON, permissionsJSON []byte

	err := r.db.QueryRow(ctx, query, email, code, otpType).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.SavedPassword,
		&user.UserType,
		&roleIDsJSON,
		&permissionsJSON,
		&user.IsSuper,
		&user.IsAdmin,
		&user.IsBusiness,
		&user.IsUser,
		&user.IsTeam,
		&user.IsActivated,
		&user.IsActive,
		&user.IsWelcomed,
		&user.IsLocked,
		&user.LoginLimit,
		&user.LockedUntil,
		&user.BusinessName,
		&user.BusinessID,
		&loginJSON,
		&onboardJSON,
		&registerJSON,
		&statusJSON,
		&oauthJSON,
		&countryJSON,
		&notificationsJSON,
		&devicesJSON,
		&user.MiddleName,
		&user.Avatar,
		&user.Logo,
		&user.PhoneNumber,
		&user.PhoneCode,
		&user.CountryPhone,
		&user.AltPhone,
		&user.PasswordType,
		&user.Passcode,
		&user.TeamRoleID,
		&user.SettingsID,
		&user.OTPCode,
		&user.OTPCodeExpire,
		&user.OTPType,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, pgx.ErrNoRows
		}

		return models.User{}, fmt.Errorf(
			"failed to find user by email and otp: %w",
			err,
		)

	}

	if len(roleIDsJSON) > 0 {
		if err := commonutils.FromJSON(roleIDsJSON, &user.RoleIDs); err != nil {
			return models.User{}, fmt.Errorf(
				"failed to decode role ids: %w",
				err,
			)
		}
	}

	if len(permissionsJSON) > 0 {
		if err := commonutils.FromJSON(permissionsJSON, &user.RolePermissions); err != nil {
			return models.User{}, fmt.Errorf(
				"failed to decode role permissions: %w",
				err,
			)
		}
	}

	if len(loginJSON) > 0 {
		if err := commonutils.FromJSON(loginJSON, &user.Login); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user login: %w", err)
		}
	}
	if len(onboardJSON) > 0 {
		if err := commonutils.FromJSON(onboardJSON, &user.Onboard); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user onboard: %w", err)
		}
	}
	if len(registerJSON) > 0 {
		if err := commonutils.FromJSON(registerJSON, &user.Register); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user register: %w", err)
		}
	}
	if len(statusJSON) > 0 {
		if err := commonutils.FromJSON(statusJSON, &user.Status); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user status: %w", err)
		}
	}
	if len(oauthJSON) > 0 {
		if err := commonutils.FromJSON(oauthJSON, &user.OAuth); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user oauth: %w", err)
		}
	}
	if len(countryJSON) > 0 {
		if err := commonutils.FromJSON(countryJSON, &user.Country); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user country: %w", err)
		}
	}
	if len(notificationsJSON) > 0 {
		if err := commonutils.FromJSON(notificationsJSON, &user.NotificationsID); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user notifications: %w", err)
		}
	}
	if len(devicesJSON) > 0 {
		if err := commonutils.FromJSON(devicesJSON, &user.DevicesID); err != nil {
			return models.User{}, fmt.Errorf("failed to decode user devices: %w", err)
		}
	}

	return user, nil

}

func (r *UserRepository) UpdateUserType(ctx context.Context, id string, userType string) error {

	res, err := r.db.Exec(
		ctx,
		utils.USER_UPDATE_USER_TYPE,
		id,
		userType,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to update user type: %w",
			err,
		)
	}

	if res.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil

}

func (r *UserRepository) UpdateLogin(ctx context.Context, id string, login utils.UserLogin) error {

	loginJSON, err := commonutils.ToJSON(login)
	if err != nil {
		return fmt.Errorf(
			"failed to marshal user login: %w",
			err,
		)
	}

	res, err := r.db.Exec(
		ctx,
		utils.USER_UPDATE_LOGIN,
		id,
		loginJSON,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to update user login: %w",
			err,
		)
	}

	if res.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil

}

func (r *UserRepository) UpdateAccountStatus(ctx context.Context, id string, activated bool, active bool, locked bool, loginLimit int) error {

	res, err := r.db.Exec(
		ctx,
		utils.USER_UPDATE_ACCOUNT_STATUS,
		id,
		activated,
		active,
		locked,
		loginLimit,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to update user account status: %w",
			err,
		)
	}

	if res.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil

}

func (r *UserRepository) UpdateOTP(ctx context.Context, id string, code string, expiry *time.Time, otpType string) error {

	res, err := r.db.Exec(
		ctx,
		utils.USER_UPDATE_OTP,
		id,
		code,
		expiry,
		otpType,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to update user otp: %w",
			err,
		)
	}

	if res.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil

}
