// Fiber-first Stack Component Demo
// Demonstrates the new Stack (HStack/VStack) component following the Fiber-first architecture
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
	newbutton "github.com/wwsheng009/mint/ui/components/button"
	newstack "github.com/wwsheng009/mint/ui/components/stack"
	newtext "github.com/wwsheng009/mint/ui/components/text"
)

// DemoApp renders Stack layouts using the Fiber-first components
func DemoApp() rtui.VNode {
	return newstack.NewVStack().
		SetGap(0).
		SetChildrenList([]rtui.VNode{
			// =====================================================
			// Section 1: Basic Layouts
			// =====================================================
			sectionTitle("═══ 1. Basic HStack & VStack ═══"),

			// 1.1 Basic HStack (Horizontal)
			subTitle("1.1 HStack (Horizontal)"),
			newstack.NewHStack().
				SetGap(2).
				SetChildrenList([]rtui.VNode{
					newtext.New("[A]"),
					newtext.New("[B]"),
					newtext.New("[C]"),
				}),

			// 1.2 Basic VStack (Vertical)
			subTitle("1.2 VStack (Vertical)"),
			newstack.NewVStack().
				SetGap(0).
				SetChildrenList([]rtui.VNode{
					newtext.New("├─ Item 1"),
					newtext.New("├─ Item 2"),
					newtext.New("└─ Item 3"),
				}),

			// =====================================================
			// Section 2: Gap Spacing
			// =====================================================
			sectionTitle("═══ 2. Gap Spacing ═══"),

			// 2.1 No Gap
			subTitle("2.1 Gap=0 (No spacing)"),
			newstack.NewHStack().
				SetGap(0).
				SetChildrenList([]rtui.VNode{
					newtext.New("[A]"),
					newtext.New("[B]"),
					newtext.New("[C]"),
				}),

			// 2.2 Gap=2
			subTitle("2.2 Gap=2"),
			newstack.NewHStack().
				SetGap(2).
				SetChildrenList([]rtui.VNode{
					newtext.New("[A]"),
					newtext.New("[B]"),
					newtext.New("[C]"),
				}),

			// 2.3 Gap=5
			subTitle("2.3 Gap=5"),
			newstack.NewHStack().
				SetGap(5).
				SetChildrenList([]rtui.VNode{
					newtext.New("[A]"),
					newtext.New("[B]"),
					newtext.New("[C]"),
				}),

			// =====================================================
			// Section 3: Main Axis Alignment
			// =====================================================
			sectionTitle("═══ 3. Main Axis Alignment ═══"),

			// 3.1 Align Start (default)
			subTitle("3.1 AlignStart (left)"),
			newstack.NewHStack().
				SetWidth(40).
				SetChildrenList([]rtui.VNode{
					newtext.New("[A]"),
					newtext.New("[B]"),
					newtext.New("[C]"),
				}),

			// 3.2 Align Center
			subTitle("3.2 AlignCenter"),
			newstack.NewHStack().
				SetWidth(40).
				Center().
				SetChildrenList([]rtui.VNode{
					newtext.New("[A]"),
					newtext.New("[B]"),
					newtext.New("[C]"),
				}),

			// 3.3 Align End (right)
			subTitle("3.3 AlignEnd (right)"),
			newstack.NewHStack().
				SetWidth(40).
				SetAlign(newstack.AlignEnd).
				SetChildrenList([]rtui.VNode{
					newtext.New("[A]"),
					newtext.New("[B]"),
					newtext.New("[C]"),
				}),

			// 3.4 SpaceBetween
			subTitle("3.4 SpaceBetween"),
			newstack.NewHStack().
				SetWidth(40).
				SetAlign(newstack.AlignSpaceBetween).
				SetChildrenList([]rtui.VNode{
					newtext.New("[A]"),
					newtext.New("[B]"),
					newtext.New("[C]"),
				}),

			// =====================================================
			// Section 4: Cross Axis Alignment
			// =====================================================
			sectionTitle("═══ 4. Cross Axis Alignment ═══"),

			// 4.1 Cross Start (top)
			subTitle("4.1 CrossStart (top)"),
			newstack.NewHStack().
				SetHeight(3).
				SetGap(1).
				SetChildrenList([]rtui.VNode{
					newtext.New("A"),
					newtext.New("B"),
					newtext.New("C"),
				}),

			// 4.2 Cross Center
			subTitle("4.2 CrossCenter"),
			newstack.NewHStack().
				SetHeight(3).
				SetGap(1).
				CenterCross().
				SetChildrenList([]rtui.VNode{
					newtext.New("A"),
					newtext.New("B"),
					newtext.New("C"),
				}),

			// 4.3 Cross End (bottom)
			subTitle("4.3 CrossEnd (bottom)"),
			newstack.NewHStack().
				SetHeight(3).
				SetGap(1).
				SetCrossAlign(newstack.AlignEnd).
				SetChildrenList([]rtui.VNode{
					newtext.New("A"),
					newtext.New("B"),
					newtext.New("C"),
				}),

			// =====================================================
			// Section 5: Stretch Cross Axis
			// =====================================================
			sectionTitle("═══ 5. Stretch Cross Axis ═══"),

			// 5.1 HStack Stretch (children fill height)
			subTitle("5.1 HStack Stretch"),
			newstack.NewHStack().
				SetHeight(3).
				SetGap(1).
				Stretch().
				SetChildrenList([]rtui.VNode{
					boxText("A", 5),
					boxText("B", 5),
					boxText("C", 5),
				}),

			// 5.2 VStack Stretch (children fill width)
			subTitle("5.2 VStack Stretch"),
			newstack.NewVStack().
				SetWidth(30).
				SetGap(0).
				Stretch().
				SetChildrenList([]rtui.VNode{
					newtext.New("┌────────────────────────────┐"),
					newtext.New("│      Stretched Width       │"),
					newtext.New("└────────────────────────────┘"),
				}),

			// =====================================================
			// Section 6: Spacer & Flex Layout
			// =====================================================
			sectionTitle("═══ 6. Spacer & Flex Layout ═══"),

			// 6.1 Spacer pushes content to right
			subTitle("6.1 Spacer (Left | Spacer | Right)"),
			newstack.NewHStack().
				SetWidth(40).
				SetChildrenList([]rtui.VNode{
					newtext.New("Left"),
					newstack.Spacer(1),
					newtext.New("Right"),
				}),

			// 6.2 Multiple Spacers
			subTitle("6.2 Multiple Spacers"),
			newstack.NewHStack().
				SetWidth(40).
				SetChildrenList([]rtui.VNode{
					newtext.New("A"),
					newstack.Spacer(1),
					newtext.New("B"),
					newstack.Spacer(1),
					newtext.New("C"),
				}),

			// 6.3 Two elements with spacer
			subTitle("6.3 [OK] Spacer [Cancel]"),
			newstack.NewHStack().
				SetWidth(40).
				SetGap(1).
				SetChildrenList([]rtui.VNode{
					newbutton.New("OK").SetVariant(newbutton.VariantPrimary),
					newstack.Spacer(1),
					newbutton.New("Cancel"),
				}),

			// =====================================================
			// Section 7: Padding
			// =====================================================
			sectionTitle("═══ 7. Padding ═══"),

			// 7.1 Padding All Sides
			subTitle("7.1 Padding(1,2,1,2)"),
			newstack.NewHStack().
				SetPadding(1, 2, 1, 2).
				SetChildrenList([]rtui.VNode{
					newtext.New("Padded"),
				}),

			// 7.2 Padding with content
			subTitle("7.2 Padding Container"),
			newstack.NewVStack().
				SetPadding(1, 3, 1, 3).
				SetChildrenList([]rtui.VNode{
					newtext.New("┌─ Inner Content ─┐"),
					newtext.New("│  With Padding   │"),
					newtext.New("└─────────────────┘"),
				}),

			// =====================================================
			// Section 8: Buttons Layout
			// =====================================================
			sectionTitle("═══ 8. Buttons Layout ═══"),

			// 8.1 Buttons in HStack
			subTitle("8.1 Buttons Row"),
			newstack.NewHStack().
				SetGap(2).
				SetChildrenList([]rtui.VNode{
					newbutton.New("OK").SetVariant(newbutton.VariantPrimary),
					newbutton.New("Cancel"),
					newbutton.New("Help"),
				}),

			// 8.2 Buttons with SpaceBetween
			subTitle("8.2 Buttons SpaceBetween"),
			newstack.NewHStack().
				SetWidth(40).
				SetAlign(newstack.AlignSpaceBetween).
				SetChildrenList([]rtui.VNode{
					newbutton.New("Back"),
					newbutton.New("Next").SetVariant(newbutton.VariantPrimary),
				}),

			// =====================================================
			// Section 9: Nested Stacks
			// =====================================================
			sectionTitle("═══ 9. Nested Stacks ═══"),

			// 9.1 Grid-like layout
			subTitle("9.1 Grid (HStack in VStack)"),
			newstack.NewVStack().
				SetGap(0).
				SetChildrenList([]rtui.VNode{
					newstack.NewHStack().
						SetGap(1).
						SetChildrenList([]rtui.VNode{
							newtext.New("[1,1]"),
							newtext.New("[1,2]"),
							newtext.New("[1,3]"),
						}),
					newstack.NewHStack().
						SetGap(1).
						SetChildrenList([]rtui.VNode{
							newtext.New("[2,1]"),
							newtext.New("[2,2]"),
							newtext.New("[2,3]"),
						}),
					newstack.NewHStack().
						SetGap(1).
						SetChildrenList([]rtui.VNode{
							newtext.New("[3,1]"),
							newtext.New("[3,2]"),
							newtext.New("[3,3]"),
						}),
				}),

			// =====================================================
			// Section 10: Complex Layout
			// =====================================================
			sectionTitle("═══ 10. Complex Layout ═══"),

			// 10.1 Toolbar-like layout
			subTitle("10.1 Toolbar"),
			newstack.NewHStack().
				SetWidth(50).
				SetAlign(newstack.AlignSpaceBetween).
				CenterCross().
				SetHeight(3).
				SetChildrenList([]rtui.VNode{
					newstack.NewHStack().
						SetGap(1).
						SetChildrenList([]rtui.VNode{
							newtext.New("📁"),
							newtext.New("File"),
							newtext.New("📝"),
							newtext.New("Edit"),
						}),
					newstack.NewHStack().
						SetGap(1).
						SetChildrenList([]rtui.VNode{
							newbutton.New("?"),
						}),
				}),

			// 10.2 Status bar
			subTitle("10.2 Status Bar"),
			newstack.NewHStack().
				SetWidth(50).
				SetAlign(newstack.AlignSpaceBetween).
				SetChildrenList([]rtui.VNode{
					newtext.New("Ready"),
					newtext.New("Ln 1, Col 1"),
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

// boxText creates a text with box markers
func boxText(text string, width int) rtui.VNode {
	return newtext.New("[" + text + "]")
}

func main() {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")
	os.Setenv("MINT_DEBUG_TEST", "true")

	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║   Fiber-First Stack Rendering Demo                         ║")
	fmt.Println("║   (HStack / VStack / Spacer components)                    ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")

	// Create framework app (required for Fiber reconciler)
	fwApp := framework.NewApp()

	// Create DeclarativeNode WITH Fiber reconciler
	node := render.NewDeclarativeNodeFromFuncWithFiber(DemoApp, fwApp)

	// Enable Fiber-first mode
	node.SetRenderMode(render.RenderModeFiberFirst)

	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("  Render Mode: %v\n", node.GetRenderMode())
	fmt.Printf("  Fiber-First Enabled: %v\n", node.IsFiberFirstEnabled())

	// Create buffer (60 wide, 120 tall to fit all tests)
	buf := paint.NewBuffer(60, 120)

	// Create paint context
	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 60, Height: 120},
		AvailableWidth:  60,
		AvailableHeight: 120,
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 60))
	fmt.Println("Rendering Stack layouts with Fiber-first pipeline...")
	fmt.Printf("%s\n\n", strings.Repeat("=", 60))

	// Render
	node.Paint(ctx, buf)

	// Output result
	printBuffer(buf, 60, 120)

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Stack Component Features:")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  - HStack: Horizontal layout (Row)")
	fmt.Println("  - VStack: Vertical layout (Column)")
	fmt.Println("  - Gap: Spacing between children")
	fmt.Println("  - Padding: Inner spacing [top, right, bottom, left]")
	fmt.Println("  - Align: Start | Center | End | SpaceBetween")
	fmt.Println("  - CrossAlign: Start | Center | End")
	fmt.Println("  - Stretch: Children fill cross axis")
	fmt.Println("  - Spacer: Flexible space distribution")
	fmt.Println(strings.Repeat("=", 60))
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
