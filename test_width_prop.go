package main

import (
	"fmt"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/runtime"
)

func main() {
	fmt.Println("=== Testing width prop in LayoutNode ===\n")

	// Create HStack with explicit width prop
	hstack := rtui.HStackBuilder(
		app.ButtonBuilder("Btn1").Build(),
		app.ButtonBuilder("Btn2").Build(),
		app.ButtonBuilder("Btn3").Build(),
	).Gap(1).Build()

	// Set width prop
	if element, ok := hstack.(interface{ SetProp(string, interface{}) }); ok {
		element.SetProp("width", 78)
	}

	// Measure with unbounded constraints first
	fmt.Println("Test 1: Measure without width prop (unbounded)")
	size1 := measureHStack(hstack, 0, runtime.Infinity)
	fmt.Printf("  Size: %dx%d\n", size1.Width, size1.Height)
	fmt.Println()

	// Measure with bounded constraints
	fmt.Println("Test 2: Measure with bounded constraints (80)")
	size2 := measureHStack(hstack, 0, 80)
	fmt.Printf("  Size: %dx%d\n", size2.Width, size2.Height)
	fmt.Println()

	// Check if width prop is read
	fmt.Println("Test 3: Measure with width prop set")
	if props := hstack.Props(); props != nil {
		if w, ok := props["width"].(int); ok {
			fmt.Printf("  width prop: %d ✅\n", w)
		} else {
			fmt.Println("  width prop: NOT SET ❌")
		}
	}
	fmt.Println()

	size3 := measureHStack(hstack, 0, runtime.Infinity)
	fmt.Printf("  Size (with width prop): %dx%d\n", size3.Width, size3.Height)
	fmt.Println()

	// Expected: With width prop=78, HStack should measure as 78 wide
	if size3.Width == 78 {
		fmt.Println("✅ SUCCESS: HStack correctly uses width prop")
	} else {
		fmt.Printf("❌ FAIL: Expected width=78, got %d\n", size3.Width)
	}
}

func measureHStack(vnode rtui.VNode, minWidth, maxWidth int) runtime.Size {
	if measurable, ok := vnode.(interface {
		Measure(runtime.BoxConstraints) runtime.Size
	}); ok {
		return measurable.Measure(runtime.BoxConstraints{
			MinWidth:  minWidth,
			MaxWidth: maxWidth,
			MinHeight: 0,
			MaxHeight: 1000,
		})
	}
	return runtime.Size{Width: 0, Height: 0}
}
