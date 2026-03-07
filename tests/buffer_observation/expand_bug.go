// Expand/Collapse Bug Reproduction
//
// Purpose: Reproduce the EXACT bug from store_mixed_demo
//
// Bug Scenario:
//   Frame 1 (count=0): Expanded=false, Counter at position Y=2
//   Frame 2 (count=1): Expanded=true,  Counter at position Y=5 (pushed down by expanded section)
//   Bug: Old button at Y=2 remains in buffer
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
// Store and State
// =============================================================================

type AppState struct {
	Count int
}

var appStore = store.NewStore(AppState{Count: 0})

// =============================================================================
// Components (copied from store_mixed_demo)
// =============================================================================

// ExpanderComponent - collapses/shows content based on count
func ExpanderComponent() ui.VNode {
	count := appStore.Get().Count
	expanded := count%2 == 0 // count=0: expanded=true, count=1: expanded=false

	fmt.Printf("  ExpanderComponent: count=%d, expanded=%v\n", count, expanded)

	items := []ui.VNode{
		ui.NewTextBuilder("=== 1. Store 订阅状态 ===").
			FgColor("green").
			Build(),
		ui.Text(fmt.Sprintf("  折叠状态: %s (基于 Count) ", map[bool]string{
			true:  "展开",
			false: "折叠",
		}[expanded])),
	}

	if expanded {
		items = append(items,
			ui.Text("  这是一个基于 Store 的状态"),
			ui.Text("  可以跨组件访问"),
			ui.Text("  数据流：Intent → Reducer → Store → UI"),
		)
	}

	return ui.VStack(items...)
}

