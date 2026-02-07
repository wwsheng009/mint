package main

import (
	"fmt"
	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/runtime"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func main() {
	// Create buttons like demo2
	allButtons := []rtui.VNode{
		app.ButtonBuilder("[1] Event").Build(),
		app.ButtonBuilder("[2]setState").Build(),
		app.ButtonBuilder("[3]Scheduler").Build(),
		app.ButtonBuilder("[4] Render").Build(),
	}

	// Wrap with FillWidth (like demo2)
	wrappedButtons := app.WrapBuilder(allButtons...).
		Gap(1).
		ScreenWidth(78).
		Align(rtui.AlignCenter).
		FillWidth().
		Build()

	fmt.Printf("Wrapped type: %T\n", wrappedButtons)
	
	// Check VStack children (rows)
	rows := wrappedButtons.Children()
	fmt.Printf("Number of rows: %d\n", len(rows))
	
	if len(rows) > 0 {
		row1 := rows[0]
		fmt.Printf("Row1 type: %T, Tag: %s\n", row1, getTag(row1))
		
		// Check HStack children (buttons)
		buttons := row1.Children()
		fmt.Printf("Number of buttons in row1: %d\n", len(buttons))
		
		for i, btn := range buttons {
			fmt.Printf("\nButton %d:\n", i+1)
			fmt.Printf("  Type: %T\n", btn)
			fmt.Printf("  Tag: %s\n", getTag(btn))
			
			props := btn.Props()
			if props != nil {
				if flex, ok := props["flex"].(int); ok {
					fmt.Printf("  flex prop: %d ✅\n", flex)
				} else {
					fmt.Printf("  flex prop: NOT SET ❌\n")
				}
			}
			
			info := rtui.GetLayoutInfo(btn)
			fmt.Printf("  GetLayoutInfo.Flex: %d\n", info.Flex)
		}
	}
	
	// Measure the wrapped node
	fmt.Println("\nMeasuring wrapped node:")
	size := measure(wrappedButtons, 0, 78)
	fmt.Printf("Size: %dx%d\n", size.Width, size.Height)
}

func getTag(vnode rtui.VNode) string {
	if tagger, ok := vnode.(interface{ Tag() string }); ok {
		return tagger.Tag()
	}
	return "no-tag"
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
