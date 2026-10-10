package seeds

import (
	"context"
	"errors"
	"fmt"

	"github.com/fatih/color"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	applogger "github.com/damolaoladipo/pricebase/internal/common/logger"
	"github.com/damolaoladipo/pricebase/internal/dtos"
	"github.com/damolaoladipo/pricebase/internal/repository"
	"github.com/damolaoladipo/pricebase/internal/services"
)

// SeedRolePermissions links roles to permission groups from
// role-permissions.json, matching FieldOS seedRolePermissions.
//
// Like SeedPermissions, all SQL lives in utils/queries.util.go and all DB
// access goes through repositories (RoleRepository + PermissionRepository) —
// this file only orchestrates (read JSON, resolve role/permission rows,
// validate actions, insert links).
//
// For each entry it finds the role by name, finds each permission by group,
// checks that requested actions exist on that permission, then appends the
// permission ID to the role's permissions array when the link is new.
// Passing a nil pool is a no-op.
func SeedRolePermissions(ctx context.Context, db *pgxpool.Pool) error {

	if db == nil {
		return nil
	}

	var rolePermissions []dtos.SeedRolePermDTO
	if err := readJSON("role-permissions.json", &rolePermissions); err != nil {
		color.New(color.FgRed, color.ReverseVideo).Println(err.Error())
		applogger.SystemLogger.Error("seed role-permissions failed", zap.Error(err))
		return fmt.Errorf("failed to read role-permissions seed: %w", err)
	}

	roleRepo := repository.NewRoleRepository(db)
	permRepo := repository.NewPermissionRepository(db)
	authService := services.NewAuthService(roleRepo, permRepo, nil, nil)
	seeded := 0

	// loop through the role permissions
	for _, item := range rolePermissions {
		// find the role by name
		role, err := roleRepo.FindByName(ctx, item.Role)
		// if the role is not found, continue
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			color.New(color.FgRed, color.ReverseVideo).Println(err.Error())
			applogger.SystemLogger.Error(
				"seed role-permissions failed",
				zap.String("role", item.Role),
				zap.Error(err),
			)
			return fmt.Errorf("failed to find role %q: %w", item.Role, err)
		}

		// loop through the permissions
		for _, perm := range item.Permissions {

			// find the permission by group (actions come populated from
			// the permissions.actions JSONB column)
			permission, err := permRepo.FindByGroup(ctx, perm.Group)
			// if the permission is not found, continue
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					continue
				}
				color.New(color.FgRed, color.ReverseVideo).Println(err.Error())
				applogger.SystemLogger.Error(
					"seed role-permissions failed",
					zap.String("role", item.Role),
					zap.String("group", perm.Group),
					zap.Error(err),
				)
				return fmt.Errorf(
					"failed to find permission group %q: %w",
					perm.Group,
					err,
				)
			}

			// extract the permission actions
			available := authService.ExtractActions(permission.Actions)

			// validate the actions
			if !authService.ValidateActions(available, perm.Actions) {
				continue
			}

			// insert the role permission
			rows, err := roleRepo.LinkPermission(ctx, role.ID, permission.ID)
			if err != nil {
				color.New(color.FgRed, color.ReverseVideo).Println(err.Error())
				applogger.SystemLogger.Error(
					"seed role-permissions failed",
					zap.String("role", item.Role),
					zap.String("group", perm.Group),
					zap.Error(err),
				)
				return fmt.Errorf(
					"failed to link role %q to permission %q: %w",
					item.Role,
					perm.Group,
					err,
				)
			}

			seeded += int(rows)
		}
	}

	if seeded > 0 {
		color.New(color.FgGreen, color.ReverseVideo).Println("Role-Permissions seeded successfully")
		applogger.SystemLogger.Info("Role-Permissions seeded successfully", zap.Int("count", seeded))
	}

	return nil
}
