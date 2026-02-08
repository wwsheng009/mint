#!/bin/bash

# Demo2 with UI Inspector - Runner Script
#
# This script runs the demo2 with integrated UI Inspector

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "=========================================="
echo "Demo2 with UI Inspector"
echo "=========================================="
echo ""
echo "Building..."

# Build the inspector demo
cd inspector_demo
go build -o demo2_inspector main.go

echo "Build complete!"
echo ""
echo "Starting demo with UI Inspector..."
echo ""
echo "Features:"
echo "  - Real-time performance monitoring (FPS, memory)"
echo "  - Layout diagnostics (constraint conflicts, overflow)"
echo "  - Tree view visualization"
echo "  - Selected element info"
echo ""
echo "Controls:"
echo "  - Click [I] Toggle Inspector button to enable/disable"
echo "  - Tab: Navigate between elements"
echo "  - Enter: View element details"
echo "  - Esc: Clear selection"
echo ""
echo "=========================================="
echo ""

# Run the demo
./demo2_inspector

echo ""
echo "Demo finished."
