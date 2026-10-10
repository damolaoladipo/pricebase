package seeds

import (
	"context"
	"errors"
	"fmt"

	"github.com/fatih/color"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	commontypes "github.com/damolaoladipo/common-go/types"
	commonutils "github.com/damolaoladipo/common-go/utils"
	applogger "github.com/damolaoladipo/pricebase/internal/common/logger"
	authdtos "github.com/damolaoladipo/pricebase/internal/dtos"
	"github.com/damolaoladipo/pricebase/internal/models"
	"github.com/damolaoladipo/pricebase/internal/repository"
)

// SeedRoles seeds the roles table in Postgres from the roles.json seed file
// under the project's _data directory (resolved via dataFile / readJSON).
//
// It runs as part of SeedData during application bootstrap so that named roles
// exist before permissions or user assignment logic needs them. Passing a nil
// *pgxpool.Pool is a no-op and the function returns nil immediately without
// side effects.
//
// Like SeedPermissions, all SQL lives in utils/queries.util.go and all DB
// access goes through repositories.RoleRepository — this file only orchestrates
// (read JSON, skip existing roles, insert new ones).
//
// ctx is passed through to every repository call for cancellation and deadline
// propagation. Each JSON entry unmarshals into SeedRoleDTO with name and
// description fields only. Entries are processed sequentially; the first hard
// failure aborts the loop.
//
// JSON read errors and Postgres errors are logged to stdout in red reverse video
// and to SystemLogger, then returned as wrapped errors. When one or more roles
// were inserted, a green reverse-video "Roles seeded successfully" message is
// printed.
func SeedRoles(ctx context.Context, db *pgxpool.Pool) error {

	if db == nil {
		return nil
	}

	var roles []authdtos.SeedRoleDTO
	if err := readJSON("roles.json", &roles); err != nil {
		color.New(color.FgRed, color.ReverseVideo).Println(err.Error())
		applogger.SystemLogger.Error("seed roles failed", zap.Error(err))
		return fmt.Errorf("failed to read roles seed: %w", err)
	}

	seeded := 0
	roleRepo := repository.NewRoleRepository(db)

	for _, item := range roles {

		// skip roles that already exist
		_, err := roleRepo.FindByName(ctx, item.Name)
		if err == nil {
			continue
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			color.New(color.FgRed, color.ReverseVideo).Println(err.Error())
			applogger.SystemLogger.Error(
				"seed roles failed",
				zap.String("role", item.Name),
				zap.Error(err),
			)
			return fmt.Errorf("failed to find role %q: %w", item.Name, err)
		}

		_, err = roleRepo.CreateRole(ctx, models.Role{
			Name:        item.Name,
			Description: item.Description,
			Slug:        commonutils.Slugit(commontypes.SlugitDTO{Data: item.Name}),
		})

		if err != nil {

			color.New(color.FgRed, color.ReverseVideo).Println(err.Error())
			applogger.SystemLogger.Error(
				"seed roles failed",
				zap.String("role", item.Name),
				zap.Error(err),
			)
			return fmt.Errorf("failed to seed role %q: %w", item.Name, err)
		}

		seeded++
	}

	if seeded > 0 {
		color.New(color.FgGreen, color.ReverseVideo).Println("Roles seeded successfully")
		applogger.SystemLogger.Info("Roles seeded successfully", zap.Int("count", seeded))
	}

	return nil
}
