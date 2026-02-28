// Package main demonstrates the Input component following the Fiber-first architecture.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/wwsheng009/mint/examples/utils"
	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/internal/render"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/input"
	newstack "github.com/wwsheng009/mint/ui/components/stack"
	newtext "github.com/wwsheng009/mint/ui/components/text"
)

// DemoApp renders Input components using the Fiber-first pipeline
func DemoApp() rtui.VNode {
	return newstack.New(newstack.Column).
		SetWidth(60).
		SetGap(1).
		SetChildrenList([]rtui.VNode{
			// Title
			newtext.New("=== Input Demo ==="),

			// Section 1: Border Styles
			newtext.New(""),
			newtext.New("1. Border Styles:"),
			newstack.New(newstack.Column).
				SetGap(0).
				SetChildrenList([]rtui.VNode{
					input.New().SetPlaceholder("Single border (default)").SetWidth(25),
					input.New().SetPlaceholder("Double border").SetBorderStyle(layout.BorderDouble).SetWidth(25),
					input.New().SetPlaceholder("Rounded border").SetBorderStyle(layout.BorderRounded).SetWidth(25),
					input.New().SetPlaceholder("No border").SetNoBorder().SetWidth(25),
				}),

			// Section 2: Input Types
			newtext.New(""),
			newtext.New("2. Input Types:"),
			newstack.New(newstack.Column).
				SetGap(0).
				SetChildrenList([]rtui.VNode{
					input.New().SetValue("Text input").SetWidth(25),
					input.New().SetPassword().SetValue("secret").SetWidth(25),
					input.New().SetPlaceholder("Email input").SetWidth(25),
				}),

			// Section 3: States
			newtext.New(""),
			newtext.New("3. States:"),
			newstack.New(newstack.Column).
				SetGap(0).
				SetChildrenList([]rtui.VNode{
					input.New().SetValue("Normal").SetWidth(25),
					input.New().SetValue("Disabled").SetDisabled(true).SetWidth(25),
					input.New().SetValue("Read-only").SetReadOnly(true).SetWidth(25),
				}),
		})
}

func main() {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")

	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║   Fiber-First Input Rendering Demo                         ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")

	// Create framework app (required for Fiber reconciler)
	fwApp := framework.NewApp()

	// Create DeclarativeNode WITH Fiber reconciler
	node := render.NewDeclarativeNodeFromFuncWithFiber(DemoApp, fwApp)

	// Enable Fiber-first mode
	node.SetRenderMode(render.RenderModeFiberFirst)

	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("  Render Mode: %v\n", node.GetRenderMode())

	// Create buffer
	buf := paint.NewBuffer(65, 30)

	// Create paint context
	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 65, Height: 30},
		AvailableWidth:  65,
		AvailableHeight: 30,
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 65))
	fmt.Println("Rendering Input components...")
	fmt.Printf("%s\n\n", strings.Repeat("=", 65))

	// Render
	node.Paint(ctx, buf)

	// Output result
	utils.PrintBuffer(buf, 65, 30)

	fmt.Println("\n" + strings.Repeat("=", 65))
	fmt.Println("Input Component Features:")
	fmt.Println(strings.Repeat("=", 65))
	fmt.Println("  - Text input with cursor positioning")
	fmt.Println("  - Placeholder support")
	fmt.Println("  - Password masking")
	fmt.Println("  - Max length constraint")
	fmt.Println("  - Disabled/Read-only states")
	fmt.Println("  - Focusable with intent-based events")
	fmt.Println("")
	fmt.Println("Layout:")
	fmt.Println("  - Min-width: 10")
	fmt.Println("  - Height: 1")
	fmt.Println(strings.Repeat("=", 65))
}
