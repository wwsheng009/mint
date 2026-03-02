package main

import (
	"fmt"
	"os"
	"reflect"

	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/compute"
	"github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/ui/components/panel"
	"github.com/wwsheng009/mint/ui/components/text"
	"github.com/wwsheng009/mint/ui/layout/visualizer"
)

// Demo: SVG visualization with actual layout engine data
func main() {
	fmt.Println("=== SVG Layout Visualization (with Layout Engine) ===\n")

	// Create layout engine
	engine := compute.NewEngine()

	// Example 1: Simple panel with actual layout
	fmt.Println("Example 1: Simple Panel")
	fmt.Println("---")
	simplePanel := panel.NewBuilder().
		Title("Settings").
		OuterSize(40, 15).
		Content(text.New("Settings go here")).
		Build()

	constraints := runtime.BoxConstraints{
		MinWidth:   0,
		MaxWidth:   80,
		MinHeight:  0,
		MaxHeight:  24,
	}

	// Run layout engine
	computedLayout, err := engine.Layout(simplePanel, nil, constraints)
	if err != nil {
		fmt.Printf("Error running layout: %v\n", err)
		return
	}

	// Create visualizer from computed layout
	vis := visualizer.VisualizeFromLayoutEngine(computedLayout)

	// Generate SVG - tree view
	svgOutput := vis.PrintSVG()

	err = os.WriteFile("simple_layout_real.svg", []byte(svgOutput), 0644)
	if err != nil {
		fmt.Printf("Error writing SVG: %v\n", err)
		return
	}
	fmt.Println("✓ Generated simple_layout_real.svg (tree view)")

	// Generate SVG - nested box view (actual layout positions)
	svgNested := vis.PrintSVGNestedBox()
	err = os.WriteFile("simple_layout_nested.svg", []byte(svgNested), 0644)
	if err != nil {
		fmt.Printf("Error writing nested SVG: %v\n", err)
		return
	}
	fmt.Println("✓ Generated simple_layout_nested.svg (nested box view)")
	fmt.Println("  The nested box SVG shows actual spatial layout!\n")

	// Example 2: Complex nested layout
	fmt.Println("Example 2: Complex Nested Layout")
	fmt.Println("---")

	complexLayout := panel.NewBuilder().
		Title("Dashboard").
		OuterSize(70, 30).
		Content(
			ui.NewVStack().
				SetChildren([]ui.VNode{
					panel.NewBuilder().
						Title("Statistics").
						OuterSize(60, 8).
						Content(text.New("Users: 100 | Sales: $5K | Orders: 50")).
						Build(),
					panel.NewBuilder().
						Title("Recent Activity").
						OuterSize(60, 10).
						Content(text.New("Latest updates and user activities go here")).
						Build(),
					panel.NewBuilder().
						Title("Quick Actions").
						OuterSize(60, 6).
						Content(text.New("Create | Edit | Delete")).
						Build(),
				}),
		).
		Build()

	// Run layout engine
	computedLayout2, err := engine.Layout(complexLayout, nil, constraints)
	if err != nil {
		fmt.Printf("Error running layout: %v\n", err)
		return
	}

	vis2 := visualizer.VisualizeFromLayoutEngine(computedLayout2)
	svgOutput2 := vis2.PrintSVG()

	err = os.WriteFile("complex_layout_real.svg", []byte(svgOutput2), 0644)
	if err != nil {
		fmt.Printf("Error writing SVG: %v\n", err)
		return
	}
	fmt.Println("✓ Generated complex_layout_real.svg (tree view)")

	svgNested2 := vis2.PrintSVGNestedBox()
	err = os.WriteFile("complex_layout_nested.svg", []byte(svgNested2), 0644)
	if err != nil {
		fmt.Printf("Error writing nested SVG: %v\n", err)
		return
	}
	fmt.Println("✓ Generated complex_layout_nested.svg (nested box view)\n")

	// Example 3: HStack with horizontal layout
	fmt.Println("Example 3: Horizontal Stack (HStack)")
	fmt.Println("---")

	hstackLayout := ui.NewHStack().
		SetChildren([]ui.VNode{
			text.New("First"),
			text.New("Second"),
			text.New("Third"),
		})

	// Run layout engine
	hstackConstraints := runtime.BoxConstraints{
		MinWidth:   0,
		MaxWidth:   60,
		MinHeight:  0,
		MaxHeight:  10,
	}
	computedLayout3, err := engine.Layout(hstackLayout, nil, hstackConstraints)
	if err != nil {
		fmt.Printf("Error running layout: %v\n", err)
		return
	}

	vis3 := visualizer.VisualizeFromLayoutEngine(computedLayout3)
	svgOutput3 := vis3.PrintSVG()

	err = os.WriteFile("hstack_layout_real.svg", []byte(svgOutput3), 0644)
	if err != nil {
		fmt.Printf("Error writing SVG: %v\n", err)
		return
	}
	fmt.Println("✓ Generated hstack_layout_real.svg (tree view)")

	svgNested3 := vis3.PrintSVGNestedBox()
	err = os.WriteFile("hstack_layout_nested.svg", []byte(svgNested3), 0644)
	if err != nil {
		fmt.Printf("Error writing nested SVG: %v\n", err)
		return
	}
	fmt.Println("✓ Generated hstack_layout_nested.svg (nested box view)\n")

	// Print summary
	fmt.Println("=== Summary ===")
	fmt.Printf("Example 1: Simple Panel\n")
	printLayoutSummary(computedLayout)
	fmt.Printf("\nExample 2: Complex Nested Layout\n")
	printLayoutSummary(computedLayout2)
	fmt.Printf("\nExample 3: HStack\n")
	printLayoutSummary(computedLayout3)

	fmt.Println("\n=== Generated SVG Files ===")
	fmt.Println("Tree view (structural):")
	fmt.Println("  simple_layout_real.svg     - Shows tree structure with detailed info")
	fmt.Println("  complex_layout_real.svg    - Shows complex nested structure")
	fmt.Println("  hstack_layout_real.svg     - Shows horizontal stack structure")
	fmt.Println()
	fmt.Println("Nested box view (spatial layout):")
	fmt.Println("  simple_layout_nested.svg   - Shows actual spatial positions!")
	fmt.Println("  complex_layout_nested.svg  - Shows nested boxes in real positions")
	fmt.Println("  hstack_layout_nested.svg   - Shows horizontal arrangement")
	fmt.Println()
	fmt.Println("Note: Compare the nested box SVGs to see the ACTUAL layout!")
	fmt.Println("      Boxes are positioned and sized according to layout engine.")
}

func printLayoutSummary(layout interface{}) {
	// Use reflection to access layout data
	rv := reflect.ValueOf(layout).Elem()
	rootBox := rv.FieldByName("Root").Elem()

	// Get Box position and size
	boxField := rootBox.FieldByName("Box")
	x := int(boxField.FieldByName("X").Int())
	y := int(boxField.FieldByName("Y").Int())
	w := int(boxField.FieldByName("Width").Int())
	h := int(boxField.FieldByName("Height").Int())

	// Count children
	childrenField := rootBox.FieldByName("Children")
	childCount := 0
	if childrenField.IsValid() && childrenField.Kind() == reflect.Slice {
		childCount = childrenField.Len()
	}

	fmt.Printf("  Root Position: (%d, %d)\n", x, y)
	fmt.Printf("  Root Size: %d × %d\n", w, h)
	fmt.Printf("  Direct Children: %d\n", childCount)
}
