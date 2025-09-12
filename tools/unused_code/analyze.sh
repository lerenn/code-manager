#!/bin/bash

# Convenience script to run the unused code analyzer
# Usage: ./analyze.sh [directory] [options]

set -e

# Get the directory of this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Check if Go is available
if ! command -v go &> /dev/null; then
    echo "❌ Error: Go is not installed or not in PATH"
    exit 1
fi

# Run the analyzer with all passed arguments
go run "$SCRIPT_DIR/main.go" "$@"
