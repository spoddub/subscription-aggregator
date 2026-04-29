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