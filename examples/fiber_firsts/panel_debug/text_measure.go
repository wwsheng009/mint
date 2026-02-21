package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/ui/components/text"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("Text.Measure with Constraints")
	fmt.Println("========================================")

	// Case 1: Text with width constraint
	fmt.Println("\nCase 1: Text with wrap=true, MaxWidth=16")
	text1 := text.New("This text is too long and will be wrapped to multiple lines.").SetWrap(true)
	inst1 := text1.CreateInstance()

	// Measure with MaxWidth=16 (from border inner area)
	constraints1 := layout.Constraints{
		MinWidth:  0,
		MaxWidth:  16,
		MinHeight: 0,
		MaxHeight: layout.MaxInt, // No height constraint
	}

	if measurable, ok := inst1.(interface{ Measure(layout.Constraints) layout.Size }); ok {
		size := measurable.Measure(constraints1)
		fmt.Printf("  Measured: %dx%d\n", size.Width, size.Height)

		if wrapGetter, ok := inst1.(interface{ GetWrapLines() []string }); ok {
			lines := wrapGetter.GetWrapLines()
			fmt.Printf("  Wrap lines: %d\n", len(lines))
		}
	}

	// Case 2: Text constrained by Panel layout
	// Panel Width=20, Border=1x1, inner width=18
	fmt.Println("\nCase 2: Text with wrap=true, MaxWidth=18 (Panel inner width)")
	text2 := text.New("This text is too long and will be wrapped to multiple lines.").SetWrap(true)
	inst2 := text2.CreateInstance()

	constraints2 := layout.Constraints{
		MinWidth:  0,
		MaxWidth:  18,
		MinHeight: 0,
		MaxHeight: layout.MaxInt,
	}

	if measurable, ok := inst2.(interface{ Measure(layout.Constraints) layout.Size }); ok {
		size := measurable.Measure(constraints2)
		fmt.Printf("  Measured: %dx%d\n", size.Width, size.Height)

		if wrapGetter, ok := inst2.(interface{ GetWrapLines() []string }); ok {
			lines := wrapGetter.GetWrapLines()
			fmt.Printf("  Wrap lines: %d\n", len(lines))
		}
	}

	// Case 3: Text constrained by border=1 from each side
	// Panel Width=20, VBox padding=0, border width=1
	// So inner width = 20 - 2*1 - 0 = 18
	fmt.Println("\nCase 3: Text with wrap=true, MaxWidth=18, MaxHeight=5")
	text3 := text.New("This text is too long and will be wrapped to multiple lines.").SetWrap(true)
	inst3 := text3.CreateInstance()

	constraints3 := layout.Constraints{
		MinWidth:  0,
		MaxWidth:  18,
		MinHeight: 0,
		MaxHeight: 5,
	}

	if measurable, ok := inst3.(interface{ Measure(layout.Constraints) layout.Size }); ok {
		size := measurable.Measure(constraints3)
		fmt.Printf("  Measured: %dx%d\n", size.Width, size.Height)

		if wrapGetter, ok := inst3.(interface{ GetWrapLines() []string }); ok {
			lines := wrapGetter.GetWrapLines()
			fmt.Printf("  Wrap lines: %d\n", len(lines))
		}
	}

	// Case 4: Text with MaxHeight=2
	fmt.Println("\nCase 4: Text with wrap=true, MaxWidth=18, MaxHeight=2")
	text4 := text.New("This text is too long and will be wrapped to multiple lines.").SetWrap(true)
	inst4 := text4.CreateInstance()

	constraints4 := layout.Constraints{
		MinWidth:  0,
		MaxWidth:  18,
		MinHeight: 0,
		MaxHeight: 2,
	}

	if measurable, ok := inst4.(interface{ Measure(layout.Constraints) layout.Size }); ok {
		size := measurable.Measure(constraints4)
		fmt.Printf("  Measured: %dx%d\n", size.Width, size.Height)

		if wrapGetter, ok := inst4.(interface{ GetWrapLines() []string }); ok {
			lines := wrapGetter.GetWrapLines()
			fmt.Printf("  Wrap lines: %d\n", len(lines))
		}
	}
}
