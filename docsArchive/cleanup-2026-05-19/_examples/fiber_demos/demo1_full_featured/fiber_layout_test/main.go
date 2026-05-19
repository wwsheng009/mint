// Fiber Layout Test Demo
//
// This demo tests BuildComputedBoxFiberOnly from runtime/compute/fiber_only_layout.go
// It demonstrates Fiber-first layout computation without VNode access.
//
// Usage:
//   go run .                    # Run all tests
//   go run . -validate          # Run automated validation only (exit code 0/1)
//   go run . -batch             # Run batch validation on all fixtures
//   go run . -visualize         # Run layout visualization with ASCII diagram
//   go run . -details           # Run visualization with detailed tree
//
// Test Flow:
// 1. Build VNode tree using component_fixtures
// 2. Convert to Fiber tree
// 3. Run BuildComputedBoxFiberOnly layout
// 4. Verify layout results
//
// See: docs/plan/fiber/fiber_first.md

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/wwsheng009/mint/examples/component_fixtures"
	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/compute"
	"github.com/wwsheng009/mint/ui"
)

func main() {
	// Command line flags
	validateOnly := flag.Bool("validate", false, "Run automated validation only (exit 0 on pass, 1 on fail)")
	batchMode := flag.Bool("batch", false, "Run batch validation on all fixtures")
	visualizeMode := flag.Bool("visualize", false, "Run layout visualization with ASCII diagram")
	detailsMode := flag.Bool("details", false, "Run visualization with detailed tree output")
	flag.Parse()

	// Ensure theme is loaded
	_ = theme.SetTheme("nord")

	// Handle command modes
	if *validateOnly {
		// Run validation and exit with appropriate code
		success := RunValidation()
		if success {
			os.Exit(0)
		}
		os.Exit(1)
	}

	if *batchMode {
		RunBatchValidation()
		return
	}

	if *visualizeMode {
		RunVisualization()
		return
	}

	if *detailsMode {
		RunDetailedVisualization()
		return
	}

	// Default: Run all tests
	fmt.Println("=== Fiber Layout Test Demo ===")
	fmt.Println("Testing BuildComputedBoxFiberOnly from fiber_only_layout.go")
	fmt.Println()

	// Test 1: Basic layout with default constraints
	fmt.Println("=== Test 1: Default Constraints (80x24) ===")
	testBasicLayout()

	// Test 2: Different constraint sizes
	fmt.Println("\n=== Test 2: Various Constraints ===")
	testVariousConstraints()

	// Test 3: Individual components
	fmt.Println("\n=== Test 3: Individual Components ===")
	testIndividualComponents()

	// Test 4: Compare with fixture node counts
	fmt.Println("\n=== Test 4: Node Count Verification ===")
	testNodeCounts()

	// Test 5: Automated constraint validation
	fmt.Println("\n=== Test 5: Automated Layout Constraint Validation ===")
	RunValidation()

	// Test 6: Diagnose layout size issue
	fmt.Println("\n=== Test 6: Layout Size Diagnosis ===")
	// testLayoutSizeDiagnosis()
}

func testBasicLayout() {
	// Build VNode tree
	vnode := component_fixtures.BuildDemo1App()

	// Convert to Fiber tree
	fiber := ui.CreateFiberFromVNode(vnode)

	// Create layout engine
	engine := compute.NewEngine()

	// Set constraints (80x24 terminal)
	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: 24,
	}

	// Run Fiber-only layout
	layout, err := engine.LayoutFiber(fiber, constraints)
	if err != nil {
		fmt.Printf("[ERR] Layout failed: %v\n", err)
		return
	}

	// Print results
	fmt.Printf("[OK] Layout completed successfully\n")
	fmt.Printf("     Root Box: X=%d, Y=%d, W=%d, H=%d\n",
		layout.Root.X, layout.Root.Y, layout.Root.Width, layout.Root.Height)
	fmt.Printf("     Children: %d\n", len(layout.Root.Children))
	fmt.Printf("     HitMap Size: %d\n", layout.HitMap.Size())

	// Print tree structure
	fmt.Println("\n--- Layout Tree Structure ---")
	printBoxTree(layout.Root, 0)
}

