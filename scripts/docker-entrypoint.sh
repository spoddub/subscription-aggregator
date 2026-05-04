#!/bin/sh

set -e

: "${DATABASE_URL:?DATABASE_URL is required}"

echo "running database migrations"
goose -dir /app/migrations postgres "$DATABASE_URL" up

echo "starting application"
exec /app/subscription-aggregator