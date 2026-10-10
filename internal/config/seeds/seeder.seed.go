package seeds

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	applogger "github.com/damolaoladipo/pricebase/internal/common/logger"
)

// SeedData runs the default Postgres seeders in FieldOS order:
// roles → permissions → role-permissions → users.
//
// It is intended to be called from application bootstrap after ConnectDB when
// seeding is enabled. ctx is passed to each seeder for cancellation. A nil
// pool makes each seeder a no-op.
//
// Errors from any seeder are returned to the caller so bootstrap can abort.
func SeedData(ctx context.Context, db *pgxpool.Pool) error {

	applogger.SystemLogger.Info("seed data started")

	if err := SeedRoles(ctx, db); err != nil {
		applogger.SystemLogger.Error("seed data failed", zap.Error(err))
		return err
	}

	if err := SeedPermissions(ctx, db); err != nil {
		applogger.SystemLogger.Error("seed data failed", zap.Error(err))
		return err
	}

	if err := SeedRolePermissions(ctx, db); err != nil {
		applogger.SystemLogger.Error("seed data failed", zap.Error(err))
		return err
	}

	if err := SeedUsers(ctx, db); err != nil {
		applogger.SystemLogger.Error("seed data failed", zap.Error(err))
		return err
	}

	if err := AttachSuperAdminRole(ctx, db); err != nil {
		applogger.SystemLogger.Error("seed data failed", zap.Error(err))
		return err
	}

	applogger.SystemLogger.Info("seed data completed")
	return nil
}
