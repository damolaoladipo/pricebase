package repository

import (
	"context"
	"errors"
	"fmt"

	commonutils "github.com/damolaoladipo/common-go/utils"
	"github.com/damolaoladipo/pricebase/internal/models"
	"github.com/damolaoladipo/pricebase/internal/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PermissionRepository struct {
	db *pgxpool.Pool
}

func NewPermissionRepository(db *pgxpool.Pool) *PermissionRepository {
	return &PermissionRepository{
		db: db,
	}
}

// CreatePermission
//
// This function is used to create a new permission.
// It returns the permission and an error if the permission is not created.
func (r *PermissionRepository) CreatePermission(ctx context.Context, data models.Permission) (models.Permission, error) {

	query := utils.PERMISSION_INSERT
	var permission models.Permission
	var actionsJSON []byte

	// Actions are stored as a JSON array; ToJSONArray guarantees [] rather
	// than null when data.Actions is nil.
	actionsJSON, err := commonutils.ToJSONArray(data.Actions)
	if err != nil {
		return models.Permission{}, fmt.Errorf(
			"failed to marshal permission actions: %w",
			err,
		)
	}

	// IDs use utils.GenerateID (24-char hex), never database defaults.
	if data.ID == "" {
		data.ID = utils.GenerateID()
	}

	err = r.db.QueryRow(
		ctx,
		query,
		data.ID,
		data.Name,
		data.Code,
		data.Group,
		actionsJSON,
		data.Slug,
		data.UpdatedBy,
	).Scan(
		&permission.ID,
		&permission.Name,
		&permission.Code,
		&permission.Group,
		&actionsJSON,
		&permission.Slug,
		&permission.UpdatedBy,
		&permission.CreatedAt,
		&permission.UpdatedAt,
		&permission.Version,
	)
	if err != nil {
		return models.Permission{}, fmt.Errorf(
			"failed to create permission: %w",
			err,
		)
	}

	return permission, nil
}

// FindByID
//
// This function is used to find a permission by its ID.
// It returns the permission and an error if the permission is not found.
func (r *PermissionRepository) FindByID(ctx context.Context, id string) (models.Permission, error) {

	query := utils.PERMISSION_FIND_BY_ID
	var permission models.Permission
	var actionsJSON []byte

	err := r.db.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&permission.ID,
		&permission.Name,
		&permission.Code,
		&permission.Group,
		&actionsJSON,
		&permission.Slug,
		&permission.UpdatedBy,
		&permission.CreatedAt,
		&permission.UpdatedAt,
		&permission.Version,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Permission{}, pgx.ErrNoRows
		}

		return models.Permission{}, fmt.Errorf(
			"failed to find permission by ID: %w",
			err,
		)
	}

	if len(actionsJSON) > 0 {
		if err := commonutils.FromJSON(actionsJSON, &permission.Actions); err != nil {
			return models.Permission{}, fmt.Errorf(
				"failed to decode permission actions: %w",
				err,
			)
		}
	}

	return permission, nil
}

// FindByName
//
// This function is used to find a permission by its name.
// It returns the permission and an error if the permission is not found.
func (r *PermissionRepository) FindByName(ctx context.Context, name string) (models.Permission, error) {

	query := utils.PERMISSION_FIND_BY_NAME
	var permission models.Permission
	var actionsJSON []byte

	err := r.db.QueryRow(
		ctx,
		query,
		name,
	).Scan(
		&permission.ID,
		&permission.Name,
		&permission.Code,
		&permission.Group,
		&actionsJSON,
		&permission.Slug,
		&permission.UpdatedBy,
		&permission.CreatedAt,
		&permission.UpdatedAt,
		&permission.Version,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Permission{}, pgx.ErrNoRows
		}

		return models.Permission{}, fmt.Errorf(
			"failed to find permission by name: %w",
			err,
		)
	}

	if len(actionsJSON) > 0 {
		if err := commonutils.FromJSON(actionsJSON, &permission.Actions); err != nil {
			return models.Permission{}, fmt.Errorf(
				"failed to decode permission actions: %w",
				err,
			)
		}
	}

	return permission, nil
}

// FindByGroup
//
// This function is used to find a permission by its group.
// It returns the permission and an error if the permission is not found.
func (r *PermissionRepository) FindByGroup(ctx context.Context, group string) (models.Permission, error) {

	var permission models.Permission
	var actionsJSON []byte

	err := r.db.QueryRow(
		ctx,
		utils.PERMISSION_FIND_BY_GROUP,
		group,
	).Scan(
		&permission.ID,
		&permission.Name,
		&permission.Code,
		&permission.Group,
		&actionsJSON,
		&permission.Slug,
		&permission.UpdatedBy,
		&permission.CreatedAt,
		&permission.UpdatedAt,
		&permission.Version,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Permission{}, pgx.ErrNoRows
		}

		return models.Permission{}, fmt.Errorf(
			"failed to find permission by group: %w",
			err,
		)
	}

	if len(actionsJSON) > 0 {
		if err := commonutils.FromJSON(actionsJSON, &permission.Actions); err != nil {
			return models.Permission{}, fmt.Errorf(
				"failed to decode permission actions: %w",
				err,
			)
		}
	}

	return permission, nil
}

// FindByRoleID returns the permissions linked to a role.
//
// Links live in the roles.permissions ID array, not a join table: the array
// is read first, then each permission is loaded by ID. Missing permissions
// are skipped.
func (r *PermissionRepository) FindByRoleID(ctx context.Context, roleID string) ([]models.Permission, error) {

	var permissionsJSON []byte
	err := r.db.QueryRow(ctx, "SELECT permissions FROM roles WHERE id = $1", roleID).Scan(&permissionsJSON)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []models.Permission{}, nil
		}
		return nil, fmt.Errorf(
			"failed to load role permissions: %w",
			err,
		)
	}

	var ids []string
	if len(permissionsJSON) > 0 {
		if err := commonutils.FromJSON(permissionsJSON, &ids); err != nil {
			return nil, fmt.Errorf(
				"failed to decode role permission ids: %w",
				err,
			)
		}
	}

	permissions := make([]models.Permission, 0, len(ids))
	for _, id := range ids {
		permission, err := r.FindByID(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			return nil, err
		}
		permissions = append(permissions, permission)
	}

	return permissions, nil
}
