// Fiber-first Wrap Component Demo
// Demonstrates the Wrap component following the Fiber-first architecture
// Wrap is a layout container that wraps children to new rows when they exceed container width
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/internal/render"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	newborder "github.com/wwsheng009/mint/ui/components/border"
	newbutton "github.com/wwsheng009/mint/ui/components/button"
	newstack "github.com/wwsheng009/mint/ui/components/stack"
	newtext "github.com/wwsheng009/mint/ui/components/text"
	"github.com/wwsheng009/mint/ui/components/wrap"
	"github.com/wwsheng009/mint/examples/utils"
)

// DemoApp renders Wrap layouts using the Fiber-first pipeline
func DemoApp() rtui.VNode {
	return newstack.NewVStack().
		SetGap(0).
		SetChildrenList([]rtui.VNode{
			// =====================================================
			// Section 1: Basic Wrap
			// =====================================================
			sectionTitle("═══ 1. Basic Wrap ═══"),

			// 1.1 Single row (fits in width)
			subTitle("1.1 Single Row (fits in width=40)"),
			wrapWithBorder(wrap.New().
				SetWidth(40).
				SetGap(1).
				SetChildrenList([]rtui.VNode{
					newtext.New("[A]"),
					newtext.New("[B]"),
					newtext.New("[C]"),
				})),

			// 1.2 Auto wrap (multiple rows)
			subTitle("1.2 Auto Wrap (children exceed width=20)"),
			wrapWithBorder(wrap.New().
				SetWidth(20).
				SetGap(1).
				SetChildrenList([]rtui.VNode{
					newtext.New("[Button1]"),
					newtext.New("[Button2]"),
					newtext.New("[Button3]"),
					newtext.New("[Button4]"),
					newtext.New("[Button5]"),
				})),

			// =====================================================
			// Section 2: Gap Settings
			// =====================================================
			sectionTitle("═══ 2. Gap Settings ═══"),

			// 2.1 Gap=0
			subTitle("2.1 Gap=0 (no spacing, width=30)"),
			wrapWithBorder(wrap.New().
				SetWidth(30).
				SetGap(0).
				SetChildrenList([]rtui.VNode{
					newtext.New("[A]"),
					newtext.New("[B]"),
					newtext.New("[C]"),
					newtext.New("[D]"),
				})),

			// 2.2 Gap=2
			subTitle("2.2 Gap=2 (width=30)"),
			wrapWithBorder(wrap.New().
				SetWidth(30).
				SetGap(2).
				SetChildrenList([]rtui.VNode{
					newtext.New("[A]"),
					newtext.New("[B]"),
					newtext.New("[C]"),
					newtext.New("[D]"),
				})),

			// 2.3 Gap=3
			subTitle("2.3 Gap=3 (width=30)"),
			wrapWithBorder(wrap.New().
				SetWidth(30).
				SetGap(3).
				SetChildrenList([]rtui.VNode{
					newtext.New("[A]"),
					newtext.New("[B]"),
					newtext.New("[C]"),
					newtext.New("[D]"),
				})),

			// =====================================================
			// Section 3: Row Gap
			// =====================================================
			sectionTitle("═══ 3. Row Gap ═══"),

			// 3.1 RowGap=0 (uses gap)
			subTitle("3.1 RowGap=0 (uses gap=1, width=25)"),
			wrapWithBorder(wrap.New().
				SetWidth(25).
				SetGap(1).
				SetRowGap(0).
				SetChildrenList([]rtui.VNode{
					newtext.New("[Item1]"),
					newtext.New("[Item2]"),
					newtext.New("[Item3]"),
					newtext.New("[Item4]"),
				})),

			// 3.2 RowGap=2
			subTitle("3.2 RowGap=2 (separate row spacing, width=25)"),
			wrapWithBorder(wrap.New().
				SetWidth(25).
				SetGap(1).
				SetRowGap(2).
				SetChildrenList([]rtui.VNode{
					newtext.New("[Item1]"),
					newtext.New("[Item2]"),
					newtext.New("[Item3]"),
					newtext.New("[Item4]"),
				})),

			// 3.3 RowGap=3
			subTitle("3.3 RowGap=3 (larger row spacing, width=25)"),
			wrapWithBorder(wrap.New().
				SetWidth(25).
				SetGap(1).
				SetRowGap(3).
				SetChildrenList([]rtui.VNode{
					newtext.New("[Item1]"),
					newtext.New("[Item2]"),
					newtext.New("[Item3]"),
					newtext.New("[Item4]"),
				})),

			// =====================================================
			// Section 4: Alignment - KEY SECTION TO SHOW CONTAINER WIDTH
			// =====================================================
			sectionTitle("═══ 4. Alignment (width=40 shown by border) ═══"),

			// 4.1 AlignStart
			subTitle("4.1 AlignStart (left)"),
			wrapWithBorder(wrap.New().
				SetWidth(40).
				SetGap(1).
				SetAlign(wrap.AlignStart).
				SetChildrenList([]rtui.VNode{
					newtext.New("[A]"),
					newtext.New("[B]"),
					newtext.New("[C]"),
				})),

			// 4.2 AlignCenter
			subTitle("4.2 AlignCenter (centered in 40-char container)"),
			wrapWithBorder(wrap.New().
				SetWidth(40).
				SetGap(1).
				SetAlign(wrap.AlignCenter).
				SetChildrenList([]rtui.VNode{
					newtext.New("[A]"),
					newtext.New("[B]"),
					newtext.New("[C]"),
				})),

			// 4.3 AlignEnd
			subTitle("4.3 AlignEnd (right in 40-char container)"),
			wrapWithBorder(wrap.New().
				SetWidth(40).
				SetGap(1).
				SetAlign(wrap.AlignEnd).
				SetChildrenList([]rtui.VNode{
					newtext.New("[A]"),
					newtext.New("[B]"),
					newtext.New("[C]"),
				})),

			// =====================================================
			// Section 5: Buttons Wrap
			// =====================================================
			sectionTitle("═══ 5. Buttons Wrap ═══"),

			// 5.1 Button Toolbar
			subTitle("5.1 Button Toolbar (width=50, wraps automatically)"),
			wrapWithBorder(wrap.New().
				SetWidth(50).
				SetGap(2).
				SetChildrenList([]rtui.VNode{
					newbutton.New("New"),
					newbutton.New("Open"),
					newbutton.New("Save"),
					newbutton.New("Cut"),
					newbutton.New("Copy"),
					newbutton.New("Paste"),
					newbutton.New("Undo"),
					newbutton.New("Redo"),
				})),

			// 5.2 Primary/Secondary Buttons
			subTitle("5.2 Mixed Button Types (width=45)"),
			wrapWithBorder(wrap.New().
				SetWidth(45).
				SetGap(2).
				SetChildrenList([]rtui.VNode{
					newbutton.New("OK").SetVariant(newbutton.VariantPrimary),
					newbutton.New("Cancel"),
					newbutton.New("Apply").SetVariant(newbutton.VariantSuccess),
					newbutton.New("Reset"),
					newbutton.New("Delete").SetVariant(newbutton.VariantDanger),
				})),

			// =====================================================
			// Section 6: Padding
			// =====================================================
			sectionTitle("═══ 6. Padding ═══"),

			// 6.1 Padding All Sides
			subTitle("6.1 Padding(1,2,1,2) - note inner spacing"),
			wrapWithBorder(wrap.New().
				SetWidth(30).
				SetGap(1).
				SetPadding(1, 2, 1, 2).
				SetChildrenList([]rtui.VNode{
					newtext.New("[A]"),
					newtext.New("[B]"),
					newtext.New("[C]"),
				})),

			// 6.2 Different Padding
			subTitle("6.2 Padding(2,4,2,4) - more spacing"),
			wrapWithBorder(wrap.New().
				SetWidth(35).
				SetGap(1).
				SetPadding(2, 4, 2, 4).
				SetChildrenList([]rtui.VNode{
					newtext.New("[Item1]"),
					newtext.New("[Item2]"),
					newtext.New("[Item3]"),
				})),

			// =====================================================
			// Section 7: Variable Width Items
			// =====================================================
			sectionTitle("═══ 7. Variable Width Items ═══"),

			// 7.1 Different lengths
			subTitle("7.1 Variable Length Items (width=40)"),
			wrapWithBorder(wrap.New().
				SetWidth(40).
				SetGap(1).
				SetChildrenList([]rtui.VNode{
					newtext.New("[Short]"),
					newtext.New("[Medium Length]"),
					newtext.New("[X]"),
					newtext.New("[Very Long Item Name]"),
					newtext.New("[Med]"),
					newtext.New("[End]"),
				})),

			// =====================================================
			// Section 8: Many Items
			// =====================================================
			sectionTitle("═══ 8. Many Items (Auto Wrap) ═══"),

			// 8.1 Many small items
			subTitle("8.1 15 Items (width=50)"),
			wrapWithBorder(wrap.New().
				SetWidth(50).
				SetGap(1).
				SetChildrenList([]rtui.VNode{
					newtext.New("[01]"),
					newtext.New("[02]"),
					newtext.New("[03]"),
					newtext.New("[04]"),
					newtext.New("[05]"),
					newtext.New("[06]"),
					newtext.New("[07]"),
					newtext.New("[08]"),
					newtext.New("[09]"),
					newtext.New("[10]"),
					newtext.New("[11]"),
					newtext.New("[12]"),
					newtext.New("[13]"),
					newtext.New("[14]"),
					newtext.New("[15]"),
				})),

			// =====================================================
			// Section 9: Narrow Container - KEY SECTION
			// =====================================================
			sectionTitle("═══ 9. Narrow Container ═══"),

			// 9.1 Very narrow
			subTitle("9.1 Width=15 (very narrow, forces wrap)"),
			wrapWithBorder(wrap.New().
				SetWidth(15).
				SetGap(0).
				SetChildrenList([]rtui.VNode{
					newtext.New("[Btn1]"),
					newtext.New("[Btn2]"),
					newtext.New("[Btn3]"),
					newtext.New("[Btn4]"),
				})),

			// =====================================================
			// Section 10: Builder API
			// =====================================================
			sectionTitle("═══ 10. Builder API ═══"),

			// 10.1 Using Builder (returns rtui.VNode interface, can't use wrapWithBorder)
			subTitle("10.1 Builder Pattern (width=40, center)"),
			wrap.NewBuilder().
				Width(40).
				Gap(2).
				Align(wrap.AlignCenter).
				Children(
					newtext.New("[Built]"),
					newtext.New("[With]"),
					newtext.New("[Builder]"),
					newtext.New("[API]"),
				).
				Build(),

			// =====================================================
			// Section 11: Convenience Functions
			// =====================================================
			sectionTitle("═══ 11. Convenience Functions ═══"),

			// 11.1 WrapWithWidth (returns rtui.VNode interface)
			subTitle("11.1 WrapWithWidth(35, ...)"),
			wrap.WrapWithWidth(35,
				newtext.New("[A]"),
				newtext.New("[B]"),
				newtext.New("[C]"),
				newtext.New("[D]"),
				newtext.New("[E]"),
			),

			// 11.2 WrapWithGap
			subTitle("11.2 WrapWithGap(3, ...) - default width=80"),
			wrap.New().
				SetGap(3).
				SetChildrenList([]rtui.VNode{
					newtext.New("[One]"),
					newtext.New("[Two]"),
					newtext.New("[Three]"),
					newtext.New("[Four]"),
				}),

			// =====================================================
			// Section 12: Real-world Example
			// =====================================================
			sectionTitle("═══ 12. Real-world: Tag Cloud ═══"),

			// 12.1 Tag Cloud
			subTitle("12.1 Tag Cloud (width=55)"),
			wrapWithBorder(wrap.New().
				SetWidth(55).
				SetGap(2).
				SetRowGap(1).
				SetChildrenList([]rtui.VNode{
					newtext.New("#golang"),
					newtext.New("#fiber"),
					newtext.New("#tui"),
					newtext.New("#terminal"),
					newtext.New("#ui"),
					newtext.New("#layout"),
					newtext.New("#wrap"),
					newtext.New("#flex"),
					newtext.New("#component"),
					newtext.New("#render"),
				})),
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

// wrapWithBorder wraps a Wrap component in a border to visualize container width
// The border shows the actual Wrap container boundary
// Border auto-measures child dimensions when width/height not explicitly set
func wrapWithBorder(w *wrap.VNode) rtui.VNode {
	// Border now auto-measures child - no need for explicit width/height
	return newborder.New().
		SetBorderStyle(newborder.BorderSingle).
		SetBorderColor("cyan").
		SetChild(w)
}

func main() {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")

	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║   Fiber-First Wrap Rendering Demo                          ║")
	fmt.Println("║   (Wrap layout component - flex-wrap: wrap)                ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")

	fwApp := framework.NewApp()
	node := render.NewDeclarativeNodeFromFuncWithFiber(DemoApp, fwApp)
	node.SetRenderMode(render.RenderModeFiberFirst)

	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("  Render Mode: %v\n", node.GetRenderMode())

	buf := paint.NewBuffer(60, 130)
	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 60, Height: 130},
		AvailableWidth:  60,
		AvailableHeight: 130,
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 60))
	fmt.Println("Rendering Wrap layouts with Fiber-first pipeline...")
	fmt.Printf("%s\n\n", strings.Repeat("=", 60))

	node.Paint(ctx, buf)
	utils.PrintBuffer(buf, 60, 130)

	// Print layout analysis
	printLayoutAnalysis()

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Wrap Component Features:")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  - Gap: Spacing between items in the same row")
	fmt.Println("  - RowGap: Spacing between rows (0 = use gap)")
	fmt.Println("  - Align: Start | Center | End (per-row alignment)")
	fmt.Println("  - Width: Container width for wrap calculation")
	fmt.Println("  - Padding: Inner spacing [top, right, bottom, left]")
	fmt.Println("  - FillWidth: Stretch rows to fill container")
	fmt.Println("  - FillHeight: Stretch wrap to fill parent height")
	fmt.Println("")
	fmt.Println("Convenience Functions:")
	fmt.Println("  - Wrap(children...): Create wrap with children")
	fmt.Println("  - WrapWithWidth(w, children...): Create with width")
	fmt.Println("  - WrapWithGap(gap, children...): Create with gap")
	fmt.Println("  - WrapConfig(w, gap, align, children...): Full config")
	fmt.Println("  - W(): Builder pattern")
	fmt.Println(strings.Repeat("=", 60))
}

