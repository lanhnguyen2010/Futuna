#!/bin/sh
# Start the Next.js front-end locally
set -e
set -a
. ./.env
set +a
./scripts/run-migrations.sh
npm --prefix web run dev
