#!/bin/sh
# Run the analyzer once to fetch analysis for all tickers
set -e
set -a
. ./.env
set +a
./scripts/run-migrations.sh
GO111MODULE=on go run ./cmd/analyzer