func CounterComponent() ui.VNode {
	count := appStore.Get().Count
	fmt.Printf("  CounterComponent: count=%d\n", count)

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

func App() ui.VNode {
	return ui.VStack(
		ExpanderComponent(),
		ui.Text(""),
		CounterComponent(),
	)
}

// =============================================================================
// Main Test
// =============================================================================

func main() {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")

	fmt.Println("╔══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║   Expand/Collapse Bug Reproduction                               ║")
	fmt.Println("║   Reproducing the EXACT bug from store_mixed_demo               ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════╝")

	fwApp := framework.NewApp()
	buf := paint.NewBuffer(80, 25)
	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 80, Height: 25},
		AvailableWidth:  80,
		AvailableHeight: 25,
	}

	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("  Buffer: %dx%d\n", buf.Width, buf.Height)

	// Create single node (simulating real app behavior)
	node := render.NewDeclarativeNodeFromFuncWithFiber(App)
	node.SetApp(fwApp)
	node.SetRenderMode(render.RenderModeFiberFirst)

	frame1Rows := make([]string, 15)
	frame2Rows := make([]string, 15)

	// Frame 1: count=0, expanded=true (Expander shows 4 lines)
	{
		appStore.Set(AppState{Count: 0})
		fmt.Printf("\n%s\n", strings.Repeat("=", 80))
		fmt.Println("FRAME 1: count=0, Expander expanded (4 lines)")
		fmt.Printf("%s\n", strings.Repeat("=", 80))

		// Paint Frame 1 (buffer NOT reset - this is the bug scenario)
		node.Paint(ctx, buf)

		for y := 0; y < 15; y++ {
			frame1Rows[y] = extractRow(buf, y, 80)
		}

		fmt.Println("Frame 1 Buffer (first 15 rows):")
		utils.PrintBuffer(buf, 80, 25)

		fmt.Println("\n  DEBUG: Frame 1 extracted rows:")
		for y, row := range frame1Rows {
			if y < 10 {
				fmt.Printf("    Row %d: %q\n", y, row[:min(50, len(row))])
			}
		}

		// Find Counter button position
		buttonLine := findButtonLine(frame1Rows)
		fmt.Printf("\nCounter button found at line: %d\n", buttonLine)
		if buttonLine < 0 {
			fmt.Println("  DEBUG: Searching for '计数'...")
			for y, row := range frame1Rows {
				if strings.Contains(row, "计数") {
					fmt.Printf("  Found at line %d: %s\n", y, row)
				}
			}
		}
		printRowContent(frame1Rows, buttonLine)
	}

	// Frame 2: count=1, expanded=false (Expander shows 2 lines)
	{
		appStore.Set(AppState{Count: 1})
		fmt.Printf("\n%s\n", strings.Repeat("=", 80))
		fmt.Println("FRAME 2: count=1, Expander collapsed (2 lines)")
		fmt.Printf("%s\n", strings.Repeat("=", 80))

		fmt.Println("Expected: Counter moves UP (since Expander collapsed)")
		fmt.Println("Bug: Old button at old position remains in buffer")

		// Paint Frame 2 on SAME buffer (no reset)
		node.Paint(ctx, buf)

		for y := 0; y < 15; y++ {
			frame2Rows[y] = extractRow(buf, y, 80)
		}

		fmt.Println("\nFrame 2 Buffer (first 15 rows):")
		utils.PrintBuffer(buf, 80, 25)

		// Find new Counter button position
		buttonLine := findButtonLine(frame2Rows)
		fmt.Printf("\nCounter button found at line: %d\n", buttonLine)
		printRowContent(frame2Rows, buttonLine)
	}

	// Analysis
	fmt.Printf("\n%s\n", strings.Repeat("=", 80))
	fmt.Println("ANALYSIS: Check for bug - old button position leftovers")
	fmt.Printf("%s\n\n", strings.Repeat("=", 80))

	bugFound := false

	// Find what should be the NEW button line in Frame 1
	// and check if it has leftover characters
	frame1ButtonLine := findButtonLine(frame1Rows)
	frame2ButtonLine := findButtonLine(frame2Rows)

	if frame1ButtonLine != frame2ButtonLine {
		fmt.Printf("✨ Button position changed: %d → %d\n", frame1ButtonLine, frame2ButtonLine)

		// Check the OLD position (frame1ButtonLine) in Frame 2
		// It should now have ExpanderCollapsed content, NOT button leftovers
		row2AtOldPos := frame2Rows[frame1ButtonLine]
		row1AtOldPos := frame1Rows[frame1ButtonLine]

		fmt.Printf("\nOld button position (line %d):\n", frame1ButtonLine)
		fmt.Printf("  Frame 1: %s\n", row1AtOldPos)
		fmt.Printf("  Frame 2: %s\n", row2AtOldPos)

		// Check for button pattern leftovers: "[", "]", "+", "-", "*"
		hasButtonChars := strings.Contains(row2AtOldPos, "[") ||
			strings.Contains(row2AtOldPos, "]") ||
			strings.Contains(row2AtOldPos, "+") ||
			strings.Contains(row2AtOldPos, "*")

		if hasButtonChars {
			bugFound = true
			fmt.Printf("  🚨 BUG DETECTED: Button characters still present at old position\n")
		} else {
			fmt.Printf("  ✅ Good: No button characters at old position\n")
		}
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 80))
	if bugFound {
		fmt.Printf("❌ BUG CONFIRMED: Buffer wipeout issue present\n")
		fmt.Printf("   When a component moves (due to layout change),\n")
		fmt.Printf("   the old position is not properly cleared.\n")
	} else {
		fmt.Printf("✅ PASS: No bug detected (buffer properly cleared)\n")
	}
	fmt.Printf(strings.Repeat("=", 80))
}

func findButtonLine(rows []string) int {
	for y, row := range rows {
		// Look for "计数:" line (the button is on the SAME line)
		if strings.Contains(row, "计数:") {
			return y
		}
	}
	return -1
}

func printRowContent(rows []string, y int) {
	if y >= 0 && y < len(rows) {
		fmt.Printf("Line %d: %s\n", y, rows[y])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func extractRow(buffer *paint.Buffer, row, width int) string {
	var sb strings.Builder
	for x := 0; x < width; x++ {
		cell := buffer.GetContent(x, row)
		// Skip continuation cells (they're part of wide characters)
		if cell.IsContinuation {
			continue
		}
		if cell.Cluster == "" || cell.Cluster == " " {
			sb.WriteRune(' ')
		} else {
			sb.WriteString(cell.Cluster)
		}
	}
	return sb.String()
}
