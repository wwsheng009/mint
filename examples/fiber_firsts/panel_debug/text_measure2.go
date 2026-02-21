package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/ui/components/text"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("Text.Measure in Detail")
	fmt.Println("========================================")

	testText := text.New("This text is too long and will be wrapped to multiple lines.").SetWrap(true)

	testCases := []struct {
		name        string
		maxWidth    int
		maxHeight   int
	}{
		{"MaxWidth=16, MaxHeight=MaxInt", 16, layout.MaxInt},
		{"MaxWidth=18, MaxHeight=MaxInt", 18, layout.MaxInt},
		{"MaxWidth=50, MaxHeight=MaxInt", 50, layout.MaxInt},
		{"MaxWidth=MaxInt, MaxHeight=MaxInt", layout.MaxInt, layout.MaxInt},
		{"MaxWidth=50, MaxHeight=2", 50, 2},
		{"MaxWidth=50, MaxHeight=50", 50, 50},
	}

	for _, tc := range testCases {
		fmt.Printf("\n%s\n", tc.name)

		// Create fresh instance for each test
		inst := testText.CreateInstance()

		constraints := layout.Constraints{
			MinWidth:  0,
			MaxWidth:  tc.maxWidth,
			MinHeight: 0,
			MaxHeight: tc.maxHeight,
		}

		if measurable, ok := inst.(interface{ Measure(layout.Constraints) layout.Size }); ok {
			size := measurable.Measure(constraints)
			fmt.Printf("  Constraints: MaxWidth=%d, MaxHeight=%d\n", constraints.MaxWidth, constraints.MaxHeight)
			fmt.Printf("  Measured: %dx%d\n", size.Width, size.Height)

			if wrapGetter, ok := inst.(interface{ GetWrapLines() []string }); ok {
				lines := wrapGetter.GetWrapLines()
				fmt.Printf("  Wrap Lines: %d\n", len(lines))
				for i, line := range lines {
					fmt.Printf("    [%d] %q (len=%d)\n", i, line, len(line))
				}
			}
		}
	}
}
