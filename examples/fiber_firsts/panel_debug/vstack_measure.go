package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/ui/components/stack"
	"github.com/wwsheng009/mint/ui/components/text"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("VStack Measurement Test")
	fmt.Println("========================================")

	// Panel creates VBox with flex(1) wrapped content
	textNode := text.New("This text is too long and will be wrapped to multiple lines.").SetWrap(true)
	wrappedNode := rtui.Flex(textNode, 1)

	vstackNode := stack.New(stack.Column)
	vstackNode.SetChildrenList([]rtui.VNode{wrappedNode})

	testCases := []struct {
		name      string
		maxWidth  int
		maxHeight int
	}{
		{"No height constraint", 20, layout.MaxInt},
		{"Height constraint=10", 20, 10},
		{"Height constraint=5", 20, 5},
	}

	for _, tc := range testCases {
		fmt.Printf("\n%s\n", tc.name)

		vstackInst := vstackNode.CreateInstance()
		if measurable, ok := vstackInst.(interface{ Measure(layout.Constraints) layout.Size }); ok {
			constraints := layout.Constraints{
				MinWidth:  0,
				MaxWidth:  tc.maxWidth,
				MinHeight: 0,
				MaxHeight: tc.maxHeight,
			}
			size := measurable.Measure(constraints)
			fmt.Printf("  Constraints: MaxWidth=%d, MaxHeight=%d\n", constraints.MaxWidth, constraints.MaxHeight)
			fmt.Printf("  VStack.Measured: %dx%d\n", size.Width, size.Height)

			// Check Stack's childMeasure cache if accessible
			if si, ok := vstackInst.(interface {
				GetChildMeasure(i int) layout.Size
			}); ok {
				csize := si.GetChildMeasure(0)
				fmt.Printf("  Child[0] (flex Text) measured: %dx%d\n", csize.Width, csize.Height)
			}
		}
	}

	// Test empty VStack
	fmt.Println("\n\nEmpty VStack:")
	emptyVStack := stack.New(stack.Column)
	emptyInst := emptyVStack.CreateInstance()
	if measurable, ok := emptyInst.(interface{ Measure(layout.Constraints) layout.Size }); ok {
		constraints := layout.Constraints{
			MinWidth:  0,
			MaxWidth:  20,
			MinHeight: 0,
			MaxHeight: layout.MaxInt,
		}
		size := measurable.Measure(constraints)
		fmt.Printf("  Measured: %dx%d\n", size.Width, size.Height)
	}
}
