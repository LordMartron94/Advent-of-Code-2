#!/usr/bin/env bash
set -euo pipefail

# Always run from repo root
cd "$(dirname "$0")"

# Configuration
BINARY_DIR="./build"
BINARY_NAME="aoc_runner"
SRC_PATH="./src" 

# ────────────────────────────────────────────────
# Argument Parsing
# ────────────────────────────────────────────────
TESTING_MODE=false
CLEAN_MODE=false
TODAY_MODE=false
PASSTHROUGH_ARGS=()

while [[ $# -gt 0 ]]; do
    case "$1" in
        --today)
            TODAY_MODE=true
            shift
            ;;
        --test|-t)
            TESTING_MODE=true
            shift
            ;;
        --clean|-c)
            CLEAN_MODE=true
            shift
            ;;
        *)
            PASSTHROUGH_ARGS+=("$1")
            shift
            ;;
    esac
done

# ────────────────────────────────────────────────
# Execution Logic
# ────────────────────────────────────────────────

# 1. Clean
if [ "$CLEAN_MODE" = true ]; then
    echo "🧹 Cleaning build artifacts..."
    rm -rf "$BINARY_DIR"
    echo "Done."
    exit 0
fi

# 2. Test
if [ "$TESTING_MODE" = true ]; then
    echo "🧪 Running tests..."
    go test ./... -v
    exit 0
fi

# 3. Handle --today logic
if [ "$TODAY_MODE" = true ]; then
    # Get current year and day. 
    # ${d#0} strips leading zeros so "05" becomes "5", preventing octal confusion.
    CURRENT_YEAR=$(date +%Y)
    d=$(date +%d)
    CURRENT_DAY=${d#0}
    
    # Overwrite args to force today's date
    PASSTHROUGH_ARGS=("$CURRENT_YEAR" "$CURRENT_DAY")
    echo "📅 Today mode active: target is $CURRENT_YEAR Day $CURRENT_DAY"
fi

# 4. Build (Optimized)
mkdir -p "$BINARY_DIR"

# Only build if binary is missing OR source is newer than binary
if [ ! -f "$BINARY_DIR/$BINARY_NAME" ] || [ "$SRC_PATH/runner.go" -nt "$BINARY_DIR/$BINARY_NAME" ]; then
    echo "🔨 Building Runner..."
    go build -o "$BINARY_DIR/$BINARY_NAME" "$SRC_PATH"
fi

# 5. Execute
echo "🚀 Executing Runner..."
echo "-----------------------------------"
"$BINARY_DIR/$BINARY_NAME" "${PASSTHROUGH_ARGS[@]}"