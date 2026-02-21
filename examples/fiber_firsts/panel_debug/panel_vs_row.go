package main

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009\mint/framework/component"
	"github.com/wwsheng009\mint/internal/render"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009\mint/ui/components/text"
	"github.com/wwsheng009\mint/ui/components/stack"
	"github.com/wwsheng009/mint/ui/components/panel"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func main() {
	os.Setenv("MINT_DEBUG_TEST", "true")

	fmt.Println("========================================")
	fmt.Println("Panel 2 Direct Measurement")
	fmt.Println("========================================")

	// Panel 2: auto height with wrap text
	panel2 := panel.NewBuilder().
		Title("With Wrap").
		Content(text.New("This text is too long and will be wrapped to multiple lines.").SetWrap(true)).
		Width(20).
		Build()

	// Get the composed Border
	panelInst := panel2.CreateInstance()
	fmt.Printf("Panel.CreateInstance() type: %T\n", panelInst)

	if borderNode, ok := panelInst.(rtui.VNode); ok {
		fmt.Printf("Composed node type: %T, Tag: %s\n", borderNode, borderNode.Tag())

		// Measure the Border directly
		engine := layout.NewEngine()

		testCases := []struct {
			name      string
			maxWidth  int
			maxHeight int
		}{
			{"MaxInt constraints", 50, layout.MaxInt},
			{"50x10 constraints", 50, 10},
			{"20x10 constraints", 20, 10},
		}

		for _, tc := range testCases {
			fmt.Printf("\n%s:\n", tc.name)
			constraints := layout.Constraints{
				MinWidth:  0,
				MaxWidth:  tc.maxWidth,
				MinHeight: 0,
				MaxHeight: tc.maxHeight,
			}
			size := engine.Measure(borderNode, constraints)
			fmt.Printf("  Constraints: %dx%d\n", constraints.MaxWidth, constraints.MaxHeight)
			fmt.Printf("  Measured: %dx%d\n", size.Width, size.Height)
		}
	}

	fmt.Println("\n\n========================================")
	fmt.Println("Row Measurement (for comparison)")
	fmt.Println("========================================")

	// Row contains Panel 1 and Panel 2
	panel1 := panel.NewBuilder().
		Title("H3").
		Content(text.New("Panel 1")).
		Width(20).
		Height(3).
		Build()

	rowNode := stack.New(stack.Row).SetGap(3).SetChildrenList([]rtui.VNode{panel1, panel2})

	engine := layout.NewEngine()
	fwApp := framework.NewApp()
	decNode := render.NewDeclarativeNodeFromFuncWithFiber(func() rtui.VNode { return rowNode }, fwApp)

	// Measure the Row
	constraints := layout.Constraints{
		MinWidth:  0,
		MaxWidth:  50,
		MinHeight: 0,
		MaxHeight: 10,
	}

	fmt.Printf("Row Constraints: %dx%d\n", constraints.MaxWidth, constraints.MaxHeight)
	size := engine.Measure(decNode, constraints)
	fmt.Printf("Row Measured: %dx%d\n", size.Width, size.Height)
}
