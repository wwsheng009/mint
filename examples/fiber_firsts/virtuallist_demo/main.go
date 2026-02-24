// Fiber-First VirtualList Component Demo
// Demonstrates the new VirtualList component following the Fiber-first architecture
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
	"github.com/wwsheng009/mint/runtime/style"
	newstack "github.com/wwsheng009/mint/ui/components/stack"
	newtext "github.com/wwsheng009/mint/ui/components/text"
	"github.com/wwsheng009/mint/ui/components/virtuallist"
)

// DemoApp creates the demo UI
func DemoApp() rtui.VNode {
	// Generate sample items for demos
	shortItems := []string{"Item 1", "Item 2", "Item 3", "Item 4", "Item 5"}
	mediumItems := make([]string, 20)
	for i := 0; i < 20; i++ {
		mediumItems[i] = fmt.Sprintf("Data Item %03d - Some description text", i+1)
	}
	longItems := make([]string, 30)
	for i := 0; i < 30; i++ {
		longItems[i] = fmt.Sprintf("File: document-%04d.txt (size: %dKB, modified: 2024)", i+1, (i+1)*5)
	}

	return newstack.New(newstack.Column).
		SetWidth(70).
		SetGap(1).
		SetChildrenList([]rtui.VNode{
			// Title
			sectionTitle("Fiber-First VirtualList Component Demo"),
			newtext.New(""),
			newtext.New("Virtual scrolling list with Fiber-first architecture:"),
			newtext.New("  • Pure descriptive VNode"),
			newtext.New("  • Virtual scrolling for large datasets"),
			newtext.New("  • Scroll offset management"),
			newtext.New("  • Selected item highlighting"),
			newtext.New("  • Customizable item and selected styles"),
			newtext.New("  • Configurable list size and dimensions"),
			newtext.New("  • Action-driven interactions"),
			newtext.New(""),

			// =====================================================
			// Section 1: Basic VirtualList
			// =====================================================
			subTitle("1. Basic List (5 items)"),
			virtuallist.NewBuilder().
				Items(shortItems).
				Width(50).
				Height(7).
				SelectedIndex(1).
				Build(),
			newtext.New(""),

			// =====================================================
			// Section 2: Medium List with Custom Colors
			// =====================================================
			subTitle("2. Medium List with Custom Colors"),
			virtuallist.NewBuilder().
				Items(mediumItems[:10]).
				Width(50).
				Height(7).
				SelectedIndex(3).
				ListStyle(style.Style{FG: style.Cyan}).
				SelectedStyle(style.Style{BG: style.Blue, FG: style.White}).
				Build(),
			newtext.New(""),

			// =====================================================
			// Section 3: Larger List - Virtual Scrolling
			// =====================================================
			subTitle("3. Larger List (20 items, visible=7)"),
			virtuallist.NewBuilder().
				Items(mediumItems).
				ItemCount(20).
				VisibleCount(7).
				Width(50).
				Height(7).
				ScrollOffset(5).
				SelectedIndex(8).
				Build(),
			newtext.New(""),

			// =====================================================
			// Section 4: List with Scroll Offset
			// =====================================================
			subTitle("4. List at Scroll Offset (offset=10)"),
			virtuallist.NewBuilder().
				Items(longItems).
				Width(50).
				Height(7).
				ScrollOffset(10).
				SelectedIndex(13).
				Build(),
			newtext.New(""),

			// =====================================================
			// Section 5: Different Sizes
			// =====================================================
			subTitle("5. Smaller List"),
			virtuallist.NewBuilder().
				Items([]string{"Option A", "Option B", "Option C"}).
				Size(40, 5).
				SelectedIndex(0).
				SelectedStyle(style.Style{FG: style.Yellow, BG: style.Black}).
				Build(),
			newtext.New(""),

			// =====================================================
			// Section 6: List with Different Item Height
			// =====================================================
			subTitle("6. List with Custom Config"),
			virtuallist.NewBuilder().
				Items(mediumItems[:8]).
				Width(50).
				Height(8).
				ItemHeight(1).
				VisibleCount(8).
				ScrollOffset(0).
				SelectedIndex(2).
				AllowScroll(true).
				ListStyle(style.Style{FG: style.Green}).
				SelectedStyle(style.Style{FG: style.White, BG: style.Green}).
				Build(),
			newtext.New(""),

			// Footer
			highlight("VirtualList: Virtual scrolling, item selection, custom styles, action-driven"),
		})
}

// sectionTitle creates a styled section title
func sectionTitle(title string) rtui.VNode {
	return newtext.New(title).
		Foreground(theme.Primary()).
		Bold(true)
}

// subTitle creates a subtitle
func subTitle(title string) rtui.VNode {
	return newtext.New("  " + title).Foreground("white")
}

// highlight creates a highlighted note
func highlight(text string) rtui.VNode {
	return newtext.New("  >>> " + text).Foreground("yellow")
}

func main() {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")
	os.Setenv("MINT_DEBUG_TEST", "true")

	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║   Fiber-First VirtualList Component Demo                   ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")

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
	buf := paint.NewBuffer(70, 60)

	// Create paint context
	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 70, Height: 60},
		AvailableWidth:  70,
		AvailableHeight: 60,
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 70))
	fmt.Println("Rendering VirtualList with Fiber-first pipeline...")
	fmt.Printf("%s\n\n", strings.Repeat("=", 70))

	// Render
	node.Paint(ctx, buf)

	// Output result
	printBuffer(buf, 70, 60)

	// Feature summary
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("VirtualList Architecture:")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("  ✓ VNode: Pure description (no state, no closures, no paint)")
	fmt.Println("  ✓ Instance: Runtime state management")
	fmt.Println("  ✓ Virtual Scrolling: Only renders visible items")
	fmt.Println("  ✓ Scroll Offset: Efficient scroll position tracking")
	fmt.Println("  ✓ Selection: Visual highlighting of selected item")
	fmt.Println("  ✓ Styles: Separate list and selected styles")
	fmt.Println("  ✓ Actions: scroll, navigate_up/down, page_up/down, select")
	fmt.Println("  ✓ Builder: Fluent API with Items, Size, SelectedIndex, etc.")
	fmt.Println("  ✓ Performance: Efficient for large datasets (virtualization)")
	fmt.Println(strings.Repeat("=", 70))
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
		}
	}
	fmt.Printf("└%s┘\n", strings.Repeat("─", width))
}
