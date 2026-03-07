// Buffer Observation Test - Single Node Instance
//
// Purpose: Observe buffer content changes when UI content updates on the SAME node instance
// This correctly simulates the real application behavior where:
//   - One DeclarativeNode instance is maintained across frames
//   - State changes through the Store trigger re-renders
//   - Fiber reconciler maintains stable IDs
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
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/ui"
)

// =============================================================================
// AppState and Store
// =============================================================================

type AppState struct {
	Count int
}

var appStore = store.NewStore(AppState{Count: 0})

// =============================================================================
// Counter View (subscribes to store)
// =============================================================================

func CounterComponent() ui.VNode {
	// Subscribe to store
	count := appStore.Get().Count
	fmt.Printf("\nRendering CounterComponent (count=%d)\n", count)

	return ui.VStack(
		ui.NewTextBuilder("=== 2. Store 计数器 ===").
			FgColor("green").
			Build(),
		ui.HStack(
			ui.Text("计数: "),
			ui.NewTextBuilder(fmt.Sprintf("%d", count)).
				FgColor("yellow").
				Bold(true).
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder(" + ").
				Variant(ui.ButtonVariantPrimary).
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder(" - ").
				Variant(ui.ButtonVariantSecondary).
				Build(),
		),
	)
}

// =============================================================================
// Main Test
// =============================================================================

func main() {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")

	fmt.Println("╔══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║   Buffer Observation Test - Single Node Instance                ║")
	fmt.Println("║   Testing: SAME node, Store state changes across frames         ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════╝")

	// Create framework app
	fwApp := framework.NewApp()

	// Create buffer (80 wide, 25 tall)
	buf := paint.NewBuffer(80, 25)

	// Create paint context
	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 80, Height: 25},
		AvailableWidth:  80,
		AvailableHeight: 25,
	}

	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("  Buffer: %dx%d\n", buf.Width, buf.Height)
	fmt.Printf("  Store initial state: Count=%d\n", appStore.Get().Count)

	fmt.Printf("\n%s\n", strings.Repeat("=", 80))
	fmt.Println("INITIALIZING: Creating SINGLE DeclarativeNode instance")
	fmt.Printf("%s\n\n", strings.Repeat("=", 80))

	// Create SINGLE DeclarativeNode instance (this is key - same instance for all frames)
	node := render.NewDeclarativeNodeFromFuncWithFiber(CounterComponent)
	node.SetApp(fwApp)
	node.SetRenderMode(render.RenderModeFiberFirst)

	// Variables to save frame data
	frame1Rows := make([]string, 10)
	frame2Rows := make([]string, 10)

	// Frame 1: Initial render (count=0)
	{
		fmt.Printf("FRAME 1: Initial render (count=%d)\n", appStore.Get().Count)

		// Simulate app.render() WITHOUT buffer reset (this is the issue case!)
		// In real app, app.render() calls buf.Reset(), but let's see if
		// the bug can even reproduce with the same node
		fmt.Println("  Note: Buffer NOT reset between frames (simulating potential bug)")

		// Paint Frame 1
		node.Paint(ctx, buf)

		// Save Frame 1 buffer
		for y := 0; y < 10; y++ {
			frame1Rows[y] = extractRow(buf, y, 80)
		}

		// Output Frame 1
		fmt.Println("\nFrame 1 Buffer:")
		utils.PrintBuffer(buf, 80, 25)

		// Print paintable tree
		fmt.Println("\nFrame 1 Paintable Tree:")
		fmt.Print(node.GetPaintableTreeString())
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 80))
	fmt.Println("FRAME 2: Update store state and re-render SAME node")
	fmt.Printf("%s\n\n", strings.Repeat("=", 80))

	// Frame 2: Update store and re-render (count=1)
	{
		// Update store state (simulating user clicking " + " button)
		appStore.Set(AppState{Count: 1})
		fmt.Println("Updated store: Count=1")

		// Re-render SAME node (this is what happens in real app after state update)
		fmt.Println("Re-rendering same DeclarativeNode instance...")

		// Paint Frame 2 on SAME buffer (no reset)
		node.Paint(ctx, buf)

		// Save Frame 2 buffer
		for y := 0; y < 10; y++ {
			frame2Rows[y] = extractRow(buf, y, 80)
		}

		// Output Frame 2
		fmt.Println("\nFrame 2 Buffer:")
		utils.PrintBuffer(buf, 80, 25)

		// Print paintable tree
		fmt.Println("\nFrame 2 Paintable Tree:")
		fmt.Print(node.GetPaintableTreeString())

		fmt.Printf("\n%s\n", strings.Repeat("=", 80))
		fmt.Println("ANALYSIS: Comparing Frame 1 and Frame 2")
		fmt.Printf("%s\n\n", strings.Repeat("=", 80))

		// Compare frames
		for y := 0; y < 10; y++ {
			row1 := frame1Rows[y]
			row2 := frame2Rows[y]
			if row1 != row2 {
				fmt.Printf("Row %d:\n", y)
				fmt.Printf("  Frame 1: %s\n", row1)
				fmt.Printf("  Frame 2: %s\n", row2)
				printRowDifferences(y, row1, row2)
			}
		}
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 80))
	fmt.Println("SUMMARY:")
	fmt.Printf("%s\n", strings.Repeat("=", 80))
	fmt.Println("✅ Test completed")
	fmt.Println("")
	fmt.Println("Key Observations:")
	fmt.Println("  1. Same DeclarativeNode instance used for both frames")
	fmt.Println("  2.fiber reconciler maintains stable component IDs")
	fmt.Println("  3. Buffer NOT reset between frames (simulating the bug scenario)")
	fmt.Println("")
	fmt.Println("If bug is present:")
	fmt.Println("  - Old content should remain in buffer")
	fmt.Println("  - Frame 2 should mix old and new content")
	fmt.Println("")
	fmt.Println("If buffer wiping works correctly:")
	fmt.Println("  - Content in same positions should be overwritten")
	fmt.Println("  - Content at different positions might leave traces")
	fmt.Printf(strings.Repeat("=", 80))
}

// =============================================================================
// Helper Functions
// =============================================================================

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

func printRowDifferences(rowNum int, row1, row2 string) {
	diffs := CompareRows(row1, row2)
	if len(diffs) > 0 {
		fmt.Printf("  Differences at columns:")
		for _, diff := range diffs {
			c1 := ' '
			if diff < len(row1) {
				c1 = []rune(row1)[diff]
			}
			c2 := ' '
			if diff < len(row2) {
				c2 = []rune(row2)[diff]
			}
			fmt.Printf(" [%d: '%c'→'%c']", diff, c1, c2)
		}
		fmt.Println()
	}
}

func CompareRows(row1, row2 string) []int {
	diffs := []int{}
	runes1 := []rune(row1)
	runes2 := []rune(row2)
	maxLen := max(len(runes1), len(runes2))
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
			diffs = append(diffs, i)
		}
	}
	return diffs
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
