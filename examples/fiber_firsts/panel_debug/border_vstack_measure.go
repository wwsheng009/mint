package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/ui/components/border"
	"github.com/wwsheng009/mint/ui/components/stack"
	"github.com/wwsheng009/mint/ui/components/text"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("Border(VStack(Text)) Measurement Test")
	fmt.Println("========================================")

	// Panel creates: Border(VStack(flex(Text)))
	textNode := text.New("This text is too long and will be wrapped to multiple lines.").SetWrap(true)
	wrappedNode := rtui.Flex(textNode, 1)
	vstackNode := stack.New(stack.Column)
	vstackNode.SetChildrenList([]rtui.VNode{wrappedNode})

	borderNode := border.New()
	borderNode.SetChild(vstackNode)
	borderNode.SetBorderLabel(" Demo ")
	borderNode.SetWidth(18)  // Panel 20 - border 2 = 18

	testCases := []struct {
		name      string
		maxWidth  int
		maxHeight int
	}{
		{"No height constraint", 50, layout.MaxInt},
		{"Height constraint=10", 50, 10},
		{"Height constraint=5", 50, 5},
	}

	for _, tc := range testCases {
		fmt.Printf("\n%s\n", tc.name)

		borderInst := borderNode.CreateInstance()
		if measurable, ok := borderInst.(interface{ Measure(layout.Constraints) layout.Size }); ok {
			constraints := layout.Constraints{
				MinWidth:  0,
				MaxWidth:  tc.maxWidth,
				MinHeight: 0,
				MaxHeight: tc.maxHeight,
			}
			size := measurable.Measure(constraints)
			fmt.Printf("  Constraints: MaxWidth=%d, MaxHeight=%d\n", constraints.MaxWidth, constraints.MaxHeight)
			fmt.Printf("  Border.Measured: %dx%d\n", size.Width, size.Height)

			// Check Border's measuredChildSize
			if bi, ok := borderInst.(interface{ GetMeasuredChildSize() layout.Size }); ok {
				csize := bi.GetMeasuredChildSize()
				fmt.Printf("  Child (VStack) measured: %dx%d\n", csize.Width, csize.Height)
			}
		}
	}

	// Test without explicit width (auto)
	fmt.Println("\n\nBorder without explicit width:")
	borderNode2 := border.New()
	borderNode2.SetChild(vstackNode)
	borderNode2.SetBorderLabel(" Demo ")

	borderInst2 := borderNode2.CreateInstance()
	if measurable, ok := borderInst2.(interface{ Measure(layout.Constraints) layout.Size }); ok {
		constraints := layout.Constraints{
			MinWidth:  0,
			MaxWidth:  50,
			MinHeight: 0,
			MaxHeight: layout.MaxInt,
		}
		size := measurable.Measure(constraints)
		fmt.Printf("  Border.Measured: %dx%d\n", size.Width, size.Height)
	}
}
