package seeds

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	applogger "github.com/damolaoladipo/pricebase/internal/common/logger"
	userdtos "github.com/damolaoladipo/pricebase/internal/dtos"
	"github.com/damolaoladipo/pricebase/internal/models"
	"github.com/damolaoladipo/pricebase/internal/repository"
)

// SeedUsers seeds the users table from users.json, matching FieldOS seedUsers.
//
// Like SeedPermissions, all SQL lives in utils/queries.util.go and all DB
// access goes through repositories.UserRepository — this file only orchestrates
// (read JSON, skip existing emails, insert new users).
//
// Existing emails are skipped. New users are inserted via
// UserRepository.CreateUser with the SeedUserDTO flags. A missing users.json
// is intentional and returns nil. Passing a nil pool is a no-op.
func SeedUsers(ctx context.Context, db *pgxpool.Pool) error {

	if db == nil {
		return nil
	}

	var users []userdtos.SeedUserDTO
	err := readJSON("users.json", &users)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		color.New(color.FgRed, color.ReverseVideo).Println(err.Error())
		applogger.SystemLogger.Error("seed users failed", zap.Error(err))
		return fmt.Errorf("failed to read users seed: %w", err)
	}

	seeded := 0
	userRepo := repository.NewUserRepository(db)

	for _, item := range users {
		if item.Email == "" {
			continue
		}

		// check if user exists
		_, err := userRepo.FindUserByEmail(ctx, item.Email)
		if err == nil {
			continue
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			color.New(color.FgRed, color.ReverseVideo).Println(err.Error())
			applogger.SystemLogger.Error("seed users failed", zap.Error(err))
			return fmt.Errorf("failed to check user %q: %w", item.Email, err)
		}

		_, err = userRepo.CreateUser(ctx, models.User{
			FirstName:     item.FirstName,
			LastName:      item.LastName,
			Email:         item.Email,
			Password:      item.Password,
			SavedPassword: item.SavedPassword,
			UserType:      item.UserType,
			IsSuper:       item.IsSuper,
			IsAdmin:       item.IsAdmin,
			IsBusiness:    item.IsBusiness,
			IsUser:        item.IsUser,
			IsTeam:        item.IsTeam,
			IsActivated:   item.IsActivated,
			IsActive:      item.IsActive,
			IsWelcomed:    item.IsWelcomed,
		})

		if err != nil {
			color.New(color.FgRed, color.ReverseVideo).Println(err.Error())
			applogger.SystemLogger.Error(
				"seed users failed",
				zap.String("email", item.Email),
				zap.Error(err),
			)
			return fmt.Errorf("failed to insert user %q: %w", item.Email, err)
		}

		seeded++
	}

	if seeded > 0 {
		color.New(color.FgGreen, color.ReverseVideo).Println("Users seeded successfully")
		applogger.SystemLogger.Info("Users seeded successfully", zap.Int("count", seeded))
	}

	return nil
}
