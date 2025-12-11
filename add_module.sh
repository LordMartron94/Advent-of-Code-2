#!/usr/bin/env bash
set -euo pipefail

# Always run from the script's directory (assumed to be repo root)
cd "$(dirname "$0")"

# --- Configuration ---
DEFAULT_PARENT_DIR="libs"
CONFIG_FILE="submodules.conf" # Default config file name

# --- ANSI Color Definitions ---
# Check if stdout is a terminal, otherwise output no color.
if tty -s; then
    # General
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    RED='\033[0;31m'
    CYAN='\033[0;36m'
    NC='\033[0m' # No Color

    # Styles
    BOLD='\033[1m'
else
    # Fallback to no colors if not running in a TTY
    GREEN=''
    YELLOW=''
    RED=''
    CYAN=''
    NC=''
    BOLD=''
fi

# Helper function to print colored output
log_success() {
    echo -e "${GREEN}${BOLD}✔ SUCCESS:${NC} ${1}"
}
log_warning() {
    echo -e "${YELLOW}${BOLD}⚠️ WARNING:${NC} ${1}" >&2
}
log_error() {
    echo -e "${RED}${BOLD}❌ ERROR:${NC} ${1}" >&2
}
log_critical() {
    echo -e "${RED}${BOLD}💥 CRITICAL ERROR:${NC} ${1}" >&2
}
log_info() {
    echo -e "${CYAN}${1}${NC}"
}
# --- End ANSI Color Definitions ---


# Function to ensure .gitmodules file exists
ensure_gitmodules() {
    if [ ! -f .gitmodules ]; then
        log_info "--> No .gitmodules file found, creating one..."
        touch .gitmodules
    fi
}

# Core function to add a single submodule
add_submodule() {
    local NAME="$1"
    local URL="$2"
    local PARENT_DIR="$3"
    local LOCATION="$PARENT_DIR/$NAME"

    echo
    echo -e "${CYAN}--- Processing Submodule: ${BOLD}$NAME${NC}${CYAN} ---${NC}"
    echo "  URL:      $URL"
    echo "  Location: $LOCATION"

    # --- Idempotency Check ---
    if [ -d "$LOCATION/.git" ] || git config --get submodule."$NAME".url >/dev/null 2>&1; then
        log_info "  --> Submodule $NAME already exists (or location '$LOCATION' is a git repo)."
        log_info "      Attempting to synchronize and update the existing submodule."
        
        if ! git submodule sync "$NAME" 2>/dev/null; then
           log_info "  Note: Submodule $NAME needed initialization/sync before update. (Benign)."
        fi
        
        if git submodule update --init --recursive "$LOCATION"; then
            log_success "Update successful. Skipping 'git submodule add'."
            return 0
        else
            log_error "Failed to update existing submodule '$NAME'. Manual intervention required."
            return 1
        fi
    fi
    # --- End Idempotency Check ---

    # Ensure parent directory exists
    mkdir -p "$PARENT_DIR"

    # Add the submodule (Only runs if the checks above passed)
    if git submodule add --name "$NAME" "$URL" "$LOCATION"; then
        log_success "Submodule added successfully."
        # Update/Init is critical here
        if git submodule update --init --recursive "$LOCATION"; then
            log_info "Initialization and update complete."
        else
            # This is less common but indicates a failure in checkout after a successful add
            log_warning "Submodule added, but failed to initialize/checkout contents."
        fi
    else
        # This branch handles failures like bad URL, network issues, or "fatal: 'libs/memarch' does not have a commit checked out"
        # which you observed, indicating a deep checkout error.
        log_critical "Failed to add submodule '$NAME'. Check URL and permissions."
        return 1
    fi
}

# Function for interactive mode
interactive_mode() {
    echo -e "${CYAN}=== Git Submodule Adder (Interactive Mode) ===${NC}"

    # Ask for details
    read -rp "Enter submodule name (e.g. vision): " NAME
    read -rp "Enter repository URL (e.g. ssh://git@localhost:22222/home/git/repo.git): " URL
    read -rp "Enter parent directory where submodule should be placed (default: $DEFAULT_PARENT_DIR): " PARENT_DIR

    # Default parent directory if empty
    if [ -z "$PARENT_DIR" ]; then
        PARENT_DIR="$DEFAULT_PARENT_DIR"
    fi

    echo
    ensure_gitmodules
    add_submodule "$NAME" "$URL" "$PARENT_DIR"
    echo
    log_info "Interactive setup complete!"
}

# Function for batch/config mode
process_config_file() {
    local CONFIG_PATH="$1"
    echo -e "${CYAN}=== Git Submodule Adder (Batch Mode) ===${NC}"
    log_info "Reading configuration from: ${BOLD}$CONFIG_PATH${NC}"
    
    if [ ! -f "$CONFIG_PATH" ]; then
        log_error "Configuration file not found at '$CONFIG_PATH'."
        exit 1
    fi

    ensure_gitmodules
    
    # Read the file line by line
    local LINE_NUM=0
    while IFS=$' \t\n' read -r LINE || [ -n "$LINE" ]; do
        LINE_NUM=$((LINE_NUM + 1))
        if [[ "$LINE" =~ ^[[:space:]]*# ]] || [[ -z "$LINE" ]]; then
            continue
        fi

        local FIELDS
        FIELDS=($LINE)
        
        local NAME="${FIELDS[0]}"
        local URL="${FIELDS[1]}"
        local PARENT_DIR="${FIELDS[2]:-$DEFAULT_PARENT_DIR}" # Use default if not provided

        # Minimal validation
        if [ -z "$NAME" ] || [ -z "$URL" ]; then
            log_warning "Skipping line $LINE_NUM. Invalid format. Expected: NAME URL [PARENT_DIR]"
            continue
        fi

        # We call the function and capture the exit status.
        # || true is used in the main loop to ensure processing continues even if one submodule fails.
        add_submodule "$NAME" "$URL" "$PARENT_DIR" || true 
    done < "$CONFIG_PATH"
    
    echo
    log_info "Batch setup complete! All configured submodules processed."
}

# --- Main Execution Flow ---

# Check if a configuration file path is provided as an argument
if [ "$#" -gt 0 ]; then
    # Argument provided, use it as the config file path
    process_config_file "$1"
else
    # No argument provided, run in interactive mode
    interactive_mode
fi