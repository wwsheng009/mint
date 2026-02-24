// Fiber-First List Component Demo
// Demonstrates the new List component following the Fiber-first architecture
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
	"github.com/wwsheng009/mint/runtime/style"
	newstack "github.com/wwsheng009/mint/ui/components/stack"
	newtext "github.com/wwsheng009/mint/ui/components/text"
	"github.com/wwsheng009/mint/ui/components/list"
)

// DemoApp creates the demo UI
func DemoApp() rtui.VNode {
	// Sample data
	users := []string{
		"alice     Active   Administrator",
		"bob       Pending  Developer",
		"charlie Active   Designer",
		"diana     Active   Manager",
		"eve       Inactive Tester",
		"frank     Active   Developer",
		"grace     Pending  Designer",
		"henry     Active   Manager",
		"iris      Inactive Developer",
		"jack      Active   Tester",
	}

	files := []string{
		"main.go           245 KB",
		"app.go            156 KB",
		"components.go     412 KB",
		"utils.go           89 KB",
		"config.yaml        32 KB",
		"README.md          15 KB",
	}

	return newstack.New(newstack.Column).
		SetWidth(70).
		SetGap(1).
		SetChildrenList([]rtui.VNode{
			// Title
			sectionTitle("Fiber-First List Component Demo"),
			newtext.New(""),
			newtext.New("General-purpose list with Fiber-first architecture:"),
			newtext.New("  • Pure descriptive VNode"),
			newtext.New("  • Optional header row"),
			newtext.New("  • Row selection with highlighting"),
			newtext.New("  • Scroll support for large lists"),
			newtext.New("  • Keyboard navigation"),
			newtext.New("  • Border and separator options"),
			newtext.New("  • Empty list handling"),
			newtext.New(""),

			// =====================================================
			// Section 1: Basic List
			// =====================================================
			subTitle("1. Basic User List"),
			list.NewBuilder().
				Header("User      Status   Role").
				Rows(users).
				ShowBorder(true).
				ShowSeparator(true).
				ViewportHeight(7).
				Build(),
			newtext.New(""),

			// =====================================================
			// Section 2: List with Selection
			// =====================================================
			subTitle("2. List with Selected Row"),
			list.NewBuilder().
				Header("File             Size").
				Rows(files).
				ShowBorder(true).
				ShowSeparator(true).
				SelectedIndex(2).
				ViewportHeight(6).
				Build(),
			newtext.New(""),

			// =====================================================
			// Section 3: List without Border
			// =====================================================
			subTitle("3. Simple List (no border)"),
			list.NewBuilder().
				Rows(users[:5]).
				ShowBorder(false).
				ViewportHeight(5).
				Build(),
			newtext.New(""),

			// =====================================================
			// Section 4: Empty List
			// =====================================================
			subTitle("4. Empty List"),
			list.NewBuilder().
				Header("Name Value").
				Rows([]string{}).
				EmptyText("(no data available)").
				ShowBorder(true).
				ViewportHeight(5).
				Build(),
			newtext.New(""),

			// =====================================================
			// Section 5: List with Custom Styles
			// =====================================================
			subTitle("5. List with Custom Styles"),
			list.NewBuilder().
				Header("Item Description").
				Rows([]string{
					"Task 1 Complete the first task",
					"Task 2 Review the documentation",
					"Task 3 Test the implementation",
				}).
				HeaderStyle(style.Style{FG: style.Yellow}).
				SelectedStyle(style.Style{BG: style.Green, FG: style.White}).
				ShowBorder(true).
				SelectedIndex(1).
				ViewportHeight(5).
				Build(),
			newtext.New(""),

			// Footer
			highlight("List: General-purpose list, header, selection, scrolling, styles"),
		})
}

// sectionTitle creates a styled section title
func sectionTitle(title string) rtui.VNode {
	return newtext.New(title).Bold(true)
}

// subTitle creates a subtitle
func subTitle(title string) rtui.VNode {
	return newtext.New("  " + title)
}

// highlight creates a highlighted note
func highlight(text string) rtui.VNode {
	return newtext.New("  >>> " + text)
}

func main() {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")
	os.Setenv("MINT_DEBUG_TEST", "true")

	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║   Fiber-First List Component Demo                         ║")
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
	fmt.Println("Rendering List with Fiber-first pipeline...")
	fmt.Printf("%s\n\n", strings.Repeat("=", 70))

	// Render
	node.Paint(ctx, buf)

	// Output result
	printBuffer(buf, 70, 60)

	// Feature summary
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("List Architecture:")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("  ✓ VNode: Pure description (no state, no closures, no paint)")
	fmt.Println("  ✓ Instance: Runtime state management")
	fmt.Println("  ✓ Header: Optional column header with custom style")
	fmt.Println("  ✓ Selection: Visual highlighting of selected row")
	fmt.Println("  ✓ Scroll: Support for large datasets with viewport")
	fmt.Println("  ✓ Navigation: Keyboard navigation (up/down, page, home/end)")
	fmt.Println("  ✓ Border: Optional border with customizable style")
	fmt.Println("  ✓ Separator: Visual separator between header and data")
	fmt.Println("  ✓ Empty: Custom empty list text")
	fmt.Println("  ✓ Builder: Fluent API with Header, Rows, AddRow, MaxRows, etc.")
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
