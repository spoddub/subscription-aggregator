![Tests](https://github.com/spoddub/subscription-aggregator/actions/workflows/tests.yml/badge.svg)
![Linter](https://github.com/spoddub/subscription-aggregator/actions/workflows/linter.yml/badge.svg)

## Subscription Aggregator

`Subscription Aggregator` is a REST API service written in Go (Gin).  
It stores online subscriptions and calculates the total subscription cost for a selected period.

The service works with monthly subscriptions. Dates are passed in `MM-YYYY` format, for example `07-2025`.

---

## Features

- Create subscription records
- List all subscriptions
- Get subscription by id
- Update subscription by id
- Delete subscription by id
- Calculate total cost for a selected period
- Optional filtering by `user_id`
- Optional filtering by `service_name`
- PostgreSQL storage
- Database migrations with goose
- Request validation for UUID, price and date format
- Unit tests for date parsing and request validation
- Swagger API documentation
- Structured application logs with `slog`
- Local development with Docker Compose
- Full Docker run with app and PostgreSQL
- Makefile commands for common development tasks

---

## Tools used

| Tool | What it is used for |
| --- | --- |
| [Go](https://go.dev/) | Language and toolchain |
| [Gin](https://github.com/gin-gonic/gin) | HTTP router and middleware |
| [pgx](https://github.com/jackc/pgx) | PostgreSQL driver and connection pool |
| [PostgreSQL](https://www.postgresql.org/) | Persistent storage |
| [goose](https://github.com/pressly/goose) | Database migrations |
| [godotenv](https://github.com/joho/godotenv) | Loading local `.env` files |
| [google/uuid](https://github.com/google/uuid) | UUID parsing and validation |
| [slog](https://pkg.go.dev/log/slog) | Structured application logging |
| [swaggo/swag](https://github.com/swaggo/swag) | Swagger documentation generation |
| [gin-swagger](https://github.com/swaggo/gin-swagger) | Swagger UI for Gin |
| [Docker](https://www.docker.com/) | Containerized application runtime |
| [Docker Compose](https://docs.docker.com/compose/) | Local app and PostgreSQL environment |
| [Make](https://www.gnu.org/software/make/) | Common development commands |
| [golangci-lint](https://golangci-lint.run/) | Go linter |
| [Air](https://github.com/air-verse/air) | Hot reload for local development |
| [testing](https://pkg.go.dev/testing) | Unit tests |

---

## API

### Healthcheck

- `GET /ping` - check that the server is running

Example:

```bash
curl http://localhost:8080/ping
```

Example response:

```json
{
  "status": "pong"
}
```

---

### Swagger

- `GET /swagger/index.html` - interactive Swagger API documentation

Open in browser:

```text
http://localhost:8080/swagger/index.html
```

Generate Swagger docs:

```bash
make swagger
```

---

### Subscriptions

- `GET /api/subscriptions` - list subscriptions
- `POST /api/subscriptions` - create a subscription
- `GET /api/subscriptions/:id` - get subscription by id
- `PUT /api/subscriptions/:id` - update subscription by id
- `DELETE /api/subscriptions/:id` - delete subscription by id
- `GET /api/subscriptions/total` - calculate total cost for a selected period

---

## Date format

The API accepts subscription dates in `MM-YYYY` format.

Examples:

```text
01-2025
07-2025
12-2025
```

Internally, these values are stored as PostgreSQL `DATE` values using the first day of the month.

Example:

```text
07-2025 -> 2025-07-01
```

---

## Create subscription

### Request

```bash
curl -X POST http://localhost:8080/api/subscriptions \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "Yandex Plus",
    "price": 400,
    "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
    "start_date": "07-2025"
  }'
```

### Request body

```json
{
  "service_name": "Yandex Plus",
  "price": 400,
  "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
  "start_date": "07-2025",
  "end_date": "12-2025"
}
```

`end_date` is optional. If it is not provided, the subscription is treated as active.

### Example response

```json
{
  "id": 1,
  "service_name": "Yandex Plus",
  "price": 400,
  "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
  "start_date": "07-2025",
  "created_at": "2026-04-29T18:32:00Z",
  "updated_at": "2026-04-29T18:32:00Z"
}
```

---

## List subscriptions

```bash
curl http://localhost:8080/api/subscriptions
```

Example response:

```json
{
  "subscriptions": [
    {
      "id": 1,
      "service_name": "Yandex Plus",
      "price": 400,
      "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
      "start_date": "07-2025",
      "created_at": "2026-04-29T18:32:00Z",
      "updated_at": "2026-04-29T18:32:00Z"
    }
  ]
}
```

---

## Get subscription by id

```bash
curl http://localhost:8080/api/subscriptions/1
```

Example response:

```json
{
  "id": 1,
  "service_name": "Yandex Plus",
  "price": 400,
  "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
  "start_date": "07-2025",
  "created_at": "2026-04-29T18:32:00Z",
  "updated_at": "2026-04-29T18:32:00Z"
}
```

---

## Update subscription

```bash
curl -X PUT http://localhost:8080/api/subscriptions/1 \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "Yandex Plus",
    "price": 500,
    "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
    "start_date": "07-2025",
    "end_date": "12-2025"
  }'
```

Example response:

```json
{
  "id": 1,
  "service_name": "Yandex Plus",
  "price": 500,
  "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
  "start_date": "07-2025",
  "end_date": "12-2025",
  "created_at": "2026-04-29T18:32:00Z",
  "updated_at": "2026-04-29T18:40:00Z"
}
```

---

## Delete subscription

```bash
curl -X DELETE http://localhost:8080/api/subscriptions/1
```

Successful response:

```text
204 No Content
```

---

## Calculate total cost

Endpoint:

```text
GET /api/subscriptions/total
```

Required query parameters:

- `from` - start of the period in `MM-YYYY` format
- `to` - end of the period in `MM-YYYY` format

Optional query parameters:

- `user_id` - filter by user UUID
- `service_name` - filter by subscription name

Example:

```bash
curl "http://localhost:8080/api/subscriptions/total?from=07-2025&to=09-2025"
```

Example response:

```json
{
  "from": "07-2025",
  "to": "09-2025",
  "user_id": "",
  "service_name": "",
  "total": 1200
}
```

Calculation example:

```text
Subscription:
price = 400
start_date = 07-2025
end_date = empty

Requested period:
from = 07-2025
to = 09-2025

Months:
07-2025, 08-2025, 09-2025

Total:
3 * 400 = 1200
```

Example with filters:

```bash
curl "http://localhost:8080/api/subscriptions/total?from=07-2025&to=09-2025&user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba&service_name=Yandex%20Plus"
```

---

## Validation and errors

The API returns `400 Bad Request` for invalid input.

Invalid JSON:

```json
{
  "error": "invalid request body"
}
```

Invalid UUID:

```json
{
  "error": "invalid user_id"
}
```

Invalid date format:

```json
{
  "error": "invalid start_date, expected format MM-YYYY"
}
```

Invalid price:

```json
{
  "error": "price must be greater than zero"
}
```

Invalid period:

```json
{
  "error": "from must be before or equal to to"
}
```

Subscription not found:

```json
{
  "error": "subscription not found"
}
```

---

## Installation and local development

### Requirements

- Go 1.25+
- Docker
- Docker Compose
- `make`
- `goose`
- `swag`
- `golangci-lint`, optional for linting
- `air`, optional for hot reload with `make dev`

Install goose:

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

Install swag:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

Install Air, optional:

```bash
go install github.com/air-verse/air@latest
```

---

## Environment variables

Create `.env` in the project root for local development:

```env
PORT=8080
DATABASE_URL=postgres://postgres:postgres@localhost:5432/subscription_aggregator?sslmode=disable
LOG_LEVEL=debug
```

Environment variables:

- `PORT` - HTTP server port, defaults to `8080`
- `DATABASE_URL` - PostgreSQL connection string
- `LOG_LEVEL` - application log level, defaults to `info`

For Docker Compose, the app uses the PostgreSQL service name as host:

```env
DATABASE_URL=postgres://postgres:postgres@postgres:5432/subscription_aggregator?sslmode=disable
```

---

## Run locally

Start PostgreSQL, apply migrations and run the application locally:

```bash
make check
```

`make check` starts PostgreSQL, applies migrations and runs the Go app on the host machine.

The app listens on:

```text
http://localhost:8080
```

Check that the server is running from another terminal:

```bash
make ping
```

---

## Development mode

If Air is installed, the app can be started in development mode with hot reload:

```bash
make dev
```

This command starts PostgreSQL, applies migrations and runs the app through Air.

---

## Full Docker run

Build and run the app together with PostgreSQL:

```bash
make docker-app
```

This command runs:

```text
docker compose up --build
```

The Docker app container applies database migrations automatically before starting the server.

The app listens on:

```text
http://localhost:8080
```

Check the app from another terminal:

```bash
make ping
```

Open Swagger UI:

```text
http://localhost:8080/swagger/index.html
```

Stop Docker Compose services:

```bash
make docker-down
```

---

## Docker commands

Start PostgreSQL only:

```bash
make docker-up
```

Build and run PostgreSQL with the app:

```bash
make docker-app
```

Stop Docker Compose services:

```bash
make docker-down
```

Show Docker logs:

```bash
make docker-logs
```

---

## Migrations

For local development, migrations are applied automatically by:

```bash
make check
```

and:

```bash
make dev
```

For full Docker run, migrations are applied automatically by the app container entrypoint.

You can also apply migrations manually:

```bash
goose -dir migrations postgres "postgres://postgres:postgres@localhost:5432/subscription_aggregator?sslmode=disable" up
```

Check migration status manually:

```bash
goose -dir migrations postgres "postgres://postgres:postgres@localhost:5432/subscription_aggregator?sslmode=disable" status
```

Rollback the last migration manually:

```bash
goose -dir migrations postgres "postgres://postgres:postgres@localhost:5432/subscription_aggregator?sslmode=disable" down
```

---

## Swagger

Generate Swagger documentation:

```bash
make swagger
```

Run the application:

```bash
make check
```

Open Swagger UI:

```text
http://localhost:8080/swagger/index.html
```

---

## Run tests

Run all tests:

```bash
make test
```

Or directly:

```bash
go test ./...
```

---

## Code quality

Format code, tidy dependencies and run `go vet`:

```bash
make tidy
```

Run linter:

```bash
make lint
```

Run linter with automatic fixes:

```bash
make lint-fix
```

---

## Common commands

```bash
make tidy
```

Runs:

```text
go mod tidy
go fmt ./...
go vet ./...
```

```bash
make test
```

Runs:

```text
go test ./...
```

```bash
make lint
```

Runs:

```text
golangci-lint run
```

```bash
make lint-fix
```

Runs:

```text
golangci-lint run --fix
```

```bash
make docker-up
```

Runs:

```text
docker compose up -d postgres
```

```bash
make docker-app
```

Runs:

```text
docker compose up --build
```

```bash
make docker-down
```

Runs:

```text
docker compose down
```

```bash
make docker-logs
```

Runs:

```text
docker compose logs -f
```

```bash
make check
```

Runs:

```text
docker compose up -d postgres
sleep 5
goose -dir migrations postgres "postgres://postgres:postgres@localhost:5432/subscription_aggregator?sslmode=disable" up
go run ./cmd/subscription-aggregator
```

```bash
make ping
```

Runs:

```text
curl http://localhost:8080/ping
```

```bash
make dev
```

Runs:

```text
docker compose up -d postgres
sleep 5
goose -dir migrations postgres "postgres://postgres:postgres@localhost:5432/subscription_aggregator?sslmode=disable" up
air -c .air.toml
```

```bash
make swagger
```

Runs:

```text
swag init -g ./cmd/subscription-aggregator/main.go -o ./docs --parseInternal
```

---

## Database schema

The service uses one main table: `subscriptions`.

Main fields:

- `id`
- `service_name`
- `price`
- `user_id`
- `start_date`
- `end_date`
- `created_at`
- `updated_at`

Important constraints:

- `service_name` must not be empty
- `price` must be greater than zero
- `start_date` is required
- `end_date` can be empty
- `end_date` must be greater than or equal to `start_date`