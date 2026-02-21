package main

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/internal/render"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/ui/components/text"
	"github.com/wwsheng009/mint/ui/components/stack"
	"github.com/wwsheng009/mint/ui/components/panel"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func main() {
	os.Setenv("MINT_DEBUG_TEST", "true")
	os.Setenv("MINT_DEBUG_LAYOUT", "true")

	fmt.Println("========================================")
	fmt.Println("Row Height Analysis")
	fmt.Println("========================================")

	// Test 1: Row with both panels having fixed height
	fmt.Println("\nTest 1: Row with both panels Fixed Height (3, 5)")
	row1 := stack.New(stack.Row).SetGap(3).SetChildrenList([]rtui.VNode{
		panel.NewBuilder().Title("H3").Content(text.New("Panel 1 content")).Width(20).Height(3).Build(),
		panel.NewBuilder().Title("H5").Content(text.New("Panel 2 content")).Width(20).Height(5).Build(),
	})
	renderAndPrint(row1, 50, 10, "Row with fixed heights")

	// Test 2: Row with first panel fixed, second auto
	fmt.Println("\nTest 2: Row with Panel 1 Fixed(3), Panel 2 Auto")
	row2 := stack.New(stack.Row).SetGap(3).SetChildrenList([]rtui.VNode{
		panel.NewBuilder().Title("H3").Content(text.New("Panel 1 content")).Width(20).Height(3).Build(),
		panel.NewBuilder().Title("Auto").Content(text.New("This text is too long and will be wrapped to multiple lines.").SetWrap(true)).Width(20).Build(),
	})
	renderAndPrint(row2, 50, 10, "Row with mixed heights")

	// Test 3: Row with both auto (should use natural heights)
	fmt.Println("\nTest 3: Row with both Auto Height")
	row3 := stack.New(stack.Row).SetGap(3).SetChildrenList([]rtui.VNode{
		panel.NewBuilder().Title("Auto1").Content(text.New("Short")).Width(20).Build(),
		panel.NewBuilder().Title("Auto2").Content(text.New("Short")).Width(20).Build(),
	})
	renderAndPrint(row3, 50, 10, "Row with both auto heights")
}

func renderAndPrint(vnode rtui.VNode, bufWidth, bufHeight int, label string) {
	fmt.Printf("Rendering: %s (Available: %dx%d)\n", label, bufWidth, bufHeight)

	fwApp := framework.NewApp()
	node := render.NewDeclarativeNodeFromFuncWithFiber(func() rtui.VNode { return vnode }, fwApp)
	node.SetRenderMode(render.RenderModeFiberFirst)

	buf := paint.NewBuffer(bufWidth, bufHeight)
	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: bufWidth, Height: bufHeight},
		AvailableWidth:  bufWidth,
		AvailableHeight: bufHeight,
	}

	node.Paint(ctx, buf)

	printBuffer(buf, 15) // Print first 15 lines
}

func printBuffer(buf *paint.Buffer, maxLines int) {
	for y := 0; y < buf.Height && y < maxLines; y++ {
		line := ""
		for x := 0; x < buf.Width; x++ {
			cell := buf.Cells[y][x]
			if cell.Cluster == "" {
				line += " "
			} else {
				line += cell.Cluster
			}
		}
		fmt.Println(line)
	}
}
