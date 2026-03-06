#!/bin/bash
# Run dashboard app with oexctl proxy (falls back to regular dev if oexctl not installed)

set -e

# Get project name from parent directory
PROJECT_NAME=$(basename "$(dirname "$(pwd)")")
APP_NAME="${PROJECT_NAME}-dashboard"

echo "Starting ${APP_NAME}..."

# Check if oexctl is installed
if command -v oexctl >/dev/null 2>&1; then
    echo "Using oexctl proxy: https://${APP_NAME}.localhost:1355"
    oexctl proxy run "${APP_NAME}" -- npx next dev --turbopack "$@"
else
    echo "oexctl not found, using regular dev server on port 3001"
    echo "Install oexctl for better URLs: openenvx install"
    npx next dev --turbopack --port 3001 "$@"
fi
