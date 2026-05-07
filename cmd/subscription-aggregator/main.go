package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/spoddub/subscription-aggregator/docs"
	"github.com/spoddub/subscription-aggregator/internal/config"
	"github.com/spoddub/subscription-aggregator/internal/db"
	"github.com/spoddub/subscription-aggregator/internal/handler"
	"github.com/spoddub/subscription-aggregator/internal/logger"
	"github.com/spoddub/subscription-aggregator/internal/repository"
)

const (
	readHeaderTimeout = 5 * time.Second
	shutdownTimeout   = 10 * time.Second
)

// @title Subscription Aggregator API
// @version 1.0
// @description REST API service for managing online subscriptions and calculating total subscription cost for selected periods.
// @host localhost:8080
// @BasePath
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

	server := &http.Server{
		Addr:              address,
		Handler:           r,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	serverErrors := make(chan error, 1)

	go func() {
		appLogger.Info("server listening on " + address)

		serverErrors <- server.ListenAndServe()
	}()

	shutdownSignals := make(chan os.Signal, 1)
	signal.Notify(shutdownSignals, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(shutdownSignals)

	select {
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			appLogger.Error("failed to start server", "error", err)
			os.Exit(1)
		}

	case sig := <-shutdownSignals:
		appLogger.Info("received shutdown signal", "signal", sig)

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			appLogger.Error("graceful shutdown failed ", "error", err)

			if closeErr := server.Close(); closeErr != nil {
				appLogger.Error("forced server close failed", "error", closeErr)

				os.Exit(1)
			}
		}

		appLogger.Info("server stopped gracefully")
	}
}
