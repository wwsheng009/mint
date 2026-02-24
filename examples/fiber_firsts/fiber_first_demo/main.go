// Fiber-first Rendering Pipeline Demo with New Button Component
// Only uses ui/components/button which has been migrated to Fiber-first architecture
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
	button "github.com/wwsheng009/mint/ui/components/button"
)

// DemoApp renders buttons using the new Fiber-first button component
func DemoApp() rtui.VNode {
	// Create a root element to hold the buttons
	root := rtui.NewElement("div")

	// Add multiple buttons as children
	// Small buttons
	btn1 := button.New("Cancel").SetVariant(button.VariantSecondary).SetSize(button.SizeSmall)
	btn2 := button.New("OK").SetVariant(button.VariantPrimary).SetSize(button.SizeSmall)

	// Medium buttons
	btn3 := button.New("Delete").SetVariant(button.VariantDanger).SetSize(button.SizeMedium)
	btn4 := button.New("Confirm").SetVariant(button.VariantSuccess).SetSize(button.SizeMedium)
	btn5 := button.New("Save").SetVariant(button.VariantPrimary).SetSize(button.SizeMedium)

	// Large buttons
	btn6 := button.New("Submit").SetVariant(button.VariantPrimary).SetSize(button.SizeLarge)

	// Disabled button
	btn7 := button.New("Disabled").SetVariant(button.VariantDefault).SetSize(button.SizeMedium).SetDisabled(true)

	// Set children
	root.SetChildren([]rtui.VNode{btn1, btn2, btn3, btn4, btn5, btn6, btn7})

	return root
}

func main() {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")
	os.Setenv("MINT_DEBUG_TEST", "true")

	fmt.Println("╔══════════════════════════════════════════════════╗")
	fmt.Println("║   Fiber-First Button Rendering Demo               ║")
	fmt.Println("║   (Only migrated button component)                ║")
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

	// Create buffer (50x20)
	buf := paint.NewBuffer(50, 20)

	// Create paint context
	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 50, Height: 20},
		AvailableWidth:  50,
		AvailableHeight: 20,
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 50))
	fmt.Println("Rendering buttons with Fiber-first pipeline...")
	fmt.Printf("%s\n\n", strings.Repeat("=", 50))

	// Render
	node.Paint(ctx, buf)

	// Output result
	printBuffer(buf)

	fmt.Println("\nButton Size Reference:")
	fmt.Println("  Small:  len(label) + 3 + 0")
	fmt.Println("  Medium: len(label) + 3 + 2")
	fmt.Println("  Large:  len(label) + 3 + 4")
	fmt.Println("\nExpected sizes:")
	fmt.Println("  Cancel (Small):   6 + 3 + 0 = 9")
	fmt.Println("  OK (Small):       2 + 3 + 0 = 5")
	fmt.Println("  Delete (Medium):  6 + 3 + 2 = 11")
	fmt.Println("  Confirm (Medium): 7 + 3 + 2 = 12")
	fmt.Println("  Save (Medium):    4 + 3 + 2 = 9")
	fmt.Println("  Submit (Large):   6 + 3 + 4 = 13")
	fmt.Println("  Disabled (Medium):8 + 3 + 2 = 13")
}

func printBuffer(buf *paint.Buffer) {
	fmt.Println("┌──────────────────────────────────────────────────┐")
	for y := 0; y < buf.Height; y++ {
		var line strings.Builder
		for x := 0; x < buf.Width; x++ {
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
		fmt.Printf("|%-50s|\n", strings.TrimRight(line.String(), " "))
	}
	fmt.Println("└──────────────────────────────────────────────────┘")
}
