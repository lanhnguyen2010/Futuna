#!/bin/sh
# Start the API server locally
set -e
set -a
. ./.env
set +a
migrate -path migrations -database "$DATABASE_URL" up
GO111MODULE=on go run ./cmd/web
