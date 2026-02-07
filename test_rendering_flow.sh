#!/bin/bash

echo "=== Testing Rendering Flow ==="
echo ""

# Build the demo
echo "Building elegant_api_demo..."
go build -o elegant_api_test.exe ./examples/elegant_api_demo

echo ""
echo "Running with FULL debug output..."
echo "=================================="

# Run with all debug flags and capture to file
TUI_LAYOUT_DEBUG=true TUI_PAINT_DEBUG=true timeout 1 ./elegant_api_test.exe 2>&1 | tee /tmp/full_debug.txt

echo ""
echo "=================================="
echo "Filtering for Button elements:"
echo "=================================="

cat /tmp/full_debug.txt | grep -E "Button|size 26x1" | head -30

echo ""
echo "=================================="
echo "Looking for DEBUG-PAINT output:"
echo "=================================="

cat /tmp/full_debug.txt | grep "DEBUG-PAINT" | head -10

echo ""
echo "Done. Full output saved to /tmp/full_debug.txt"
