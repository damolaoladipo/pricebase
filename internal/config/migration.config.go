package config

import (
	"context"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"

	"github.com/damolaoladipo/pricebase/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RunMigrations(db *pgxpool.Pool) error {
	ctx := context.Background()

	if err := createMigrationTable(ctx, db); err != nil {
		return err
	}

	files, err := fs.ReadDir(migrations.Files, ".")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		version, err := getMigrationVersion(file.Name())
		if err != nil {
			return fmt.Errorf("failed to get migration version: %w", err)
		}

		applied, err := isMigrationApplied(ctx, db, version)
		if err != nil {
			return fmt.Errorf("failed to check if migration is applied: %w", err)
		}

		if applied {
			continue
		}

		if err := applyMigration(ctx, db, version, file.Name()); err != nil {
			return fmt.Errorf("failed to apply migration: %w", err)
		}
	}

	return nil
}

func createMigrationTable(ctx context.Context, db *pgxpool.Pool) error {
	query := `
	CREATE TABLE IF NOT EXISTS migrations (
		version BIGINT PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	`

	_, err := db.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create migration table: %w", err)
	}

	return nil
}

func getMigrationVersion(filename string) (int64, error) {
	parts := strings.SplitN(filename, "_", 2)
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid migration filename: %s", filename)
	}

	version, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid migration version: %w", err)
	}

	return version, nil
}

func isMigrationApplied(ctx context.Context, db *pgxpool.Pool, version int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM migrations
			WHERE version = $1
		)
	`

	var applied bool
	err := db.QueryRow(ctx, query, version).Scan(&applied)
	if err != nil {
		return false, fmt.Errorf("failed to check if migration is applied %d: %w", version, err)
	}

	return applied, nil
}

func applyMigration(ctx context.Context, db *pgxpool.Pool, version int64, filename string) error {
	content, err := fs.ReadFile(migrations.Files, filename)
	if err != nil {
		return fmt.Errorf("failed to read migration %s: %w", filename, err)
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err = tx.Exec(ctx, string(content)); err != nil {
		return fmt.Errorf("failed to execute migration %s: %w", filename, err)
	}

	if _, err = tx.Exec(ctx, `INSERT INTO migrations (version) VALUES ($1)`, version); err != nil {
		return fmt.Errorf("failed to insert migration %d: %w", version, err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction %d: %w", version, err)
	}

	return nil
}
