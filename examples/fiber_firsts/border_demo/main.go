// Fiber-first Border Component Demo
// Demonstrates the Border container component following the Fiber-first architecture
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
	newborder "github.com/wwsheng009/mint/ui/components/border"
	newstack "github.com/wwsheng009/mint/ui/components/stack"
	newtext "github.com/wwsheng009/mint/ui/components/text"
)

// DemoApp renders Border containers using the Fiber-first components
func DemoApp() rtui.VNode {
	return newstack.NewVStack().
		SetGap(1).
		SetChildrenList([]rtui.VNode{
			// =====================================================
			// Section 1: Border Styles
			// =====================================================
			sectionTitle("═══ 1. Border Styles ═══"),

			// 1.1 Single Line Border
			subTitle("1.1 Single (default)"),
			newborder.New().
				Single().
				SetWidth(20).
				SetHeight(2).
				SetChild(newtext.New("Single line border")),

			// 1.2 Double Line Border
			subTitle("1.2 Double"),
			newborder.New().
				Double().
				SetWidth(20).
				SetHeight(2).
				SetChild(newtext.New("Double line border")),

			// 1.3 Rounded Border
			subTitle("1.3 Rounded"),
			newborder.New().
				Rounded().
				SetWidth(20).
				SetHeight(2).
				SetChild(newtext.New("Rounded corners")),

			// 1.4 Dashed Border
			subTitle("1.4 Dashed"),
			newborder.New().
				Dashed().
				SetWidth(20).
				SetHeight(2).
				SetChild(newtext.New("Dashed line border")),

			// =====================================================
			// Section 2: Border with Label
			// =====================================================
			sectionTitle("═══ 2. Border with Label ═══"),

			// 2.1 Label on top
			subTitle("2.1 With Label"),
			newborder.New().
				Label(" Settings ").
				SetWidth(30).
				SetHeight(3).
				SetChild(newtext.New("Configuration options here")),

			// 2.2 Label with double border
			subTitle("2.2 Double + Label"),
			newborder.New().
				Double().
				Label(" Important ").
				SetWidth(25).
				SetHeight(2).
				SetChild(newtext.New("Critical information!")),

			// 2.3 Label with rounded border
			subTitle("2.3 Rounded + Label"),
			newborder.New().
				Rounded().
				Label(" Panel ").
				SetWidth(25).
				SetHeight(2).
				SetChild(newtext.New("Modern UI panel")),

			// =====================================================
			// Section 3: Border Colors
			// =====================================================
			sectionTitle("═══ 3. Border Colors ═══"),

			// 3.1 Blue border (default)
			subTitle("3.1 Blue (default)"),
			newborder.New().
				Color("blue").
				SetWidth(15).
				SetHeight(1).
				SetChild(newtext.New("Blue border")),

			// 3.2 Green border
			subTitle("3.2 Green"),
			newborder.New().
				Color("green").
				SetWidth(15).
				SetHeight(1).
				SetChild(newtext.New("Green border")),

			// 3.3 Red border
			subTitle("3.3 Red"),
			newborder.New().
				Color("red").
				SetWidth(15).
				SetHeight(1).
				SetChild(newtext.New("Red border")),

			// 3.4 Yellow border
			subTitle("3.4 Yellow"),
			newborder.New().
				Color("yellow").
				SetWidth(15).
				SetHeight(1).
				SetChild(newtext.New("Yellow border")),

			// =====================================================
			// Section 4: Border Sizes
			// =====================================================
			sectionTitle("═══ 4. Border Sizes ═══"),

			// 4.1 Small content
			subTitle("4.1 Small (10x1)"),
			newborder.New().
				SetWidth(10).
				SetHeight(1).
				SetChild(newtext.New("Small")),

			// 4.2 Medium content
			subTitle("4.2 Medium (25x3)"),
			newborder.New().
				SetWidth(25).
				SetHeight(3).
				SetChild(newtext.New("Medium sized content\nwith multiple\nlines")),

			// 4.3 Wide content
			subTitle("4.3 Wide (40x2)"),
			newborder.New().
				Rounded().
				SetWidth(40).
				SetHeight(2).
				SetChild(newtext.New("This is a wide border with more content space")),

			// =====================================================
			// Section 5: Border with Stack Content
			// =====================================================
			sectionTitle("═══ 5. Border + Stack ═══"),

			// 5.1 Border with VStack
			subTitle("5.1 Border + VStack"),
			newborder.New().
				Rounded().
				Label(" Menu ").
				SetWidth(20).
				SetHeight(4).
				SetChild(
					newstack.NewVStack().
						SetGap(0).
						SetChildrenList([]rtui.VNode{
							newtext.New("1. Option One"),
							newtext.New("2. Option Two"),
							newtext.New("3. Option Three"),
						}),
				),

			// 5.2 Border with HStack
			subTitle("5.2 Border + HStack"),
			newborder.New().
				SetWidth(30).
				SetHeight(1).
				SetChild(
					newstack.NewHStack().
						SetGap(2).
						SetChildrenList([]rtui.VNode{
							newtext.New("[File]"),
							newtext.New("[Edit]"),
							newtext.New("[View]"),
							newtext.New("[Help]"),
						}),
				),

			// =====================================================
			// Section 6: Nested Borders
			// =====================================================
			sectionTitle("═══ 6. Nested Borders ═══"),

			// 6.1 Nested borders
			subTitle("6.1 Nested Borders"),
			newborder.New().
				Double().
				Label(" Outer ").
				SetWidth(32).
				SetHeight(6).
				SetChild(
					newborder.New().
						Single().
						Label(" Inner ").
						SetWidth(24).
						SetHeight(2).
						SetChild(newtext.New("Nested content inside")),
				),

			// =====================================================
			// Section 7: Side by Side Borders
			// =====================================================
			sectionTitle("═══ 7. Side by Side ═══"),

			// 7.1 Two panels side by side
			subTitle("7.1 Two Panels"),
			newstack.NewHStack().
				SetGap(1).
				SetChildrenList([]rtui.VNode{
					newborder.New().
						Rounded().
						Label(" Left ").
						SetWidth(15).
						SetHeight(3).
						SetChild(newtext.New("Left panel\ncontent")),
					newborder.New().
						Rounded().
						Label(" Right ").
						SetWidth(15).
						SetHeight(3).
						SetChild(newtext.New("Right panel\ncontent")),
				}),

			// =====================================================
			// Section 8: Real World Examples
			// =====================================================
			sectionTitle("═══ 8. Real World ═══"),

			// 8.1 Dialog box
			subTitle("8.1 Dialog Box"),
			newborder.New().
				Double().
				Label(" Confirm ").
				SetWidth(35).
				SetHeight(4).
				SetChild(
					newstack.NewVStack().
						SetGap(1).
						SetChildrenList([]rtui.VNode{
							newtext.New("Are you sure you want to delete?"),
							newstack.NewHStack().
								SetGap(2).
								SetAlign(newstack.AlignCenter).
								SetChildrenList([]rtui.VNode{
									newtext.New("[Yes]"),
									newtext.New("[No]"),
								}),
						}),
				),

			// 8.2 Info panel
			subTitle("8.2 Info Panel"),
			newborder.New().
				Rounded().
				Color("green").
				Label(" Info ").
				SetWidth(35).
				SetHeight(3).
				SetChild(
					newstack.NewVStack().
						SetGap(0).
						SetChildrenList([]rtui.VNode{
							newtext.New("Status: Running"),
							newtext.New("Progress: 75%"),
						}),
				),

			// 8.3 Error box
			subTitle("8.3 Error Box"),
			newborder.New().
				Color("red").
				Label(" ERROR ").
				SetWidth(35).
				SetHeight(2).
				SetChild(newtext.New("Failed to connect to server!")),
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

func main() {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")
	os.Setenv("MINT_DEBUG_TEST", "true")

	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║   Fiber-First Border Rendering Demo                        ║")
	fmt.Println("║   (Border container components)                            ║")
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

	// Create buffer (60 wide, 80 tall to fit all tests)
	buf := paint.NewBuffer(60, 80)

	// Create paint context
	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 60, Height: 80},
		AvailableWidth:  60,
		AvailableHeight: 80,
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 60))
	fmt.Println("Rendering Border containers with Fiber-first pipeline...")
	fmt.Printf("%s\n\n", strings.Repeat("=", 60))

	// Render
	node.Paint(ctx, buf)

	// Output result
	printBuffer(buf, 60, 80)

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Border Component Features:")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  - Single: ┌─┐ │ └─┘ (single line)")
	fmt.Println("  - Double: ╔═╗ ║ ╚═╝ (double line)")
	fmt.Println("  - Rounded: ╭─╮ │ ╰─╯ (rounded corners)")
	fmt.Println("  - Dashed: +-+ | +-+ (dashed line)")
	fmt.Println("  - Label: Title on top border")
	fmt.Println("  - Color: Border color (blue, green, red, yellow...)")
	fmt.Println("")
	fmt.Println("Border Layout Impact:")
	fmt.Println("  - Single/Dashed/Rounded: +2 width, +2 height")
	fmt.Println("  - Double: +4 width, +4 height")
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
