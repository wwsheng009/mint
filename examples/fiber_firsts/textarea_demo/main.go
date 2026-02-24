// Package main demonstrates the Textarea component following the Fiber-first architecture.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/internal/render"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	newstack "github.com/wwsheng009/mint/ui/components/stack"
	newtext "github.com/wwsheng009/mint/ui/components/text"
	"github.com/wwsheng009/mint/ui/components/textarea"
)

// DemoApp renders Textarea components using the Fiber-first pipeline
func DemoApp() rtui.VNode {
	return newstack.New(newstack.Column).
		SetWidth(50).
		SetGap(1).
		SetChildrenList([]rtui.VNode{
			// Title
			newtext.New("=== Textarea Demo ==="),

			// Section 1: Basic Textarea
			newtext.New(""),
			newtext.New("1. Basic Textarea (3 rows):"),
			textarea.New().
				SetPlaceholder("Enter your text here...").
				SetRows(3).
				SetCols(30),

			// Section 2: With Initial Value
			newtext.New(""),
			newtext.New("2. With Initial Value:"),
			textarea.New().
				SetValue("Line 1: Hello World\nLine 2: This is a test\nLine 3: Multi-line input").
				SetRows(4).
				SetCols(35),

			// Section 3: Disabled Textarea
			newtext.New(""),
			newtext.New("3. Disabled Textarea:"),
			textarea.New().
				SetValue("This textarea is disabled.\nYou cannot edit this content.").
				SetDisabled(true).
				SetRows(2).
				SetCols(35),

			// Section 4: Different Sizes
			newtext.New(""),
			newtext.New("4. Small Textarea (2 rows, 20 cols):"),
			textarea.New().
				SetPlaceholder("Small input").
				SetRows(2).
				SetCols(20),
		})
}

func main() {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")

	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║   Fiber-First Textarea Rendering Demo                  ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")

	fwApp := framework.NewApp()
	node := render.NewDeclarativeNodeFromFuncWithFiber(DemoApp, fwApp)
	node.SetRenderMode(render.RenderModeFiberFirst)

	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("  Render Mode: %v\n", node.GetRenderMode())

	buf := paint.NewBuffer(55, 32)
	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 55, Height: 32},
		AvailableWidth:  55,
		AvailableHeight: 32,
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 55))
	fmt.Println("Rendering Textarea components...")
	fmt.Printf("%s\n\n", strings.Repeat("=", 55))

	node.Paint(ctx, buf)
	printBuffer(buf, 55, 32)

	fmt.Println("\n" + strings.Repeat("=", 55))
	fmt.Println("Textarea Component Features:")
	fmt.Println(strings.Repeat("=", 55))
	fmt.Println("  - Multi-line text input")
	fmt.Println("  - Configurable rows and columns")
	fmt.Println("  - Placeholder support")
	fmt.Println("  - Disabled/read-only states")
	fmt.Println("  - Max length constraint")
	fmt.Println("  - Border rendering with +/- and |")
	fmt.Println("")
	fmt.Println("Layout:")
	fmt.Println("  - Width = cols + 4 (for borders)")
	fmt.Println("  - Height = rows + 2 (for borders)")
	fmt.Println(strings.Repeat("=", 55))
}

func printBuffer(buf *paint.Buffer, width, height int) {
	fmt.Printf("┌%s┐\n", strings.Repeat("─", width))
	for y := 0; y < height; y++ {
		var line strings.Builder
		for x := 0; x < width; x++ {
			cell := buf.GetContent(x, y)
			// 跳过宽字符的延续单元格
			if cell.IsContinuation {
				continue
			}
			if cell.Cluster != "" {
				line.WriteString(cell.Cluster)
			} else {
				line.WriteString(" ")
			}
		}
		trimmed := strings.TrimRight(line.String(), " ")
		if trimmed != "" {
			fmt.Printf("|%-*s|\n", width, trimmed)
		} else if y < height-1 {
			fmt.Printf("|%s|\n", strings.Repeat(" ", width))
		}
	}
	fmt.Printf("└%s┘\n", strings.Repeat("─", width))
}
