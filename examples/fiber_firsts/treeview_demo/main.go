// Fiber-First TreeView Component Demo
// Demonstrates the new TreeView component following the Fiber-first architecture
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
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
	newtext "github.com/wwsheng009/mint/ui/components/text"
	"github.com/wwsheng009/mint/ui/components/treeview"
)

// DemoApp creates the demo UI
func DemoApp() rtui.VNode {
	// Create sample tree nodes
	folderTree := []string{
		"src/",
		"    main.go",
		"    app.go",
		"    components/",
		"        button.go",
		"        text.go",
		"        panel.go",
		"    tests/",
		"        button_test.go",
		"        text_test.go",
		"README.md",
		"package.json",
		".gitignore",
	}

	return ui.NewVStack().
		SetWidth(70).
		SetGap(1).
		SetChildrenList([]rtui.VNode{
			// Title
			sectionTitle("Fiber-First TreeView Component Demo"),
			newtext.New(""),
			newtext.New("Hierarchical tree structure with Fiber-first architecture:"),
			newtext.New("  • Pure descriptive VNode"),
			newtext.New("  • Expand/collapse nodes"),
			newtext.New("  • Node selection with highlighting"),
			newtext.New("  • Virtual scrolling for large trees"),
			newtext.New("  • Keyboard navigation"),
			newtext.New("  • Customizable styles"),
			newtext.New(""),

			// =====================================================
			// Section 1: Basic Tree
			// =====================================================
			subTitle("1. Basic Project Tree"),
			treeview.NewBuilder().
				FromLines(folderTree).
				ShowIcons(true).
				ExpandLevel(1).
				ViewportHeight(7).
				Build(),
			newtext.New(""),

			// =====================================================
			// Section 2: Tree with Default Expanded
			// =====================================================
			subTitle("2. Fully Expanded Tree"),
			treeview.TreeView().
				FromLines(folderTree).
				ShowIcons(true).
				ExpandLevel(3). // Expand up to depth 3
				ViewportHeight(8).
				Build(),
			newtext.New(""),

			// =====================================================
			// Section 3: Compact Tree without Icons
			// =====================================================
			subTitle("3. Compact Tree (no icons)"),
			treeview.NewBuilder().
				FromLines(folderTree).
				ShowIcons(false).
				ExpandLevel(2).
				ViewportHeight(7).
				Build(),
			newtext.New(""),

			// =====================================================
			// Section 4: Tree with Selection
			// =====================================================
			subTitle("4. Tree with Selection Highlighted"),
			treeview.NewBuilder().
				FromLines(folderTree).
				ShowIcons(true).
				ExpandLevel(2).
				ViewportHeight(7).
				SelectedIndex(3).
				SelectedStyle(style.Style{FG: style.White, BG: style.Blue}).
				Build(),
			newtext.New(""),

			// =====================================================
			// Section 5: Simple List Tree
			// =====================================================
			subTitle("5. Simple Hierarchical List"),
			treeview.NewBuilder().
				FromLines([]string{
					"Root Level",
					"    Child 1",
					"    Child 2",
					"        Grandchild",
					"    Child 3",
				}).
				ViewportHeight(6).
				Build(),
			newtext.New(""),

			// Footer
			highlight("TreeView: Hierarchical data, expand/collapse, selection, navigation"),
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
	fmt.Println("║   Fiber-First TreeView Component Demo                    ║")
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
	buf := paint.NewBuffer(70, 65)

	// Create paint context
	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 70, Height: 65},
		AvailableWidth:  70,
		AvailableHeight: 65,
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 70))
	fmt.Println("Rendering TreeView with Fiber-first pipeline...")
	fmt.Printf("%s\n\n", strings.Repeat("=", 70))

	// Render
	node.Paint(ctx, buf)

	// Output result
	utils.PrintBuffer(buf, 70, 65)

	// Feature summary
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("TreeView Architecture:")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("  ✓ VNode: Pure description (no state, no closures, no paint)")
	fmt.Println("  ✓ Instance: Runtime state management")
	fmt.Println("  ✓ Expand State: Track which nodes are expanded/collapsed")
	fmt.Println("  ✓ Virtual Scrolling: Only render visible nodes")
	fmt.Println("  ✓ Selection: Visual highlighting of selected node")
	fmt.Println("  ✓ Icons: Visual node type indicators (📁, 📂, 📄)")
	fmt.Println("  ✓ Actions: navigate_up/down, toggle_expand, select")
	fmt.Println("  ✓ Builder: Fluent API with Nodes, FromLines, ExpandLevel, etc.")
	fmt.Println("  ✓ Parsing: Auto-detect node structure from indentation")
	fmt.Println(strings.Repeat("=", 70))
}
