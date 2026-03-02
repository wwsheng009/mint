// Package main demonstrates the Progress component following the Fiber-first architecture.
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
	"github.com/wwsheng009/mint/ui/components/progress"
	newtext "github.com/wwsheng009/mint/ui/components/text"
)

// DemoApp renders Progress components using the Fiber-first pipeline
func DemoApp() rtui.VNode {
	return ui.NewVStack().
		SetWidth(50).
		SetGap(1).
		SetChildrenList([]rtui.VNode{
			// Title
			newtext.New("=== Progress Demo ==="),

			// Section 1: Basic Progress Bars
			newtext.New(""),
			newtext.New("1. Basic Progress Bars:"),
			ui.NewVStack().
				SetGap(0).
				SetChildrenList([]rtui.VNode{
					progress.New().SetValue(0).SetWidth(30),
					progress.New().SetValue(25).SetWidth(30),
					progress.New().SetValue(50).SetWidth(30),
					progress.New().SetValue(75).SetWidth(30),
					progress.New().SetValue(100).SetWidth(30),
				}),

			// Section 2: With Labels
			newtext.New(""),
			newtext.New("2. With Labels:"),
			ui.NewVStack().
				SetGap(0).
				SetChildrenList([]rtui.VNode{
					progress.New().SetValue(30).SetLabel("Downloading").SetWidth(35),
					progress.New().SetValue(60).SetLabel("Processing").SetWidth(35),
					progress.New().SetValue(90).SetLabel("Almost done").SetWidth(35),
				}),

			// Section 3: Without Percentage
			newtext.New(""),
			newtext.New("3. Without Percentage:"),
			ui.NewVStack().
				SetGap(0).
				SetChildrenList([]rtui.VNode{
					progress.New().SetValue(50).SetLabel("Loading...").SetShowPercent(false).SetWidth(30),
					progress.New().SetValue(80).SetShowPercent(false).SetWidth(30),
				}),

			// Section 4: Different Widths
			newtext.New(""),
			newtext.New("4. Different Widths:"),
			ui.NewVStack().
				SetGap(0).
				SetChildrenList([]rtui.VNode{
					progress.New().SetValue(50).SetWidth(10),
					progress.New().SetValue(50).SetWidth(20),
					progress.New().SetValue(50).SetWidth(40),
				}),

			// Section 5: Custom Max Value
			newtext.New(""),
			newtext.New("5. Custom Max Value (200):"),
			ui.NewVStack().
				SetGap(0).
				SetChildrenList([]rtui.VNode{
					progress.New().SetValue(50).SetMax(200).SetLabel("50/200").SetWidth(30),
					progress.New().SetValue(150).SetMax(200).SetLabel("150/200").SetWidth(30),
				}),
		})
}

func main() {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")

	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║   Fiber-First Progress Rendering Demo                  ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")

	fwApp := framework.NewApp()
	node := render.NewDeclarativeNodeFromFuncWithFiber(DemoApp, fwApp)
	node.SetRenderMode(render.RenderModeFiberFirst)

	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("  Render Mode: %v\n", node.GetRenderMode())

	buf := paint.NewBuffer(55, 34)
	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 55, Height: 34},
		AvailableWidth:  55,
		AvailableHeight: 34,
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 55))
	fmt.Println("Rendering Progress components...")
	fmt.Printf("%s\n\n", strings.Repeat("=", 55))

	node.Paint(ctx, buf)
	utils.PrintBuffer(buf, 55, 34)

	fmt.Println("\n" + strings.Repeat("=", 55))
	fmt.Println("Progress Component Features:")
	fmt.Println(strings.Repeat("=", 55))
	fmt.Println("  - Progress bar display: [======>     ]")
	fmt.Println("  - Configurable value and max")
	fmt.Println("  - Percentage display")
	fmt.Println("  - Label support")
	fmt.Println("  - Auto-sizing based on width")
	fmt.Println("")
	fmt.Println("Layout:")
	fmt.Println("  - Width: configurable (min 10)")
	fmt.Println("  - Height: 1 (bar only) or 2 (with label)")
	fmt.Println("")
	fmt.Println("Bar Format:")
	fmt.Println("  - '=' for filled area")
	fmt.Println("  - '>' for progress indicator")
	fmt.Println("  - ' ' for empty area")
	fmt.Println(strings.Repeat("=", 55))
}
