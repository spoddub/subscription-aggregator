package main

import (
	"context"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/spoddub/subscription-aggregator/docs"
	"github.com/spoddub/subscription-aggregator/internal/config"
	"github.com/spoddub/subscription-aggregator/internal/db"
	"github.com/spoddub/subscription-aggregator/internal/handler"
	"github.com/spoddub/subscription-aggregator/internal/logger"
	"github.com/spoddub/subscription-aggregator/internal/repository"
)

func main() {
	dotenvErr := godotenv.Load()
	cfg := config.Load()
	appLogger := logger.New(cfg.LogLevel)

	if dotenvErr != nil {
		appLogger.Info("no .env file found, using environment variables")
	}

	appLogger.Info(
		"application starting",
		"port", cfg.Port,
		"log_level", cfg.LogLevel)

	ctx := context.Background()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		appLogger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	appLogger.Info("database connected")

	subscriptionRepo := repository.NewSubscriptionRepository(pool)

	r := handler.NewRouter(subscriptionRepo, appLogger)

	address := ":" + cfg.Port
	if err := r.Run(address); err != nil {
		appLogger.Error("failed to start server", "error", err)
		os.Exit(1)
	}
}
