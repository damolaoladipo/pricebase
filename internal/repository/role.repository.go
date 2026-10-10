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

type RoleRepository struct {
	db *pgxpool.Pool
}

func NewRoleRepository(db *pgxpool.Pool) *RoleRepository {
	return &RoleRepository{
		db: db,
	}
}

func (r *RoleRepository) CreateRole(ctx context.Context, data models.Role) (models.Role, error) {

	// utils.ROLE_INSERT is INSERT ... ON CONFLICT (name) DO NOTHING with no
	// RETURNING clause, so Exec (not QueryRow+Scan) is the correct call.
	// On conflict RowsAffected is 0 and we return the existing row.
	// IDs use utils.GenerateID (24-char hex), never database defaults.
	if data.ID == "" {
		data.ID = utils.GenerateID()
	}

	permissions := data.Permissions
	if permissions == nil {
		permissions = []string{}
	}
	permissionsJSON, err := commonutils.ToJSON(permissions)
	if err != nil {
		return models.Role{}, fmt.Errorf(
			"failed to marshal role permissions: %w",
			err,
		)
	}

	res, err := r.db.Exec(
		ctx,
		utils.ROLE_INSERT,
		data.ID,
		data.Name,
		data.Description,
		data.Slug,
		permissionsJSON,
	)
	if err != nil {
		return models.Role{}, fmt.Errorf(
			"failed to create role: %w",
			err,
		)
	}

	if res.RowsAffected() == 0 {
		return r.FindByName(ctx, data.Name)
	}

	return r.FindByName(ctx, data.Name)
}

// FindByID
//
// This function is used to find a role by its ID.
// It returns the role and an error if the role is not found.
func (r *RoleRepository) FindByID(ctx context.Context, id string) (models.Role, error) {

	query := utils.ROLE_FIND_BY_ID
	var role models.Role
	var permissionsJSON []byte

	err := r.db.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&role.ID,
		&role.Name,
		&role.Description,
		&role.Slug,
		&permissionsJSON,
		&role.CreatedAt,
		&role.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Role{}, pgx.ErrNoRows
		}

		return models.Role{}, fmt.Errorf(
			"failed to find role by ID: %w",
			err,
		)
	}

	if len(permissionsJSON) > 0 {
		if err := commonutils.FromJSON(permissionsJSON, &role.Permissions); err != nil {
			return models.Role{}, fmt.Errorf(
				"failed to decode role permissions: %w",
				err,
			)
		}
	}

	return role, nil
}

func (r *RoleRepository) FindByName(ctx context.Context, name string) (models.Role, error) {

	query := utils.ROLE_FIND_BY_NAME
	var role models.Role
	var permissionsJSON []byte

	err := r.db.QueryRow(
		ctx,
		query,
		name,
	).Scan(
		&role.ID,
		&role.Name,
		&role.Description,
		&role.Slug,
		&permissionsJSON,
		&role.CreatedAt,
		&role.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Role{}, pgx.ErrNoRows
		}

		return models.Role{}, fmt.Errorf(
			"failed to find role by name: %w",
			err,
		)
	}

	if len(permissionsJSON) > 0 {
		if err := commonutils.FromJSON(permissionsJSON, &role.Permissions); err != nil {
			return models.Role{}, fmt.Errorf(
				"failed to decode role permissions: %w",
				err,
			)
		}
	}

	return role, nil
}

// LinkPermission appends a permission ID to the role's permissions array.
// Links live in roles.permissions, not a join table. It returns 1 when the
// ID was added and 0 when it was already present.
func (r *RoleRepository) LinkPermission(ctx context.Context, roleID, permissionID string) (int64, error) {

	role, err := r.FindByID(ctx, roleID)
	if err != nil {
		return 0, err
	}

	for _, id := range role.Permissions {
		if id == permissionID {
			return 0, nil
		}
	}

	permissions := append(role.Permissions, permissionID)
	permissionsJSON, err := commonutils.ToJSON(permissions)
	if err != nil {
		return 0, fmt.Errorf(
			"failed to marshal role permissions: %w",
			err,
		)
	}

	res, err := r.db.Exec(
		ctx,
		utils.ROLE_UPDATE_PERMISSIONS,
		roleID,
		permissionsJSON,
	)
	if err != nil {
		return 0, fmt.Errorf(
			"failed to link role to permission: %w",
			err,
		)
	}

	return res.RowsAffected(), nil
}
