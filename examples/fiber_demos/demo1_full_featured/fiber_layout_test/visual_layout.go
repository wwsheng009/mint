package main

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/examples/component_fixtures"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/compute"
	"github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Layout Size Diagnosis
// =============================================================================

// testLayoutSizeDiagnosis diagnoses why root box has 0x0 size
func testLayoutSizeDiagnosis() {
	vnode := component_fixtures.BuildDemo1App()
	fiber := ui.CreateFiberFromVNode(vnode)

	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  80,
		MinHeight: 0,
		MaxHeight: 24,
	}

	// Print Fiber tree properties
	fmt.Println("\n--- Fiber Tree Properties ---")
	printFiberProperties(fiber, 0)

	// Measure root Fiber
	fmt.Println("\n--- Root Measurement ---")
	engine := compute.NewEngine()
	measurement := fiber.MeasureLayout(engine, constraints)

	fmt.Printf("Root Fiber MeasureLayout result:\n")
	fmt.Printf("  Size: %dx%d\n", measurement.Size.Width, measurement.Size.Height)
	fmt.Printf("  Child Constraints count: %d\n", len(measurement.ChildConstraints))

	// Check if direction is correct
	fmt.Printf("\n--- Direction Check ---")
	fmt.Printf("  Root Direction: %v (should be Column for VStack)\n", fiber.GetDirection())
	fmt.Printf("  Root Tag: %s\n", fiber.Tag)

	// Check children
	children := fiber.GetChildFibers()
	fmt.Printf("  Children count: %d\n", len(children))

	for i, child := range children {
		fmt.Printf("\n  Child[%d]:\n", i)
		fmt.Printf("    Tag: %s\n", child.Tag)
		fmt.Printf("    Direction: %v\n", child.GetDirection())
		fmt.Printf("    Gap: %d\n", child.GetGap())
		fmt.Printf("    Flex: %d\n", child.GetFlex())

		// Measure this child
		childConstraints := constraints
		if i < len(measurement.ChildConstraints) {
			childConstraints = measurement.ChildConstraints[i]
		}
		childMeasurement := child.MeasureLayout(engine, childConstraints)
		fmt.Printf("    MeasureLayout Size: %dx%d\n", childMeasurement.Size.Width, childMeasurement.Size.Height)

		// Also try MeasureChild directly
		childSize := engine.MeasureChild(child, childConstraints)
		fmt.Printf("    MeasureChild Size: %dx%d\n", childSize.Width, childSize.Height)
	}

	// Now run full layout and compare
	fmt.Println("\n--- Full Layout Comparison ---")
	layout, err := engine.BuildComputedBoxFiberOnly(fiber, constraints)
	if err != nil {
		fmt.Printf("[ERR] Layout failed: %v\n", err)
		return
	}

	fmt.Printf("Root ComputedBox: %dx%d at (%d,%d)\n",
		layout.Root.Width, layout.Root.Height, layout.Root.X, layout.Root.Y)

	// Compare measurement vs computed
	if measurement.Size.Width != layout.Root.Width || measurement.Size.Height != layout.Root.Height {
		fmt.Println("[WARN] Measurement size != ComputedBox size!")
		fmt.Printf("  Measurement: %dx%d\n", measurement.Size.Width, measurement.Size.Height)
		fmt.Printf("  ComputedBox: %dx%d\n", layout.Root.Width, layout.Root.Height)
	}
}

func printFiberProperties(fiber *ui.Fiber, depth int) {
	if fiber == nil {
		return
	}

	indent := strings.Repeat("  ", depth)
	fmt.Printf("%sNodeID=%d Tag=%s Type=%v\n", indent, fiber.NodeID, fiber.Tag, fiber.Type)
	fmt.Printf("%s  Direction: %v\n", indent, fiber.GetDirection())
	fmt.Printf("%s  Gap: %d\n", indent, fiber.GetGap())
	fmt.Printf("%s  Flex: %d\n", indent, fiber.GetFlex())
	fmt.Printf("%s  Padding: %v\n", indent, fiber.GetPadding())

	// Check text content
	if fiber.Tag == "text" {
		// Try to get content from Props
		if fiber.Props != nil {
			if content, ok := fiber.Props["content"]; ok {
				fmt.Printf("%s  Content (from Props): %q\n", indent, content)
			}
		}
		// Check MemoizedState
		if fiber.MemoizedState != nil {
			fmt.Printf("%s  MemoizedState: %v (type: %T)\n", indent, fiber.MemoizedState, fiber.MemoizedState)
		} else {
			fmt.Printf("%s  MemoizedState: nil\n", indent)
		}
	}

	// Only print first 3 levels
	if depth >= 2 {
		return
	}

	children := fiber.GetChildFibers()
	for i, child := range children {
		fmt.Printf("%s  Child[%d]:\n", indent, i)
		printFiberProperties(child, depth+1)
	}
}
