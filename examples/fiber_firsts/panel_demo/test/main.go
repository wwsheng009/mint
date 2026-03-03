// Package main demonstrates the Fiber-first Panel component.
// Panel is a high-level container that manages borders, headers, and content layout.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/wwsheng009/mint/examples/utils"
	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/internal/render"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
)

// DemoApp creates the demo UI
func DemoApp() rtui.VNode {
	return ui.NewVStack().SingleBorder().SetChildrenList([]ui.VNode{ui.Text("Single Border")}).Build()
}

func main() {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")

	fmt.Println("╔══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║   Fiber-First Panel Component Demo                               ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════╝")

	// Create framework app (required for Fiber reconciler)
	fwApp := framework.NewApp()

	// Create DeclarativeNode WITH Fiber reconciler
	node := render.NewDeclarativeNodeFromFuncWithFiber(DemoApp)
    node.SetApp(fwApp)

	// Enable Fiber-first mode
	node.SetRenderMode(render.RenderModeFiberFirst)

	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("  Render Mode: %v\n", node.GetRenderMode())

	// Create buffer (70 wide, 90 tall)
	buf := paint.NewBuffer(70, 90)

	// Create paint context
	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 70, Height: 90},
		AvailableWidth:  70,
		AvailableHeight: 90,
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 70))
	fmt.Println("Rendering Panel components...")
	fmt.Printf("%s\n\n", strings.Repeat("=", 70))

	// Render
	node.Paint(ctx, buf)

	// Output result
	utils.PrintBuffer(buf, 70, 90)

	// Debug: Print each coordinate of the buffer
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("Buffer Coordinates (first 40 rows, focused on panel area)")
	fmt.Println(strings.Repeat("=", 70))
	utils.PrintBufferCoordinates(buf, 70, 90)

	// Get layout boxes for debugging
	// var boxes []*layout.LayoutBox
	// nodeBoxes := node.GetLayoutBoxes()
	// if nodeBoxes != nil {
	// 	boxes = nodeBoxes
	// }

	// Print layout box debug info
	// fmt.Println("\n" + strings.Repeat("=", 70))
	// fmt.Println("Layout Box Debug Info (Flattened)")
	// fmt.Println(strings.Repeat("=", 70))
	// printLayoutBoxes(boxes)

	// Print layout tree with hierarchical structure
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Print(node.GetLayoutTreeString())

	// Print paintable tree with hierarchical structure
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Print(node.GetPaintableTreeString())

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("Panel Component Features:")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("  Layout Options:")
	fmt.Println("    - Width/Height: Fixed dimensions")
	fmt.Println("    - Height=0: Auto height based on content")
	fmt.Println("    - Flex: Flex factor for expansion")
	fmt.Println("    - Padding: Inner padding")
	fmt.Println("")
	fmt.Println("  Border Styles:")
	fmt.Println("    - Rounded: ╭╮╰╯ (default for titled panels)")
	fmt.Println("    - Single:  ┌┐└┘")
	fmt.Println("    - Double:  ╔╗╚╝")
	fmt.Println("    - None:    No border")
	fmt.Println("")
	fmt.Println("  Content Areas:")
	fmt.Println("    - Header: Optional header component")
	fmt.Println("    - Content: Main content (required)")
	fmt.Println("    - Footer: Optional footer component")
	fmt.Println("")
	fmt.Println("  Styling:")
	fmt.Println("    - BorderColor: Color of the border")
	fmt.Println("    - Title: Sets title (appears in border label)")
	fmt.Println("    - Label: Custom border label")
	fmt.Println("")
	fmt.Println("  Convenience Functions:")
	fmt.Println("    - panel.Of(content)")
	fmt.Println("    - panel.OfSize(content, w, h)")
	fmt.Println("    - panel.Titled(title, content)")
	fmt.Println("    - panel.Bordered(content, w, h)")
	fmt.Println(strings.Repeat("=", 70))
}
