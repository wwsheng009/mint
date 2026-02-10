#!/bin/bash
# Debug script for Inspector TreeView click issue

echo "=== Inspector Mouse Click Debug ==="
echo ""
echo "Running demo with verbose inspector output..."
echo ""

# Enable verbose inspector output
export TUI_INSPECTOR_VERBOSE=true

# Enable debug UI output
export TUI_DEBUG_UI=true

# Run the demo
cd /e/projects/yao/wwsheng009/mint
echo "Starting demo..."
echo ""
echo "Instructions:"
echo "1. Press F12 to toggle Inspector"
echo "2. Click on Elements tab"
echo "3. Try clicking on TreeView items"
echo "4. Check console output for debug messages"
echo ""
echo "Look for:"
echo "  [APP] Routing mouse (...) to Inspector"
echo "  [Inspector] TreeView click: localY=..."
echo "  [Inspector] TreeView clicked at line ..."
echo ""

go run ./examples/ui_demos/demo2_runtime_internals/inspector_overlay/main.go
