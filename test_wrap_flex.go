package main

import (
	"fmt"
	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/components/layout"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/runtime"
)

func main() {
	fmt.Println("=== Testing Wrap FillWidth ===\n")

	// Create buttons
	btn1 := app.ButtonBuilder("[1] Event").Build()
	btn2 := app.ButtonBuilder("[2] setState").Build()
	btn3 := app.ButtonBuilder("[3] Scheduler").Build()
	btn4 := app.ButtonBuilder("[4] Render").Build()

	// Wrap with FillWidth
	wrapped := layout.NewWrapBuilder(btn1, btn2, btn3, btn4).
		Gap(1).
		ScreenWidth(78).
		FillWidth().
		Build()

	// Check if wrapped is VStack
	fmt.Printf("Wrapped type: %T\n", wrapped)
	if tagger, ok := wrapped.(interface{ Tag() string }); ok {
		fmt.Printf("Wrapped tag: %s\n", tagger.Tag())
	}
	fmt.Println()

	// Check children (should be HStacks)
	children := wrapped.Children()
	fmt.Printf("Number of children (rows): %d\n", len(children))
	fmt.Println()

	// Check first row
	if len(children) > 0 {
		row1 := children[0]
		fmt.Printf("Row 1 type: %T\n", row1)
		if tagger, ok := row1.(interface{ Tag() string }); ok {
			fmt.Printf("Row 1 tag: %s\n", tagger.Tag())
		}

		// Check row children (should be buttons)
		row1Children := row1.Children()
		fmt.Printf("Number of buttons in row 1: %d\n", len(row1Children))
		fmt.Println()

		// Check each button's flex prop
		for i, btn := range row1Children {
			props := btn.Props()
			if props == nil {
				fmt.Printf("Button %d: props = nil ❌\n", i+1)
			} else {
				if flex, ok := props["flex"].(int); ok {
					fmt.Printf("Button %d: flex = %d ✅\n", i+1, flex)
				} else {
					fmt.Printf("Button %d: flex prop NOT SET ❌\n", i+1)
				}
			}

			// Check GetLayoutInfo
			info := rtui.GetLayoutInfo(btn)
			fmt.Printf("  GetLayoutInfo.Flex = %d\n", info.Flex)
		}
	}

	fmt.Println()

	// Measure the wrapped node
	fmt.Println("Measuring wrapped node:")
	size := measure(wrapped, 0, 78)
	fmt.Printf("Size: %dx%d\n", size.Width, size.Height)
}

func measure(vnode rtui.VNode, minWidth, maxWidth int) runtime.Size {
	if measurable, ok := vnode.(interface {
		Measure(runtime.BoxConstraints) runtime.Size
	}); ok {
		return measurable.Measure(runtime.BoxConstraints{
			MinWidth:  minWidth,
			MaxWidth:  maxWidth,
			MinHeight: 0,
			MaxHeight: 1000,
		})
	}
	return runtime.Size{Width: 0, Height: 0}
}
