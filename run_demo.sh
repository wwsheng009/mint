#!/bin/bash
export TUI_LAYER_DEBUG=true
timeout 5 go run examples/ui_demos/demo1_full_featured/main.go 2>&1 | grep -E "\[.*\]|hasLayers|layer=" || true
