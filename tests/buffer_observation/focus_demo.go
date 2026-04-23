//go:build ignore
// +build ignore

// Focus Change Buffer Test
//
// Purpose: Test buffer behavior when focus state changes
// Focus state change might affect button layout (focus marker, colors, etc.)
//
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
	"github.com/wwsheng009/mint/ui"
)

// Simple test with two buttons, focus can change
func ButtonComponent(buttonID string, focused bool) ui.VNode {
	btnText := "TEST"
	if focused {
		btnText = "*TST"  // Simulate focus indicator (shorter text)
	}

	if buttonID == "btn1" {
		return ui.NewButtonBuilder(btnText).
			Variant(ui.ButtonVariantPrimary).
			Build()
	} else {
		return ui.NewButtonBuilder(btnText).
			Variant(ui.ButtonVariantSecondary).
			Build()
	}
}

func TestView(focusedButton string) ui.VNode {
	fmt.Printf("\nRendering TestView (focused=%s)\n", focusedButton)

	return ui.VStack(
		ui.Text("Buttons Test:"),
		ui.HStack(
			ButtonComponent("btn1", focusedButton == "btn1"),
			ui.Text("   "),
			ButtonComponent("btn2", focusedButton == "btn2"),
		),
	)
}

func extractRow(buffer *paint.Buffer, row, width int) string {
	var sb strings.Builder
	for x := 0; x < width; x++ {
		cell := buffer.GetContent(x, row)
		if cell.Cluster == "" {
			sb.WriteRune(' ')
		} else {
			sb.WriteString(cell.Cluster)
		}
	}
	return sb.String()
}

func printRowDiff(rowNum int, row1, row2 string) {
	runes1 := []rune(row1)
	runes2 := []rune(row2)
	maxLen := max(len(runes1), len(runes2))

	hasDiff := false
	for i := 0; i < maxLen; i++ {
		c1 := ' '
		if i < len(runes1) {
			c1 = runes1[i]
		}
		c2 := ' '
		if i < len(runes2) {
			c2 = runes2[i]
		}
		if c1 != c2 {
			if !hasDiff {
				fmt.Printf("  🚨 Row %d:\n", rowNum)
				fmt.Printf("     Frame 1: %s\n", row1)
				fmt.Printf("     Frame 2: %s\n", row2)
				fmt.Printf("     Diffs:")
				hasDiff = true
			}
			fmt.Printf(" [%d:'%c'→'%c']", i, c1, c2)
		}
	}
	if hasDiff {
		fmt.Println()
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")

	fmt.Println("╔══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║   Focus Change Buffer Test                                       ║")
	fmt.Println("║   Testing: Button focus change, possible layout shift            ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════╝")

	fwApp := framework.NewApp()
	buf := paint.NewBuffer(80, 25)
	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 80, Height: 25},
		AvailableWidth:  80,
		AvailableHeight: 25,
	}

	// Create single node
	node := render.NewDeclarativeNodeFromFuncWithFiber(func() ui.VNode {
		return TestView("btn1")  // Start with btn1 focused
	})
	node.SetApp(fwApp)
	node.SetRenderMode(render.RenderModeFiberFirst)

	frame1Rows := make([]string, 10)
	frame2Rows := make([]string, 10)

	// Frame 1: btn1 focused
	{
		fmt.Printf("\n%s\n", strings.Repeat("=", 70))
		fmt.Println("FRAME 1: btn1 focused")
		fmt.Printf("%s\n", strings.Repeat("=", 70))

		node.Paint(ctx, buf)

		for y := 0; y < 10; y++ {
			frame1Rows[y] = extractRow(buf, y, 80)
		}

		fmt.Println("Frame 1 Buffer:")
		utils.PrintBuffer(buf, 80, 25)

		fmt.Println("\nFrame 1 Paintable Tree:")
		fmt.Print(node.GetPaintableTreeString())
	}

	// Frame 2: btn2 focused (update component function)
	{
		fmt.Printf("\n%s\n", strings.Repeat("=", 70))
		fmt.Println("FRAME 2: btn2 focused (changed component function)")
		fmt.Printf("%s\n", strings.Repeat("=", 70))

		// Update the component function (simulating focus state change)
		node2 := render.NewDeclarativeNodeFromFuncWithFiber(func() ui.VNode {
			return TestView("btn2")  // Now btn2 focused
		})
		node2.SetApp(fwApp)
		node2.SetRenderMode(render.RenderModeFiberFirst)

		// Paint WITHOUT buffer reset
		node2.Paint(ctx, buf)

		for y := 0; y < 10; y++ {
			frame2Rows[y] = extractRow(buf, y, 80)
		}

		fmt.Println("Frame 2 Buffer:")
		utils.PrintBuffer(buf, 80, 25)

		fmt.Println("\nFrame 2 Paintable Tree:")
		fmt.Print(node2.GetPaintableTreeString())

		fmt.Printf("\n%s\n", strings.Repeat("=", 70))
		fmt.Println("ANALYSIS: Compare Frame 1 vs Frame 2")
		fmt.Printf("%s\n", strings.Repeat("=", 70))

		for y := 0; y < 10; y++ {
			if frame1Rows[y] != frame2Rows[y] {
				printRowDiff(y, frame1Rows[y], frame2Rows[y])
			}
		}

		fmt.Println("\n" + strings.Repeat("=", 70))
		fmt.Println("NOTES:")
		fmt.Println(strings.Repeat("=", 70))
		fmt.Println("This test uses TWO DeclarativeNode instances (simulating a real scenario)")
		fmt.Println("where the component tree might be rebuilt.")
		fmt.Println("")
		fmt.Println("Expected behavior:")
		fmt.Println("  - Old content should be cleared")
		fmt.Println("  - Only new content should be visible")
		fmt.Println("")
		fmt.Println("Bug symptoms:")
		fmt.Println("  - Characters from btn1 (Frame 1) remain in buffer")
		fmt.Println("  - e.g., if Frame 2 button is at different position")
		fmt.Println(strings.Repeat("=", 70))
	}
}
