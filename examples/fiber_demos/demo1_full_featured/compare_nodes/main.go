// Demo 1: Full-Featured Demo App
//
// This demo demonstrates the complete TUI engine architecture, covering:
// - Declarative components
// - State system (Hooks)
// - Layout system (Flex, VStack, HStack, Table)
// - Modal (Layer) - Using Layer system
// - Input with Focus management
// - Theme system with semantic colors
// - Button variants (Primary, Secondary, Danger, Success)
// - Scroll containers
// - VirtualList for large data
// - Event handling
// - Animation
//
// This is an integration acceptance test for the UI Runtime.
//
// Based on: framework/docs/ui/demo/demo1.md
//
// Component definitions are in: examples/component_fixtures/

package main

import (
	"fmt"

	"github.com/wwsheng009/mint/examples/component_fixtures"
	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/ui"
)

func main() {
	// ============================================================
	// VNode vs Fiber Tree Comparison Mode
	// ============================================================
	runComparisonMode()

	// ============================================================
	// 演示使用预定义 fixtures
	// ============================================================
	fmt.Println("\n=== Predefined Fixtures Demo ===")
	demoFixtures()
}

// runComparisonMode executes the VNode vs Fiber comparison
func runComparisonMode() {
	fmt.Println("=== VNode vs Fiber Tree Comparison Mode ===")
	fmt.Println()

	// Ensure default theme is loaded
	_ = theme.SetTheme("nord")

	// Build the demo app's VNode tree using component fixtures
	vnode := component_fixtures.BuildDemo1App()

	// Convert to Fiber tree
	fiber := ui.CreateFiberFromVNode(vnode)

	fmt.Println("Fiber tree created successfully")
	fmt.Printf("Root Fiber: NodeID=%d, Tag=%s, Type=%v\n",
		fiber.NodeID, fiber.Tag, fiber.Type)
	fmt.Printf("Children count: %d\n", len(fiber.GetChildFibers()))
	fmt.Println()

	// Run comparison
	result := CompareTrees(vnode, fiber)

	// Print results
	PrintComparisonResult(result)

	// Check information preservation
	fmt.Println("=== Fiber Information Preservation ===")
	CheckFiberPreservation(fiber, 0)

	// Print summary
	PrintSummary(result)
}

// demoFixtures demonstrates using predefined fixtures
func demoFixtures() {
	fixtures := component_fixtures.StandardFixtures()

	fmt.Printf("Available fixtures: %d\n\n", len(fixtures))

	for _, f := range fixtures {
		vnode := f.Build()
		if vnode == nil {
			fmt.Printf("[ERR] %s: failed to build\n", f.Name)
			continue
		}

		count := component_fixtures.CountNodes(vnode)
		fmt.Printf("[OK] %s: %d nodes - %s\n", f.Name, count, f.Description)
	}

	// Demo custom configuration
	fmt.Println("\n=== Custom Configuration Demo ===")

	customVNode := component_fixtures.BuildDemo1App(
		component_fixtures.WithCount(42),
		component_fixtures.WithInput("hello world"),
		component_fixtures.WithItems([]string{"Custom A", "Custom B", "Custom C"}),
	)

	fmt.Printf("Custom app built: %d nodes\n", component_fixtures.CountNodes(customVNode))

	// Demo individual components
	fmt.Println("\n=== Individual Components Demo ===")

	header := component_fixtures.BuildDemo1Header(100)
	fmt.Printf("Header (count=100): %d nodes\n", component_fixtures.CountNodes(header))

	body := component_fixtures.BuildDemo1MainBody(5, "test", []string{"A", "B", "C"})
	fmt.Printf("MainBody: %d nodes\n", component_fixtures.CountNodes(body))

	modal := component_fixtures.BuildDemo1ConfirmModal(func() { fmt.Println("Modal closed") })
	fmt.Printf("Modal: %d nodes\n", component_fixtures.CountNodes(modal))
}
