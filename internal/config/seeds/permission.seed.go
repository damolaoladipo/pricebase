package seeds

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	commonrandom "github.com/damolaoladipo/common-go/random"
	commontypes "github.com/damolaoladipo/common-go/types"
	commonutils "github.com/damolaoladipo/common-go/utils"
	applogger "github.com/damolaoladipo/pricebase/internal/common/logger"
	"github.com/damolaoladipo/pricebase/internal/dtos"
	"github.com/damolaoladipo/pricebase/internal/models"
	"github.com/damolaoladipo/pricebase/internal/repository"
	"github.com/damolaoladipo/pricebase/internal/utils"
)

// SeedPermissions seeds the permissions table in Postgres from the
// permissions.json seed file under the project's _data directory
// (resolved via dataFile / readJSON).
//
// It is intended to run during application bootstrap (via SeedData after
// ConnectDB) so that RBAC permission groups exist before identity routes rely
// on them. Passing a nil *pgxpool.Pool is a no-op and the function returns nil
// immediately.
//
// All SQL lives in utils/queries.util.go and all DB access goes through
// repository.PermissionRepository — this file only orchestrates (read JSON,
// skip existing groups, insert permission with its actions JSON).
//
// ctx is forwarded to every repository call for cancellation and deadline
// propagation. db is the shared application pool. For each seed entry it skips
// existing groups, then inserts the permission row with Actions stored as JSON
// in permissions.actions — SeedRolePermissions reads the names back from the
// populated permission struct.
//
// Failures are logged in red reverse video and returned as wrapped errors.
// When at least one permission group was seeded, a green success message is
// printed.
func SeedPermissions(ctx context.Context, db *pgxpool.Pool) error {

	if db == nil {
		return nil
	}

	var permissions []dtos.SeedPermDTO
	if err := readJSON("permissions.json", &permissions); err != nil {
		color.New(color.FgRed, color.ReverseVideo).Println(err.Error())
		applogger.SystemLogger.Error("seed permissions failed", zap.Error(err))
		return fmt.Errorf("failed to read permissions seed: %w", err)
	}

	seeded := 0
	permRepo := repository.NewPermissionRepository(db)

	for _, item := range permissions {

		// check if the permission group already exists
		_, err := permRepo.FindByGroup(ctx, item.Group)
		if err == nil {
			continue
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			color.New(color.FgRed, color.ReverseVideo).Println(err.Error())
			applogger.SystemLogger.Error("seed permissions failed", zap.Error(err))
			return fmt.Errorf("failed to find permissions by group %q: %w", item.Group, err)
		}

		// create the permission code
		suffix, err := commonrandom.RandomAlpha(6)
		if err != nil {
			color.New(color.FgRed, color.ReverseVideo).Println(err.Error())
			applogger.SystemLogger.Error("seed permissions failed", zap.Error(err))
			return fmt.Errorf("failed to generate permission code for %q: %w", item.Group, err)
		}
		code := strings.ToUpper(fmt.Sprintf("%s%s", utils.PrefixPermission, suffix))
		name := fmt.Sprintf("%s Permissions", commonutils.Capitalize(item.Group))
		slug := commonutils.Slugit(commontypes.SlugitDTO{
			Data: item.Group + "-permission",
		})

		// prepare the permission payload data
		permission := models.Permission{
			Name:      name,
			Code:      code,
			Group:     item.Group,
			Actions:   item.Actions,
			Slug:      slug,
			UpdatedBy: "",
		}

		// insert the permission (actions are stored as JSON in
		// permissions.actions by CreatePermission)

		_, err = permRepo.CreatePermission(ctx, permission)
		if err != nil {
			color.New(color.FgRed, color.ReverseVideo).Println(err.Error())
			applogger.SystemLogger.Error("seed permissions failed", zap.Error(err))
			return fmt.Errorf("failed to insert permission %q: %w", item.Group, err)
		}

		seeded++
	}

	if seeded > 0 {
		color.New(color.FgGreen, color.ReverseVideo).Println("Permissions seeded successfully")
		applogger.SystemLogger.Info("Permissions seeded successfully", zap.Int("count", seeded))
	}

	return nil
}
