// Fiber-first Grid Component Demo
// Demonstrates the Grid component following the Fiber-first architecture
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
	"github.com/wwsheng009/mint/ui/components/grid"
	newstack "github.com/wwsheng009/mint/ui/components/stack"
	newtext "github.com/wwsheng009/mint/ui/components/text"
)

// DemoApp renders Grid layouts using the Fiber-first pipeline
func DemoApp() rtui.VNode {
	return newstack.NewVStack().
		SetGap(0).
		SetChildrenList([]rtui.VNode{
			// =====================================================
			// Section 1: Basic Grid
			// =====================================================
			sectionTitle("═══ 1. Basic Grid ═══"),

			// 1.1 Two Column Grid
			subTitle("1.1 Two Column Grid (auto-position)"),
			grid.TwoColumnGrid(
				newtext.New("Name:"),
				newtext.New("John Doe"),
				newtext.New("Age:"),
				newtext.New("30"),
				newtext.New("City:"),
				newtext.New("New York"),
			),

			// 1.2 Three Column Grid
			subTitle("1.2 Three Column Grid"),
			grid.ThreeColumnGrid(
				newtext.New("[A]"),
				newtext.New("[B]"),
				newtext.New("[C]"),
				newtext.New("[D]"),
				newtext.New("[E]"),
				newtext.New("[F]"),
			),

			// =====================================================
			// Section 2: Fixed Columns
			// =====================================================
			sectionTitle("═══ 2. Fixed Column Widths ═══"),

			// 2.1 Fixed Width Columns
			subTitle("2.1 Fixed(10,20,15)"),
			grid.FixedGrid([]int{10, 20, 15},
				newtext.New("ID"),
				newtext.New("Name"),
				newtext.New("Score"),
				newtext.New("001"),
				newtext.New("Alice"),
				newtext.New("95"),
				newtext.New("002"),
				newtext.New("Bob"),
				newtext.New("87"),
			),

			// =====================================================
			// Section 3: Flex Columns
			// =====================================================
			sectionTitle("═══ 3. Flex Column Distribution ═══"),

			// 3.1 Equal Flex
			subTitle("3.1 Equal Flex (1:1:1)"),
			grid.New().
				SetColumns(grid.Flex{Factor: 1}, grid.Flex{Factor: 1}, grid.Flex{Factor: 1}).
				SetWidth(45).
				SetChildrenAuto([]rtui.VNode{
					newtext.New("┌──────────┐"),
					newtext.New("┌──────────┐"),
					newtext.New("┌──────────┐"),
					newtext.New("│ Col 1    │"),
					newtext.New("│ Col 2    │"),
					newtext.New("│ Col 3    │"),
					newtext.New("└──────────┘"),
					newtext.New("└──────────┘"),
					newtext.New("└──────────┘"),
				}),

			// 3.2 Unequal Flex
			subTitle("3.2 Unequal Flex (1:2:1)"),
			grid.New().
				SetColumns(grid.Flex{Factor: 1}, grid.Flex{Factor: 2}, grid.Flex{Factor: 1}).
				SetWidth(45).
				SetChildrenAuto([]rtui.VNode{
					newtext.New("[Small]"),
					newtext.New("[=======Wide=======]"),
					newtext.New("[Small]"),
				}),

			// =====================================================
			// Section 4: Mixed Dimensions
			// =====================================================
			sectionTitle("═══ 4. Mixed Dimensions ═══"),

			// 4.1 Fixed + Flex + Auto
			subTitle("4.1 Fixed(15) + Flex(1) + Auto"),
			grid.New().
				SetColumns(grid.Fixed(15), grid.Flex{Factor: 1}, grid.Auto{}).
				SetWidth(50).
				SetChildrenAuto([]rtui.VNode{
					newtext.New("Label:"),
					newtext.New("[Flexible Content Area]"),
					newtext.New("[X]"),
					newtext.New("Status:"),
					newtext.New("[====================]"),
					newtext.New("[?]"),
				}),

			// =====================================================
			// Section 5: Grid with Gap
			// =====================================================
			sectionTitle("═══ 5. Grid with Gap ═══"),

			// 5.1 Column Gap
			subTitle("5.1 Column Gap=2"),
			grid.NewBuilder().
				Columns(grid.Fixed(10), grid.Fixed(10), grid.Fixed(10)).
				Gap(2, 0).
				Children([]rtui.VNode{
					newtext.New("[Cell A]"),
					newtext.New("[Cell B]"),
					newtext.New("[Cell C]"),
				}).
				Build(),

			// 5.2 Row Gap
			subTitle("5.2 Row Gap=1"),
			grid.NewBuilder().
				Columns(grid.Fixed(10), grid.Fixed(10)).
				Rows(grid.Auto{}, grid.Auto{}).
				Gap(1, 1).
				Children([]rtui.VNode{
					newtext.New("[R1C1]"),
					newtext.New("[R1C2]"),
					newtext.New("[R2C1]"),
					newtext.New("[R2C2]"),
				}).
				Build(),

			// =====================================================
			// Section 6: Explicit Cell Position
			// =====================================================
			sectionTitle("═══ 6. Explicit Cell Position ═══"),

			// 6.1 Positioned Cells
			subTitle("6.1 Explicit Cell Position"),
			grid.New().
				SetColumns(grid.Fixed(10), grid.Fixed(10), grid.Fixed(10)).
				SetRows(grid.Fixed(1), grid.Fixed(1), grid.Fixed(1)).
				AddCell(0, 0, newtext.New("[0,0]")).
				AddCell(0, 2, newtext.New("[0,2]")).
				AddCell(2, 0, newtext.New("[2,0]")).
				AddCell(2, 2, newtext.New("[2,2]")),

			// 6.2 Cell with Span
			subTitle("6.2 Cell Span (RowSpan=2)"),
			grid.New().
				SetColumns(grid.Fixed(15), grid.Fixed(15)).
				SetRows(grid.Fixed(1), grid.Fixed(1)).
				AddCellSpan(0, 0, 2, 1, newtext.New("[Span 2 Rows]")).
				AddCell(0, 1, newtext.New("[R0C1]")).
				AddCell(1, 1, newtext.New("[R1C1]")),

			// =====================================================
			// Section 7: Simple Table
			// =====================================================
			sectionTitle("═══ 7. Simple Table Layout ═══"),

			// 7.1 Data Table
			subTitle("7.1 Data Table"),
			grid.New().
				SetColumns(
					grid.Fixed(5),
					grid.Fixed(20),
					grid.Fixed(10),
					grid.Fixed(8),
				).
				SetGap(1, 0).
				SetChildrenAuto([]rtui.VNode{
					// Header
					newtext.New("ID").Foreground(theme.Primary()),
					newtext.New("Name").Foreground(theme.Primary()),
					newtext.New("Dept").Foreground(theme.Primary()),
					newtext.New("Score").Foreground(theme.Primary()),
					// Row 1
					newtext.New("001"),
					newtext.New("Alice Smith"),
					newtext.New("Engineering"),
					newtext.New("95"),
					// Row 2
					newtext.New("002"),
					newtext.New("Bob Jones"),
					newtext.New("Marketing"),
					newtext.New("87"),
					// Row 3
					newtext.New("003"),
					newtext.New("Carol White"),
					newtext.New("Sales"),
					newtext.New("92"),
				}),

			// =====================================================
			// Section 8: Dashboard Layout
			// =====================================================
			sectionTitle("═══ 8. Dashboard Layout ═══"),

			// 8.1 Dashboard Grid
			subTitle("8.1 Dashboard (2x2)"),
			grid.New().
				SetColumns(grid.Flex{Factor: 1}, grid.Flex{Factor: 1}).
				SetRows(grid.Fixed(3), grid.Fixed(3)).
				SetGap(2, 1).
				SetWidth(50).
				SetChildrenAuto([]rtui.VNode{
					boxPanel("Stats", "CPU: 45%"),
					boxPanel("Stats", "MEM: 60%"),
					boxPanel("Logs", "[INFO] OK"),
					boxPanel("Status", "Running"),
				}),
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
	return newtext.New("  " + title)
}

// boxPanel creates a simple panel
func boxPanel(title, content string) rtui.VNode {
	return newstack.NewVStack().
		SetChildrenList([]rtui.VNode{
			newtext.New("┌─ " + title + " ─"),
			newtext.New("│ " + content),
			newtext.New("└─"),
		})
}

func main() {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")

	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║   Fiber-First Grid Rendering Demo                          ║")
	fmt.Println("║   (Grid layout component)                                  ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")

	fwApp := framework.NewApp()
	node := render.NewDeclarativeNodeFromFuncWithFiber(DemoApp, fwApp)
	node.SetRenderMode(render.RenderModeFiberFirst)

	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("  Render Mode: %v\n", node.GetRenderMode())

	buf := paint.NewBuffer(60, 65)
	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 60, Height: 65},
		AvailableWidth:  60,
		AvailableHeight: 65,
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 60))
	fmt.Println("Rendering Grid layouts with Fiber-first pipeline...")
	fmt.Printf("%s\n\n", strings.Repeat("=", 60))

	node.Paint(ctx, buf)
	printBuffer(buf, 60, 65)

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Grid Component Features:")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  - Fixed: Fixed column/row size")
	fmt.Println("  - Flex: Flexible size with factor")
	fmt.Println("  - Auto: Size to content")
	fmt.Println("  - Gap: Spacing between cells")
	fmt.Println("  - Padding: Inner spacing")
	fmt.Println("  - Cell Position: Explicit row/col placement")
	fmt.Println("  - Cell Span: RowSpan/ColSpan for merged cells")
	fmt.Println("")
	fmt.Println("Convenience Functions:")
	fmt.Println("  - TwoColumnGrid(): 2 equal columns")
	fmt.Println("  - ThreeColumnGrid(): 3 equal columns")
	fmt.Println("  - SimpleGrid(n): n equal columns")
	fmt.Println("  - FixedGrid(widths[]): fixed width columns")
	fmt.Println(strings.Repeat("=", 60))
}

func printBuffer(buf *paint.Buffer, width, height int) {
	fmt.Printf("┌%s┐\n", strings.Repeat("─", width))
	for y := 0; y < height; y++ {
		var line strings.Builder
		for x := 0; x < width; x++ {
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
		trimmed := strings.TrimRight(line.String(), " ")
		if trimmed != "" {
			fmt.Printf("|%-*s|\n", width, trimmed)
		}
	}
	fmt.Printf("└%s┘\n", strings.Repeat("─", width))
}
