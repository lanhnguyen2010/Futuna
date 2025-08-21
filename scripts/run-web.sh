#!/bin/sh
# Start the web dashboard locally
set -e
GO111MODULE=on go run ./cmd/web
