#!/bin/sh
# Start the API server locally
set -e
set -a
. ./.env
set +a
./scripts/run-migrations.sh
GO111MODULE=on go run ./cmd/web
