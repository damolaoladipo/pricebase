package config

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	applogger "github.com/damolaoladipo/pricebase/internal/common/logger"
	redis "github.com/damolaoladipo/pricebase/internal/common/redis"
	"github.com/damolaoladipo/pricebase/internal/utils"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func ConnectDB(databaseURL string) (*pgxpool.Pool, error) {

	if databaseURL == "" {
		return nil, errors.New("Database URL is not defined")
	}

	dbConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("Database config parsing failed: %w", err)
	}

	dbConfig.MaxConns = 20
	dbConfig.MinConns = 2
	dbConfig.MaxConnIdleTime = 5 * time.Minute
	dbConfig.MaxConnLifetime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		return nil, fmt.Errorf("Database pool creation failed: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		applogger.SystemLogger.Error("Database connection failed", zap.Error(err))

		return nil, fmt.Errorf("Database connection failed: %w", err)
	}

	return pool, nil
}

func ConnectToRedis(ctx context.Context, env *utils.ENV) error {

	port, err := strconv.Atoi(env.REDIS_PORT)
	if err != nil {
		return fmt.Errorf("invalid Redis port: %w", err)
	}

	host := env.REDIS_HOST
	if host == "" {
		host = "127.0.0.1"
	}

	password := env.REDIS_PASSWORD
	user := env.REDIS_USER

	platform := env.REDIS_PLATFORM
	if platform == "" {
		platform = "cloud-managed"
	}

	err = redis.REDIS.Connect(ctx, redis.Options{
		User:     user,
		Password: password,
		Host:     host,
		Port:     port,
		Platform: redis.Platform(platform),
	})

	if err != nil {
		applogger.SystemLogger.Error("could not connect to redis", zap.Error(err))
		return err
	}

	return nil
}
