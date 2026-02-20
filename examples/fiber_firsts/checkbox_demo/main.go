// Package main demonstrates the Checkbox component following the Fiber-first architecture.
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
	"github.com/wwsheng009/mint/ui/components/checkbox"
	newstack "github.com/wwsheng009/mint/ui/components/stack"
	newtext "github.com/wwsheng009/mint/ui/components/text"
)

// DemoApp renders Checkbox components using the Fiber-first pipeline
func DemoApp() rtui.VNode {
	return newstack.New(newstack.Column).
		SetWidth(50).
		SetGap(1).
		SetChildrenList([]rtui.VNode{
			// Title
			newtext.New("=== Checkbox Demo ==="),

			// Section 1: Basic Checkboxes
			newtext.New(""),
			newtext.New("1. Basic Checkboxes:"),
			newstack.New(newstack.Column).
				SetGap(0).
				SetChildrenList([]rtui.VNode{
					checkbox.New("Unchecked option").SetChecked(false),
					checkbox.New("Checked option").SetChecked(true),
					checkbox.New("Disabled unchecked").SetDisabled(true).SetChecked(false),
					checkbox.New("Disabled checked").SetDisabled(true).SetChecked(true),
				}),

			// Section 2: Different states
			newtext.New(""),
			newtext.New("2. Interaction States:"),
			newstack.New(newstack.Row).
				SetGap(2).
				SetChildrenList([]rtui.VNode{
					checkbox.New("Normal"),
					checkbox.New("Hovered"), // Would be hovered in real app
					checkbox.New("Focused"), // Would be focused in real app
				}),

			// Section 3: Form example
			newtext.New(""),
			newtext.New("3. Form Example:"),
			newstack.New(newstack.Column).
				SetGap(0).
				SetChildrenList([]rtui.VNode{
					checkbox.New("I agree to the terms and conditions").SetChecked(true),
					checkbox.New("Subscribe to newsletter").SetChecked(false),
					checkbox.New("Remember me").SetChecked(true),
				}),
		})
}

func main() {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")
	os.Setenv("MINT_DEBUG_TEST", "false")

	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║   Fiber-First Checkbox Rendering Demo                  ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")

	// Create framework app (required for Fiber reconciler)
	fwApp := framework.NewApp()

	// Create DeclarativeNode WITH Fiber reconciler
	node := render.NewDeclarativeNodeFromFuncWithFiber(DemoApp, fwApp)

	// Enable Fiber-first mode
	node.SetRenderMode(render.RenderModeFiberFirst)

	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("  Render Mode: %v\n", node.GetRenderMode())
	fmt.Printf("  Fiber-First Enabled: %v\n", node.IsFiberFirstEnabled())

	// Create buffer
	buf := paint.NewBuffer(55, 20)

	// Create paint context
	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 55, Height: 20},
		AvailableWidth:  55,
		AvailableHeight: 20,
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 55))
	fmt.Println("Rendering Checkbox components...")
	fmt.Printf("%s\n\n", strings.Repeat("=", 55))

	// Render
	node.Paint(ctx, buf)

	// Output result
	printBuffer(buf, 55, 20)

	fmt.Println("\n" + strings.Repeat("=", 55))
	fmt.Println("Checkbox Component Features:")
	fmt.Println(strings.Repeat("=", 55))
	fmt.Println("  - [X] / [ ] indicator")
	fmt.Println("  - Label support")
	fmt.Println("  - Disabled state")
	fmt.Println("  - Focusable (when enabled)")
	fmt.Println("  - Intent-based events (no closures)")
	fmt.Println("")
	fmt.Println("Layout:")
	fmt.Println("  - Width: 4 + len(label)")
	fmt.Println("  - Height: 1")
	fmt.Println(strings.Repeat("=", 55))
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
		trimmed := strings.TrimRight(line.String(), " ")
		if trimmed != "" {
			fmt.Printf("|%-*s|\n", width, trimmed)
		} else if y < height-1 {
			fmt.Printf("|%s|\n", strings.Repeat(" ", width))
		}
	}
	fmt.Printf("└%s┘\n", strings.Repeat("─", width))
}
