#!/bin/bash
# Run app with oexctl proxy
# Usage: ./dev.sh <app-name> [port]

set -e

APP_NAME="$1"
PORT="${2:-3000}"

if [ -z "$APP_NAME" ]; then
    echo "Error: App name is required"
    echo "Usage: ./dev.sh <app-name> [port]"
    exit 1
fi

echo "Starting ${APP_NAME}..."
echo "Using oexctl proxy: https://${APP_NAME}.localhost:1355"
oexctl proxy run "${APP_NAME}" -- npx next dev --turbopack --port "${PORT}"
