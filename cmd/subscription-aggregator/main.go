package main

import (
	"context"
	"log"

	"github.com/spoddub/subscription-aggregator/internal/repository"

	"github.com/joho/godotenv"
	"github.com/spoddub/subscription-aggregator/internal/config"
	"github.com/spoddub/subscription-aggregator/internal/db"

	"github.com/spoddub/subscription-aggregator/internal/handler"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	cfg := config.Load()
	ctx := context.Background()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer pool.Close()

	subscriptionRepo := repository.NewSubscriptionRepository(pool)

	r := handler.NewRouter(subscriptionRepo)

	address := ":" + cfg.Port
	if err := r.Run(address); err != nil {
		log.Fatalf("faild to run server: %v", err)
	}
}
