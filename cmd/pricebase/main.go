package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	applogger "github.com/damolaoladipo/pricebase/internal/common/logger"
	"github.com/damolaoladipo/pricebase/internal/common/redis"
	"github.com/damolaoladipo/pricebase/internal/config"
	"github.com/damolaoladipo/pricebase/internal/config/seeds"
	"github.com/damolaoladipo/pricebase/internal/factory"
	"github.com/damolaoladipo/pricebase/internal/middleware"
	"github.com/damolaoladipo/pricebase/internal/utils"
	"github.com/fatih/color"
	"go.uber.org/zap"
)

func main() {

	defer applogger.Sync()

	ctx := context.Background()

	// laod ENV
	env, err := utils.NewENV()
	if err != nil {
		log.Fatal(err)
	}

	// Connect to database
	db, err := config.ConnectDB(env.POSTGRES_URL)
	if err != nil {
		applogger.DBLogger.Error("could not connect to database", zap.Error(err))
		log.Fatal(err)
	}
	defer db.Close()

	applogger.DBLogger.Info("database connected")
	color.New(color.FgCyan, color.Bold).Println("Database connected successfully")

	// Run migrations
	if err := config.RunMigrations(db); err != nil {
		applogger.DBLogger.Error("could not run migrations", zap.Error(err))
		log.Fatal(err)
	}
	applogger.DBLogger.Info("database migrations completed")
	color.New(color.FgGreen, color.Bold).Println("Database migrations completed")

	//Seed data
	if err := seeds.SeedData(ctx, db); err != nil {
		applogger.DBLogger.Error("could not seed data", zap.Error(err))
		log.Fatal(err)
	}

	// Connect to Redis
	if err := config.ConnectToRedis(ctx, env); err != nil {
		applogger.SystemLogger.Error("could not connect to redis", zap.Error(err))
		log.Fatal(err)
	}
	defer func() {
		if client := redis.REDIS.Client(); client != nil {
			_ = client.Close()
		}
	}()
	color.New(color.FgBlue, color.Bold).Println("Redis connected successfully")

	// Initialize factories
	// This is where you initialize the dependencies for the application
	factory.Factories(db)

	// Initialize rate limiters (must run before routes are registered)
	middleware.RateLimiters()

	// Start server with Socket.IO
	// Build the handler, then hand it to the Socket.IO-aware server.
	app := config.App(db)
	server := config.StartServer(app, env)

	// Listen returns immediately, so block until SIGINT/SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	applogger.SystemLogger.Info("shutting down")

	// Give in-flight requests 10s to finish before closing the pool and Redis.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		applogger.SystemLogger.Error("graceful shutdown failed", zap.Error(err))
	}

}

// Start server without Socket.IO
// app := config.App(db)
// atddr := fmt.Sprintf(":%d", env.APP_PORT)
// applogger.SystemLogger.Info("server listening",
// 	zap.String("env", env.APP_ENV),
// 	zap.Int("port", env.APP_PORT),
// )
// color.New(color.FgYellow, color.Bold).Printf("Server running in %s on port %d\n", env.APP_ENV, env.APP_PORT)
// if err := app.Run(addr); err != nil {
// 	applogger.SystemLogger.Error("server failed", zap.Error(err))
// 	log.Fatal(err)
// }
