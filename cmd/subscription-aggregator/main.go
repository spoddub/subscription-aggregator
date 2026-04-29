package main

import (
	"log"

	"github.com/spoddub/subscription-aggregator/internal/handler"
)

const address = ":8080"

func main() {
	r := handler.NewRouter()

	if err := r.Run(address); err != nil {
		log.Fatalf("faild to run server: %v", err)
	}
}
