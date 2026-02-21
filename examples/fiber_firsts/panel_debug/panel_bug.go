package main

import (
	"fmt"

	"github.com/wwsheng009/mint/ui/components/text"
	"github.com/wwsheng009/mint/ui/components/stack"
	"github.com/wwsheng009/mint/ui/components/border"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/runtime/layout"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("Panel Bug Analysis")
	fmt.Println("========================================")

	// Panel 2 structure
	textNode := text.New("This text is too long and will be wrapped to multiple lines.").SetWrap(true)
	wrapped := rtui.Flex(textNode, 1)
	vbox := stack.New(stack.Column)
	vbox.SetChildrenList([]rtui.VNode{wrapped})

	border := border.New()
	border.SetChild(vbox)
	border.SetBorderLabel(" Demo ")

	// Test 1: Border with width=18
	fmt.Println("\nTest 1: Border with width=18")
	border.SetWidth(18)
	borderInst := border.CreateInstance()

	if measurable, ok := borderInst.(interface{ Measure(layout.Constraints) layout.Size }); ok {
		constraints := layout.Constraints{
			MinWidth:  0,
			MaxWidth:  layout.MaxInt,
			MinHeight: 0,
			MaxHeight: layout.MaxInt,
		}
		fmt.Printf("  Border width prop: %d\n", borderInst.(interface{GetWidth() int}).GetWidth())
		boundsW, _, _, boundsH := borderInst.(interface{ GetBounds() (int, int, int, int) }).GetBounds()
		fmt.Printf("  GetBounds: %dx%d\n", boundsW, boundsH)

		size := measurable.Measure(constraints)
		fmt.Printf("  Measured: %dx%d\n", size.Width, size.Height)
	}

	// Test 2: Border without width
	fmt.Println("\nTest 2: Border without width")
	border2 := border.New()
	border2.SetChild(vbox)
	border2.SetBorderLabel(" Demo ")

	borderInst2 := border2.CreateInstance()
	if measurable, ok := borderInst2.(interface{ Measure(layout.Constraints) layout.Size }); ok {
		constraints := layout.Constraints{
			MinWidth:  0,
			MaxWidth:  layout.MaxInt,
			MinHeight: 0,
			MaxHeight: layout.MaxInt,
		}
		width := borderInst2.(interface{ GetWidth() int }).GetWidth()
		fmt.Printf("  Border width prop: %d\n", width)
		boundsW, _, _, boundsH := borderInst2.(interface{ GetBounds() (int, int, int, int) }).GetBounds()
		fmt.Printf("  GetBounds: %dx%d\n", boundsW, boundsH)

		size := measurable.Measure(constraints)
		fmt.Printf("  Measured: %dx%d\n", size.Width, size.Height)
	}
}
