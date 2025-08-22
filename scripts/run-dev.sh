#!/bin/sh
# Run analyzer then start API and Next.js front-end for development
set -e
set -a
. ./.env
set +a
migrate -path migrations -database "$DATABASE_URL" up
ANALYZE_ON_START=1 GO111MODULE=on go run ./cmd/web &
npm --prefix web run dev
