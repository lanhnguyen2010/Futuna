#!/bin/sh
# Start the Next.js front-end locally
set -e
set -a
. ./.env
set +a
npm --prefix web install
npm --prefix web run dev
