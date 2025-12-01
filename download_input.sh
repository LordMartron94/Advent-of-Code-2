#!/usr/bin/env bash
set -euo pipefail

# Always run from repo root
cd "$(dirname "$0")"

# Ensure build directory exists
mkdir -p build

# Default entrypoint is orchestrator main.go at root
ENTRYPOINT="./tools/input/main.go"
BINARY="./build/input_downloader"

echo "🔹 Building Analyzer project..."
echo "   > go build -o $BINARY $ENTRYPOINT"
go build -o "$BINARY" "$ENTRYPOINT"

echo
echo "🔹 Running Analyzer project..."
echo "   > $BINARY"
echo

"$BINARY"
