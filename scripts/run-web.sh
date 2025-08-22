#!/bin/sh
# Start the API server locally
set -e
GO111MODULE=on go run ./cmd/web
