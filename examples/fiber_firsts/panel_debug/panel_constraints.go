package main

import (
	"fmt"

	"github.com/wwsheng009/mint/internal/render"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/ui/components/panel"
	"github.com/wwsheng009/mint/ui/components/text"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("Panel.Measure with Various Constraints")
	fmt.Println("========================================")

	panelNode := panel.NewBuilder().
		Title("With Wrap").
		Content(text.New("This text is too long and will be wrapped to multiple lines.").SetWrap(true)).
		Width(20).
		Build()

	testCases := []struct {
		name        string
		maxWidth    int
		maxHeight   int
	}{
		{"MaxInt constraints", layout.MaxInt, layout.MaxInt},
		{"Bounded 50x50", 50, 50},
		{"Bounded 20x10", 20, 10},
		{"Bounded 20x5", 20, 5},
		{"Bounded 20x3", 20, 3},
		{"Bounded 20x2", 20, 2},
	}

	for _, tc := range testCases {
		fmt.Printf("\n%s\n", tc.name)

		engine := layout.NewEngine()
		adapter := render.NewVNodeToNodeAdapter(panelNode)

		constraints := layout.Constraints{
			MinWidth:  0,
			MaxWidth:  tc.maxWidth,
			MinHeight: 0,
			MaxHeight: tc.maxHeight,
		}

		size := engine.Measure(adapter, constraints)
		fmt.Printf("  Constraints: MaxWidth=%d, MaxHeight=%d\n", constraints.MaxWidth, constraints.MaxHeight)
		fmt.Printf("  Panel Measured: %dx%d\n", size.Width, size.Height)

		// Also test the panel in a Row context (with another panel)
		if tc.maxWidth >= 50 {
			fmt.Println("\n  (Testing in Row with another panel)")
			rowNode := render.NewVNodeToNodeAdapter(panelNode)
			rowSize := engine.Measure(rowNode, constraints)
			fmt.Printf("  Row Measured: %dx%d\n", rowSize.Width, rowSize.Height)
		}
	}
}