func testVariousConstraints() {
	constraints := []struct {
		name string
		w, h int
	}{
		{"Small (40x12)", 40, 12},
		{"Medium (80x24)", 80, 24},
		{"Large (120x40)", 120, 40},
		{"Wide (200x10)", 200, 10},
		{"Tall (60x50)", 60, 50},
	}

	for _, c := range constraints {
		vnode := component_fixtures.BuildDemo1App()
		fiber := ui.CreateFiberFromVNode(vnode)
		engine := compute.NewEngine()

		constraint := runtime.BoxConstraints{
			MinWidth:  0,
			MaxWidth:  c.w,
			MinHeight: 0,
			MaxHeight: c.h,
		}

		layout, err := engine.LayoutFiber(fiber, constraint)
		if err != nil {
			fmt.Printf("[ERR] %s: %v\n", c.name, err)
			continue
		}

		fmt.Printf("[OK] %s: Root=%dx%d at (%d,%d), Children=%d, HitMap=%d\n",
			c.name, layout.Root.Width, layout.Root.Height,
			layout.Root.X, layout.Root.Y,
			len(layout.Root.Children), layout.HitMap.Size())
	}
}

func testIndividualComponents() {
	components := []struct {
		name  string
		build func() ui.VNode
	}{
		{"Header", func() ui.VNode { return component_fixtures.BuildDemo1Header(42) }},
		{"MainBody", func() ui.VNode { return component_fixtures.BuildDemo1MainBody(5, "test", []string{"A", "B", "C"}) }},
		{"Modal", func() ui.VNode { return component_fixtures.BuildDemo1ConfirmModal(func() {}) }},
		{"SimpleVStack", func() ui.VNode { return component_fixtures.GetFixture("simple_vstack").Build() }},
		{"SimpleHStack", func() ui.VNode { return component_fixtures.GetFixture("simple_hstack").Build() }},
		{"NestedLayout", func() ui.VNode { return component_fixtures.GetFixture("nested_layout").Build() }},
		{"FlexLayout", func() ui.VNode { return component_fixtures.GetFixture("flex_layout").Build() }},
	}

	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: 24,
	}

	for _, comp := range components {
		vnode := comp.build()
		fiber := ui.CreateFiberFromVNode(vnode)
		engine := compute.NewEngine()

		layout, err := engine.LayoutFiber(fiber, constraints)
		if err != nil {
			fmt.Printf("[ERR] %s: %v\n", comp.name, err)
			continue
		}

		fmt.Printf("[OK] %s: %dx%d, Children=%d, HitMap=%d\n",
			comp.name, layout.Root.Width, layout.Root.Height,
			len(layout.Root.Children), layout.HitMap.Size())
	}
}

func testNodeCounts() {
	fixtures := component_fixtures.StandardFixtures()

	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: 24,
	}

	fmt.Println("Comparing VNode count vs ComputedBox count:")
	fmt.Printf("%-20s %10s %12s %10s\n", "Fixture", "VNode", "ComputedBox", "Match")
	fmt.Println(strings.Repeat("-", 55))

	for _, f := range fixtures {
		vnode := f.Build()
		vnodeCount := component_fixtures.CountNodes(vnode)

		fiber := ui.CreateFiberFromVNode(vnode)
		engine := compute.NewEngine()

		layout, err := engine.LayoutFiber(fiber, constraints)
		if err != nil {
			fmt.Printf("%-20s %10d %12s %10s\n", f.Name, vnodeCount, "ERROR", "X")
			continue
		}

		boxCount := countBoxes(layout.Root)
		match := "OK"
		if vnodeCount != boxCount {
			match = "X"
		}

		fmt.Printf("%-20s %10d %12d %10s\n", f.Name, vnodeCount, boxCount, match)
	}
}

// Helper functions

func printBoxTree(box *compute.ComputedBox, depth int) {
	if box == nil {
		return
	}

	indent := strings.Repeat("  ", depth)
	fmt.Printf("%sBox: X=%d, Y=%d, W=%d, H=%d, NodeID=%d\n",
		indent, box.X, box.Y, box.Width, box.Height, box.NodeID)

	for _, child := range box.Children {
		printBoxTree(child, depth+1)
	}
}

func countBoxes(box *compute.ComputedBox) int {
	if box == nil {
		return 0
	}
	count := 1
	for _, child := range box.Children {
		count += countBoxes(child)
	}
	return count
}
