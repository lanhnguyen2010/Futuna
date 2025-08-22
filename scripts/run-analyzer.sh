#!/bin/sh
# Run the analyzer once to fetch analysis for all tickers
set -e
set -a
. ./.env
set +a
migrate -path migrations -database "$DATABASE_URL" up
GO111MODULE=on go run ./cmd/analyzer
