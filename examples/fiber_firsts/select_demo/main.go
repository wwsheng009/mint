// Package main demonstrates the Select component following the Fiber-first architecture.
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
	newselect "github.com/wwsheng009/mint/ui/components/select"
	newtext "github.com/wwsheng009/mint/ui/components/text"
)

// DemoApp renders Select components using the Fiber-first pipeline
func DemoApp() rtui.VNode {
	return newstack.New(newstack.Column).
		SetWidth(50).
		SetGap(1).
		SetChildrenList([]rtui.VNode{
			// Title
			newtext.New("=== Select Demo ==="),

			// Section 1: Basic Select
			newtext.New(""),
			newtext.New("1. Basic Select:"),
			newstack.New(newstack.Column).
				SetGap(0).
				SetChildrenList([]rtui.VNode{
					newselect.New().
						AddOption("opt1", "Option 1").
						AddOption("opt2", "Option 2").
						AddOption("opt3", "Option 3").
						SetSelectedIndex(0),
					newselect.New().
						AddOption("apple", "Apple").
						AddOption("banana", "Banana").
						AddOption("orange", "Orange").
						SetSelectedIndex(1),
				}),

			// Section 2: Disabled Select
			newtext.New(""),
			newtext.New("2. States:"),
			newstack.New(newstack.Column).
				SetGap(0).
				SetChildrenList([]rtui.VNode{
					newselect.New().
						AddOption("a", "Enabled Select").
						SetSelectedIndex(0),
					newselect.New().
						AddOption("b", "Disabled Select").
						SetDisabled(true).
						SetSelectedIndex(0),
				}),

			// Section 3: Long Options
			newtext.New(""),
			newtext.New("3. Long Options:"),
			newstack.New(newstack.Column).
				SetGap(0).
				SetChildrenList([]rtui.VNode{
					newselect.New().
						AddOption("short", "Short").
						AddOption("medium", "Medium Length").
						AddOption("long", "This is a very long option label").
						AddOption("longest", "The longest option label in the list").
						SetSelectedIndex(2),
				}),

			// Section 4: No Selection
			newtext.New(""),
			newtext.New("4. No Selection:"),
			newstack.New(newstack.Column).
				SetGap(0).
				SetChildrenList([]rtui.VNode{
					newselect.New().
						AddOption("opt1", "First Option").
						AddOption("opt2", "Second Option"),
				}),
		})
}

func main() {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")

	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║   Fiber-First Select Rendering Demo                    ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")

	fwApp := framework.NewApp()
	node := render.NewDeclarativeNodeFromFuncWithFiber(DemoApp, fwApp)
	node.SetRenderMode(render.RenderModeFiberFirst)

	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("  Render Mode: %v\n", node.GetRenderMode())

	buf := paint.NewBuffer(55, 24)
	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 55, Height: 24},
		AvailableWidth:  55,
		AvailableHeight: 24,
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 55))
	fmt.Println("Rendering Select components...")
	fmt.Printf("%s\n\n", strings.Repeat("=", 55))

	node.Paint(ctx, buf)
	printBuffer(buf, 55, 24)

	fmt.Println("\n" + strings.Repeat("=", 55))
	fmt.Println("Select Component Features:")
	fmt.Println(strings.Repeat("=", 55))
	fmt.Println("  - Dropdown display: < label >")
	fmt.Println("  - Option navigation (up/down)")
	fmt.Println("  - Disabled state")
	fmt.Println("  - Auto-sizing based on longest option")
	fmt.Println("  - Intent-based events (no closures)")
	fmt.Println("")
	fmt.Println("Actions:")
	fmt.Println("  - select/click/enter/down: Next option")
	fmt.Println("  - up: Previous option")
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
