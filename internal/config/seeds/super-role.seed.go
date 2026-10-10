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
	"github.com/damolaoladipo/pricebase/internal/repository"
	"github.com/damolaoladipo/pricebase/internal/services"
	"github.com/damolaoladipo/pricebase/internal/utils"
)

func AttachSuperAdminRole(ctx context.Context, db *pgxpool.Pool) error {

	if db == nil {
		return nil
	}

	env := utils.Current()
	if env == nil {
		var err error
		env, err = utils.NewENV()
		if err != nil {
			return nil
		}
	}
	if env.SUPERADMIN_EMAIL == "" {
		return nil
	}

	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	permRepo := repository.NewPermissionRepository(db)
	authService := services.NewAuthService(roleRepo, permRepo, userRepo, nil)
	userService := services.NewUserService(userRepo, services.NewSystemService(), roleRepo)

	superadmin, err := userRepo.FindUserByEmail(ctx, env.SUPERADMIN_EMAIL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		color.New(color.FgRed, color.ReverseVideo).Println(err.Error())
		applogger.SystemLogger.Error("attach superadmin role failed", zap.Error(err))
		return fmt.Errorf("failed to find superadmin %q: %w", env.SUPERADMIN_EMAIL, err)
	}

	userRole, err := roleRepo.FindByName(ctx, "user")
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		color.New(color.FgRed, color.ReverseVideo).Println(err.Error())
		applogger.SystemLogger.Error("attach superadmin role failed", zap.Error(err))
		return fmt.Errorf("failed to find role %q: %w", "user", err)
	}

	superRole, err := roleRepo.FindByName(ctx, "superadmin")
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		color.New(color.FgRed, color.ReverseVideo).Println(err.Error())
		applogger.SystemLogger.Error("attach superadmin role failed", zap.Error(err))
		return fmt.Errorf("failed to find role %q: %w", "superadmin", err)
	}

	hasUserRole, err := userRepo.HasRole(ctx, superadmin.ID, userRole.ID)
	if err != nil {
		color.New(color.FgRed, color.ReverseVideo).Println(err.Error())
		applogger.SystemLogger.Error("attach superadmin role failed", zap.Error(err))
		return err
	}

	hasSuperRole, err := userRepo.HasRole(ctx, superadmin.ID, superRole.ID)
	if err != nil {
		color.New(color.FgRed, color.ReverseVideo).Println(err.Error())
		applogger.SystemLogger.Error("attach superadmin role failed", zap.Error(err))
		return err
	}

	if hasUserRole || hasSuperRole {
		return nil
	}

	// attach the user role to the superadmin

	if _, err := userRepo.AttachRole(ctx, superadmin.ID, userRole.ID); err != nil {
		color.New(color.FgRed, color.ReverseVideo).Println(err.Error())
		applogger.SystemLogger.Error("attach superadmin role failed", zap.Error(err))
		return fmt.Errorf("failed to attach role %q: %w", "user", err)
	}
	if _, err := userRepo.AttachRole(ctx, superadmin.ID, superRole.ID); err != nil {
		color.New(color.FgRed, color.ReverseVideo).Println(err.Error())
		applogger.SystemLogger.Error("attach superadmin role failed", zap.Error(err))
		return fmt.Errorf("failed to attach role %q: %w", "superadmin", err)
	}

	// derive + store the permission snapshot (TS: UpdateRolesAndPermissions)
	uPerms, err := authService.ExtractPermissions(ctx, userRole.ID)
	if err != nil {
		color.New(color.FgRed, color.ReverseVideo).Println(err.Error())
		applogger.SystemLogger.Error("attach superadmin role failed", zap.Error(err))
		return err
	}
	sPerms, err := authService.ExtractPermissions(ctx, superRole.ID)
	if err != nil {
		color.New(color.FgRed, color.ReverseVideo).Println(err.Error())
		applogger.SystemLogger.Error("attach superadmin role failed", zap.Error(err))
		return err
	}

	permissions := make([]utils.UserPermission, 0, len(uPerms)+len(sPerms))
	permissions = append(permissions, uPerms...)
	permissions = append(permissions, sPerms...)

	if err := userRepo.UpdateRolesPermissions(ctx, superadmin.ID, []string{userRole.ID, superRole.ID}, permissions); err != nil {
		color.New(color.FgRed, color.ReverseVideo).Println(err.Error())
		applogger.SystemLogger.Error("attach superadmin role failed", zap.Error(err))
		return err
	}

	// encrypt saved password (runs once: guarded above, so re-seeds never
	// double-encrypt)
	if _, err := userService.EncryptUserPassword(ctx, &superadmin, superadmin.SavedPassword); err != nil {
		color.New(color.FgRed, color.ReverseVideo).Println(err.Error())
		applogger.SystemLogger.Error("attach superadmin role failed", zap.Error(err))
		return err
	}

	// generate API key
	if _, err := authService.GenerateAPIKey(ctx, &superadmin); err != nil {
		color.New(color.FgRed, color.ReverseVideo).Println(err.Error())
		applogger.SystemLogger.Error("attach superadmin role failed", zap.Error(err))
		return err
	}

	color.New(color.FgMagenta, color.ReverseVideo).Println("Superadmin role attached successfully")
	applogger.SystemLogger.Info("Superadmin role attached successfully")
	return nil

}
