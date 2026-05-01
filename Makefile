tidy:
	go mod tidy
	go fmt ./...
	go vet ./...
test:
	go test ./...
lint:
	golangci-lint run
lint-fix:
	golangci-lint run --fix
docker-up:
	docker compose up -d
docker-down:
	docker compose down
docker-logs:
	docker compose logs -f
check:
	docker compose up -d
	sleep 5
	goose -dir migrations postgres "postgres://postgres:postgres@localhost:5432/subscription_aggregator?sslmode=disable" up
	go run ./cmd/subscription-aggregator
ping:
	curl http://localhost:8080/ping
dev:
	docker compose up -d
	sleep 5
	goose -dir migrations postgres "postgres://postgres:postgres@localhost:5432/subscription_aggregator?sslmode=disable" up
	air -c .air.toml