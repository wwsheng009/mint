// Fiber-First Table Component Demo
// Demonstrates the new Table component following the Fiber-first architecture
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/wwsheng009/mint/examples/utils"
	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/internal/render"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	newstack "github.com/wwsheng009/mint/ui/components/stack"
	newtext "github.com/wwsheng009/mint/ui/components/text"
	"github.com/wwsheng009/mint/ui/components/table"
)

// DemoApp creates the demo UI
func DemoApp() rtui.VNode {
	return newstack.New(newstack.Column).
		SetWidth(70).
		SetGap(1).
		SetChildrenList([]rtui.VNode{
			// Title
			sectionTitle("Fiber-First Table Component Demo"),
			newtext.New(""),
			newtext.New("Table data display with Fiber-first architecture:"),
			newtext.New("  • Pure descriptive VNode"),
			newtext.New("  • Column and row definitions"),
			newtext.New("  • Customizable header and table styles"),
			newtext.New("  • Adjustable gap between header and data"),
			newtext.New("  • Automatic separator generation"),
			newtext.New(""),

			// =====================================================
			// Section 1: Basic Table
			// =====================================================
			subTitle("1. Basic User Table"),
			table.NewBuilder().
				Columns([]table.TableColumn{
					{Title: "ID", Width: 5},
					{Title: "Name", Width: 15},
					{Title: "Email", Width: 20},
				}).
				AddRow("1", "Alice Smith", "alice@example.com").
				AddRow("2", "Bob Johnson", "bob@example.com").
				AddRow("3", "Carol Davis", "carol@example.com").
				Build(),
			newtext.New(""),

			// =====================================================
			// Section 2: Auto-width Columns
			// =====================================================
			subTitle("2. Auto-width Columns (Width = 0)"),
			table.NewBuilder().
				Columns([]table.TableColumn{
					{Title: "Status"},
					{Title: "Description"},
					{Title: "Priority"},
				}).
				AddRow("Active", "Task in progress", "High").
				AddRow("Pending", "Waiting for approval", "Medium").
				AddRow("Done", "Completed successfully", "Low").
				Build(),
			newtext.New(""),

			// =====================================================
			// Section 3: Table with Custom Gap
			// =====================================================
			subTitle("3. Table with Custom Gap (gap=2)"),
			table.NewBuilder().
				Columns([]table.TableColumn{
					{Title: "Role"},
					{Title: "User"},
				}).
				AddRow("Admin", "root").
				AddRow("User", "alice").
				AddRow("Guest", "bob").
				Gap(2).
				Build(),
			newtext.New(""),

			// =====================================================
			// Section 4: Mixed Width Table
			// =====================================================
			subTitle("4. Mixed Width Columns"),
			table.NewBuilder().
				Columns([]table.TableColumn{
					{Title: "Code", Width: 8},
					{Title: "Product", Width: 25},
					{Title: "Stock"},
					{Title: "Price"},
				}).
				AddRow("P101", "Laptop 15\"", "50", "$1299").
				AddRow("P102", "Wireless Mouse", "120", "$29").
				AddRow("P103", "USB-C Hub", "30", "$45").
				Build(),
			newtext.New(""),

			// Footer
			highlight("Table: Data display, columns, rows, custom styles, configurable gap"),
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
	fmt.Println("║   Fiber-First Table Component Demo                        ║")
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
	buf := paint.NewBuffer(70, 35)

	// Create paint context
	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 70, Height: 35},
		AvailableWidth:  70,
		AvailableHeight: 35,
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 70))
	fmt.Println("Rendering Table with Fiber-first pipeline...")
	fmt.Printf("%s\n\n", strings.Repeat("=", 70))

	// Render
	node.Paint(ctx, buf)

	// Output result
	utils.PrintBuffer(buf, 70, 35)

	// Feature summary
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("Table Architecture:")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("  ✓ VNode: Pure description (no state, no closures, no paint)")
	fmt.Println("  ✓ Instance: Runtime state management")
	fmt.Println("  ✓ Columns: Configurable column widths (0 = auto)")
	fmt.Println("  ✓ Rows: Simple string-based data structure")
	fmt.Println("  ✓ Styles: Separate header and table styles")
	fmt.Println("  ✓ Gap: Configurable gap between header and data")
	fmt.Println("  ✓ Auto-Separator: Automatic separator generation")
	fmt.Println("  ✓ Display-Only: Clean, focused on data presentation")
	fmt.Println("  ✓ Builder: Fluent API with Columns, Rows, AddRow, Gap")
	fmt.Println(strings.Repeat("=", 70))
}
