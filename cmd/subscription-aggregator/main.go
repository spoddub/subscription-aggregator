package main

import (
	"context"
	"log"

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

	r := handler.NewRouter()

	address := ":" + cfg.Port
	if err := r.Run(address); err != nil {
		log.Fatalf("faild to run server: %v", err)
	}
}
