#!/bin/sh
# Run analyzer once then start web server for development
set -e
ANALYZE_ON_START=1 GO111MODULE=on go run ./cmd/web
