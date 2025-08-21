#!/bin/sh
# Run the analyzer once to fetch analysis for all tickers
set -e
GO111MODULE=on go run ./cmd/analyzer
