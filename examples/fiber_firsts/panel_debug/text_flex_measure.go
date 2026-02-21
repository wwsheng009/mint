package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/ui/components/text"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("Text Measurement with Flex Property")
	fmt.Println("========================================")

	// Text with flex=1 (like Panel wraps content)
	textNode := text.New("This text is too long and will be wrapped to multiple lines.").SetWrap(true)
	textNode.SetProp("flex", 1)  // Add flex property like Panel does

	textInst := textNode.CreateInstance()
	if measurable, ok := textInst.(interface{ Measure(layout.Constraints) layout.Size }); ok {
		testCases := []struct {
			name      string
			maxWidth  int
			maxHeight int
		}{
			{"No height constraint (MaxInt)", 18, layout.MaxInt},
			{"Height constraint=5", 18, 5},
			{"Height constraint=3", 18, 3},
			{"Height constraint=10", 18, 10},
		}

		for _, tc := range testCases {
			constraints := layout.Constraints{
				MinWidth:  0,
				MaxWidth:  tc.maxWidth,
				MinHeight: 0,
				MaxHeight: tc.maxHeight,
			}
			size := measurable.Measure(constraints)
			fmt.Printf("\nText.Measure %s (MaxWidth=%d, MaxHeight=%d): %dx%d\n",
				tc.name, tc.maxWidth, tc.maxHeight, size.Width, size.Height)

			if wrapGetter, ok := textInst.(interface{ GetWrapLines() []string }); ok {
				lines := wrapGetter.GetWrapLines()
				fmt.Printf("  Wrap lines: %d\n", len(lines))
			}

			// Also test from props
			props := textInst.(interface{ Props() map[string]interface{} })
			if p := props.Props(); p != nil {
				if flex, ok := p["flex"]; ok {
					fmt.Printf("  flex prop: %v\n", flex)
				}
			}
		}
	}
}
