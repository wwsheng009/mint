// Fiber-first Text Component Demo
// Demonstrates the new text component following the Fiber-first architecture
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/internal/render"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	newbutton "github.com/wwsheng009/mint/ui/components/button"
	newtext "github.com/wwsheng009/mint/ui/components/text"
)

// DemoApp renders text and buttons using the Fiber-first components
func DemoApp() rtui.VNode {
	// Create a root element to hold all components
	root := rtui.NewElement("div")

	// Create text components with various styles
	text1 := newtext.New("Welcome to Fiber-first Text Demo!").Foreground(theme.Primary())
	text2 := newtext.New("This demonstrates the new text component.")
	text3 := newtext.New("") // Empty line
	text4 := newtext.New("Bold Text").Bold(true)
	text5 := newtext.New("Underlined Text").Underline(true)
	text6 := newtext.New("Colored Text").Foreground(theme.Error())
	text7 := newtext.New("") // Empty line

	// Create buttons for comparison
	btn1 := newbutton.New("Click Me").SetVariant(newbutton.VariantPrimary).SetSize(newbutton.SizeMedium)
	btn2 := newbutton.New("Cancel").SetVariant(newbutton.VariantDefault).SetSize(newbutton.SizeSmall)

	// Set all children
	root.SetChildren([]rtui.VNode{
		text1, text2, text3,
		text4, text5, text6, text7,
		btn1, btn2,
	})

	return root
}

func main() {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")
	os.Setenv("MINT_DEBUG_TEST", "true")

	fmt.Println("╔══════════════════════════════════════════════════╗")
	fmt.Println("║   Fiber-First Text Rendering Demo                ║")
	fmt.Println("║   (Text + Button components)                      ║")
	fmt.Println("╚══════════════════════════════════════════════════╝")

	// Create framework app (required for Fiber reconciler)
	fwApp := framework.NewApp()

	// Create DeclarativeNode WITH Fiber reconciler
	node := render.NewDeclarativeNodeFromFuncWithFiber(DemoApp, fwApp)

	// Enable Fiber-first mode
	node.SetRenderMode(render.RenderModeFiberFirst)

	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("  Render Mode: %v\n", node.GetRenderMode())
	fmt.Printf("  Fiber-First Enabled: %v\n", node.IsFiberFirstEnabled())

	// Create buffer (60x18)
	buf := paint.NewBuffer(60, 18)

	// Create paint context
	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 60, Height: 18},
		AvailableWidth:  60,
		AvailableHeight: 18,
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 60))
	fmt.Println("Rendering text and buttons with Fiber-first pipeline...")
	fmt.Printf("%s\n\n", strings.Repeat("=", 60))

	// Render
	node.Paint(ctx, buf)

	// Output result
	printBuffer(buf, 60, 18)

	fmt.Println("\nText Size Formula:")
	fmt.Println("  width = len(content) + left_padding + right_padding")
	fmt.Println("  height = 1 (always single line)")
	fmt.Println()
	fmt.Println("Expected text widths:")
	fmt.Println("  \"Welcome to Fiber-first Text Demo!\" = 33")
	fmt.Println("  \"Bold Text\" = 9")
	fmt.Println("  Empty text = 1 (minimal space)")
}

func printBuffer(buf *paint.Buffer, width, height int) {
	fmt.Printf("┌%s┐\n", strings.Repeat("─", width))
	for y := 0; y < height; y++ {
		var line strings.Builder
		for x := 0; x < width; x++ {
			cell := buf.GetContent(x, y)
			if cell.Cluster != "" {
				line.WriteString(cell.Cluster)
			} else {
				line.WriteString(" ")
			}
		}
		fmt.Printf("|%-*s|\n", width, strings.TrimRight(line.String(), " "))
	}
	fmt.Printf("└%s┘\n", strings.Repeat("─", width))
}
