#!/bin/sh
# Ensure migrate CLI is installed and apply database migrations
set -e

# Determine installation directory for Go tools
GOBIN="$(go env GOBIN)"
if [ -z "$GOBIN" ]; then
  GOBIN="$(go env GOPATH)/bin"
fi

MIGRATE_BIN="$(command -v migrate || true)"
if [ -z "$MIGRATE_BIN" ]; then
  echo "Installing migrate CLI..."
  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
  MIGRATE_BIN="$GOBIN/migrate"
fi

"$MIGRATE_BIN" -path migrations -database "$DATABASE_URL" up