func printLayoutAnalysis() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Layout Analysis (Measure Phase):")
	fmt.Println(strings.Repeat("=", 60))

	// Test various wrap configurations and print layout info
	testCases := []struct {
		name     string
		width    int
		gap      int
		rowGap   int
		children []rtui.VNode
	}{
		{
			name:   "Section 1.2: Auto Wrap",
			width:  20,
			gap:    1,
			rowGap: 0,
			children: []rtui.VNode{
				newtext.New("[Button1]"),
				newtext.New("[Button2]"),
				newtext.New("[Button3]"),
				newtext.New("[Button4]"),
				newtext.New("[Button5]"),
			},
		},
		{
			name:   "Section 3.2: RowGap=2",
			width:  25,
			gap:    1,
			rowGap: 2,
			children: []rtui.VNode{
				newtext.New("[Item1]"),
				newtext.New("[Item2]"),
				newtext.New("[Item3]"),
				newtext.New("[Item4]"),
			},
		},
		{
			name:   "Section 8.1: 15 Items",
			width:  50,
			gap:    1,
			rowGap: 0,
			children: []rtui.VNode{
				newtext.New("[01]"),
				newtext.New("[02]"),
				newtext.New("[03]"),
				newtext.New("[04]"),
				newtext.New("[05]"),
				newtext.New("[06]"),
				newtext.New("[07]"),
				newtext.New("[08]"),
				newtext.New("[09]"),
				newtext.New("[10]"),
				newtext.New("[11]"),
				newtext.New("[12]"),
				newtext.New("[13]"),
				newtext.New("[14]"),
				newtext.New("[15]"),
			},
		},
	}

	for _, tc := range testCases {
		fmt.Printf("\n  %s:\n", tc.name)
		fmt.Printf("    Container Width: %d, Gap: %d, RowGap: %d\n", tc.width, tc.gap, tc.rowGap)

		// Create instance and measure
		inst := wrap.NewInstance(rtui.Props{
			"width":    tc.width,
			"gap":      tc.gap,
			"rowGap":   tc.rowGap,
			"children": tc.children,
		})

		size := inst.Measure(layout.Constraints{
			MinWidth:  0,
			MaxWidth:  100,
			MinHeight: 0,
			MaxHeight: 100,
		})

		rows := inst.GetRows()
		rowHeights := inst.GetRowHeights()

		fmt.Printf("    Total Size: %dw x %dh\n", size.Width, size.Height)
		fmt.Printf("    Row Count: %d\n", len(rows))

		for i, row := range rows {
			effectiveRowGap := tc.rowGap
			if effectiveRowGap == 0 {
				effectiveRowGap = tc.gap
			}
			fmt.Printf("    Row %d: %d items, height=%d, children=%v\n",
				i+1, len(row), rowHeights[i], row)
			if i < len(rows)-1 {
				fmt.Printf("           (row gap: %d)\n", effectiveRowGap)
			}
		}
	}
}
