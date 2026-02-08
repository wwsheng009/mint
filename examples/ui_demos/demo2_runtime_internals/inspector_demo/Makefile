# UI Inspector Demo Makefile

.PHONY: all build run clean demo simple help

# Default target
all: build

# Build the full inspector demo
build:
	@echo "Building UI Inspector Demo..."
	go build -o demo2_inspector main.go
	@echo "✓ Build complete: ./demo2_inspector"

# Build simple demo
simple:
	@echo "Building Simple Inspector Demo..."
	go build -o simple_demo simple_demo.go
	@echo "✓ Build complete: ./simple_demo"

# Run the full inspector demo
run: build
	@echo "Starting UI Inspector Demo..."
	@echo ""
	./demo2_inspector

# Run simple demo
run-simple: simple
	@echo "Starting Simple Inspector Demo..."
	@echo ""
	./simple_demo

# Run with environment variable enabled
run-enabled:
	@echo "Starting UI Inspector Demo (Enabled)..."
	@echo ""
	TUI_INSPECTOR=true ./demo2_inspector

# Generate reports
report:
	@echo "Generating Inspector Reports..."
	@echo ""
	@echo "This feature is integrated into the demo."
	@echo "Run the demo and press Ctrl+C to see the final report."

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f demo2_inspector simple_demo
	@echo "✓ Clean complete"

# Help
help:
	@echo "UI Inspector Demo - Makefile"
	@echo ""
	@echo "Targets:"
	@echo "  make all         - Build the demo (default)"
	@echo "  make build       - Build the demo"
	@echo "  make simple      - Build simple demo"
	@echo "  make run         - Build and run the demo"
	@echo "  make run-simple  - Build and run simple demo"
	@echo "  make run-enabled - Run with TUI_INSPECTOR=true"
	@echo "  make report      - Show report generation info"
	@echo "  make clean       - Remove build artifacts"
	@echo "  make help        - Show this help"
	@echo ""
	@echo "Examples:"
	@echo "  make              # Build the demo"
	@echo "  make run          # Build and run"
	@echo "  make run-enabled  # Run with inspector auto-enabled"
	@echo ""
	@echo "Features:"
	@echo "  • Real-time performance monitoring"
	@echo "  • Layout diagnostics and problem detection"
	@echo "  • Tree view visualization"
	@echo "  • Selected element information"
	@echo "  • Report export (Text, Markdown, JSON)"
