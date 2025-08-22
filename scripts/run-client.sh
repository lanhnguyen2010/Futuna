#!/bin/sh
# Start the Next.js front-end locally
set -e
set -a
. ./.env
set +a
migrate -path migrations -database "$DATABASE_URL" up
npm --prefix web run dev
