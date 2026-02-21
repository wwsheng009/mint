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

	rowNode := stack.New(stack.Row).
		SetGap(3).
		SetChildrenList([]rtui.VNode{
			panel.NewBuilder().
				Title("No Wrap").
				Content(text.New("This text is too long and will be truncated.").SetWrap(false)).
				Width(20).
				Height(3).
				BorderColor("red").
				Build(),
			panel.NewBuilder().
				Title("With Wrap").
				Content(text.New("This text is too long and will be wrapped to multiple lines.").SetWrap(true)).
				Width(20).
				BorderColor("green").
				Build(),
		})

	fmt.Println("========================================")
	fmt.Println("Auto-Height Panel in Row Stack")
	fmt.Println("========================================")
	fmt.Println("Row contains:")
	fmt.Println("  - Panel 1: Width=20, Height=3 (fixed)")
	fmt.Println("  - Panel 2: Width=20, Height=auto (0)")
	fmt.Println()

	fwApp := framework.NewApp()
	node := render.NewDeclarativeNodeFromFuncWithFiber(func() rtui.VNode { return rowNode }, fwApp)
	node.SetRenderMode(render.RenderModeFiberFirst)

	buf := paint.NewBuffer(50, 10)
	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 50, Height: 10},
		AvailableWidth:  50,
		AvailableHeight: 10,
	}

	fmt.Println("Paint context: AvailableWidth=50, AvailableHeight=10")
	fmt.Println()

	node.Paint(ctx, buf)

	println()
	printBuffer(buf)
}

func printBuffer(buf *paint.Buffer) {
	for y := 0; y < buf.Height; y++ {
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
