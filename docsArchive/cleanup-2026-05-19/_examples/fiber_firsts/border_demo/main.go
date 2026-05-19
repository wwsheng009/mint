// Fiber-first Border Component Demo
// Demonstrates native border properties on containers
//
// Border is now a native property of all containers:
// - Stack:   SingleBorder(), DoubleBorder(), RoundedBorder(), DashedBorder()
// - Grid:    Same border methods
// - Wrap:    Same border methods
// - Absolute: Same border methods
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
	"github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/ui/components/button"
	"github.com/wwsheng009/mint/ui/components/text"
	"github.com/wwsheng009/mint/ui/components/wrap"
)

// DemoApp renders containers with native border properties
func DemoApp() rtui.VNode {
	return ui.NewVStack().
		SetGap(1).
		SetChildrenList([]rtui.VNode{
			// =====================================================
			// Section 1: AUTO-MEASURE Feature
			// =====================================================
			sectionTitle("═══ 1. Auto-Measure ═══"),

			subTitle("1.1 No size - auto measures child"),
			highlight("Border auto-sizes to fit child content!"),
			// Simple text with rounded border - wrap in Stack
			ui.NewVStack().
				RoundedBorder().
				BorderColor("green").
				SetChildrenList([]rtui.VNode{
					text.New("Hello World"),
				}),

			subTitle("1.2 Auto height with explicit width=30"),
			ui.NewVStack().
				SingleBorder().
				SetWidth(30).
				SetChildrenList([]rtui.VNode{
					text.New("This text is wrapped in a 30-char wide border"),
				}),

			subTitle("1.3 Auto width with explicit height=2"),
			ui.NewVStack().
				SingleBorder().
				SetHeight(2).
				SetChildrenList([]rtui.VNode{
					text.New("Short"),
				}),

			// =====================================================
			// Section 2: Auto-Measure with Complex Children
			// =====================================================
			sectionTitle("═══ 2. Auto-Measure Complex Children ═══"),

			subTitle("2.1 VStack child (auto-sized)"),
			ui.NewVStack().
				RoundedBorder(" Menu ").
				BorderColor("cyan").
				SetGap(0).
				SetChildrenList([]rtui.VNode{
					text.New("1. New File"),
					text.New("2. Open File"),
					text.New("3. Save"),
					text.New("4. Exit"),
				}),

			subTitle("2.2 HStack child (auto-sized)"),
			ui.NewHStack().
				SingleBorder().
				BorderColor("yellow").
				SetGap(2).
				SetChildrenList([]rtui.VNode{
					text.New("[File]"),
					text.New("[Edit]"),
					text.New("[View]"),
					text.New("[Help]"),
				}),

			subTitle("2.3 Wrap child (auto-sized)"),
			wrap.New().
				SetWidth(40).
				RoundedBorder(" Tags ").
				BorderColor("magenta").
				SetGap(1).
				SetChildrenList([]rtui.VNode{
					text.New("#golang"),
					text.New("#tui"),
					text.New("#terminal"),
					text.New("#fiber"),
					text.New("#ui"),
				}),

			subTitle("2.4 Button children (auto-sized)"),
			ui.NewHStack().
				SingleBorder().
				BorderColor("blue").
				SetGap(1).
				SetChildrenList([]rtui.VNode{
					button.New("OK").SetVariant(button.VariantPrimary),
					button.New("Cancel"),
					button.New("Apply"),
				}),

			// =====================================================
			// Section 3: Explicit vs Auto Comparison
			// =====================================================
			sectionTitle("═══ 3. Explicit vs Auto ═══"),

			subTitle("3.1 Explicit size (25x2)"),
			ui.NewVStack().
				SingleBorder().
				SetWidth(25).
				SetHeight(2).
				BorderColor("blue").
				SetChildrenList([]rtui.VNode{
					text.New("Fixed size border"),
				}),

			subTitle("3.2 Auto size (same content)"),
			ui.NewVStack().
				SingleBorder().
				BorderColor("green").
				SetChildrenList([]rtui.VNode{
					text.New("Fixed size border"),
				}),

			subTitle("3.3 Explicit with extra space"),
			ui.NewVStack().
				SingleBorder().
				SetWidth(40).
				SetHeight(3).
				BorderColor("yellow").
				SetChildrenList([]rtui.VNode{
					text.New("Content in larger box"),
				}),

			// =====================================================
			// Section 4: Border Styles
			// =====================================================
			sectionTitle("═══ 4. Border Styles ═══"),

			subTitle("4.1 Single (default)"),
			ui.NewVStack().
				SingleBorder().
				SetChildrenList([]rtui.VNode{
					text.New("Single line border"),
				}),

			subTitle("4.2 Double"),
			ui.NewVStack().
				DoubleBorder().
				SetChildrenList([]rtui.VNode{
					text.New("Double line border"),
				}),

			subTitle("4.3 Rounded"),
			ui.NewVStack().
				RoundedBorder().
				SetChildrenList([]rtui.VNode{
					text.New("Rounded corners"),
				}),

			subTitle("4.4 Dashed"),
			ui.NewVStack().
				DashedBorder().
				SetChildrenList([]rtui.VNode{
					text.New("Dashed line border"),
				}),

			// =====================================================
			// Section 5: Border with Label
			// =====================================================
			sectionTitle("═══ 5. Border with Label ═══"),

			subTitle("5.1 Label (auto-sized)"),
			ui.NewVStack().
				RoundedBorder(" Settings ").
				BorderColor("cyan").
				SetGap(0).
				SetChildrenList([]rtui.VNode{
					text.New("Theme: Dark"),
					text.New("Font: Mono"),
				}),

			subTitle("5.2 Long label (auto-sized)"),
			ui.NewVStack().
				DoubleBorder(" Configuration Panel ").
				BorderColor("yellow").
				SetChildrenList([]rtui.VNode{
					text.New("Settings content here"),
				}),

			// =====================================================
			// Section 6: Border Colors
			// =====================================================
			sectionTitle("═══ 6. Border Colors ═══"),

			subTitle("6.1 Green"),
			ui.NewVStack().
				RoundedBorder().
				BorderColor("green").
				SetChildrenList([]rtui.VNode{
					text.New("Success message"),
				}),

			subTitle("6.2 Red"),
			ui.NewVStack().
				SingleBorder().
				BorderColor("red").
				SetChildrenList([]rtui.VNode{
					text.New("Error message"),
				}),

			subTitle("6.3 Yellow"),
			ui.NewVStack().
				SingleBorder().
				BorderColor("yellow").
				SetChildrenList([]rtui.VNode{
					text.New("Warning message"),
				}),

			subTitle("6.4 Cyan"),
			ui.NewVStack().
				RoundedBorder().
				BorderColor("cyan").
				SetChildrenList([]rtui.VNode{
					text.New("Info message"),
				}),

			// =====================================================
			// Section 7: Nested Borders
			// =====================================================
			sectionTitle("═══ 7. Nested Borders ═══"),

			subTitle("7.1 Nested auto-sized"),
			ui.NewVStack().
				DoubleBorder(" Outer ").
				BorderColor("blue").
				SetChildrenList([]rtui.VNode{
					ui.NewVStack().
						RoundedBorder(" Inner ").
						BorderColor("green").
						SetChildrenList([]rtui.VNode{
							text.New("Deeply nested content"),
						}),
				}),

			// =====================================================
			// Section 8: Real World Examples
			// =====================================================
			sectionTitle("═══ 8. Real World ═══"),

			subTitle("8.1 Dialog (auto-sized to content)"),
			ui.NewVStack().
				DoubleBorder(" Confirm ").
				BorderColor("yellow").
				SetGap(1).
				SetChildrenList([]rtui.VNode{
					text.New("Delete this file?"),
					ui.NewHStack().
						SetGap(2).
						SetChildrenList([]rtui.VNode{
							button.New("Yes").SetVariant(button.VariantDanger),
							button.New("No"),
						}),
				}),

			subTitle("8.2 Info Panel"),
			ui.NewVStack().
				RoundedBorder(" Status ").
				BorderColor("green").
				SetGap(0).
				SetChildrenList([]rtui.VNode{
					text.New("Server: Running"),
					text.New("Memory: 256MB"),
					text.New("CPU: 12%"),
				}),

			subTitle("8.3 Error Box"),
			ui.NewVStack().
				SingleBorder(" ERROR ").
				BorderColor("red").
				SetChildrenList([]rtui.VNode{
					text.New("Connection failed: timeout"),
				}),
		})
}

// sectionTitle creates a styled section title
func sectionTitle(title string) rtui.VNode {
	return text.New(title).
		Foreground(theme.Primary()).
		Bold(true)
}

// subTitle creates a subtitle
func subTitle(title string) rtui.VNode {
	return text.New("  " + title)
}

// highlight creates a highlighted note
func highlight(str string) rtui.VNode {
	return text.New("    >>> " + str + " <<<").
		Foreground("yellow")
}

func main() {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")

	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║   Fiber-First Border Rendering Demo                        ║")
	fmt.Println("║   (Native Border Properties)                              ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")

	// Create framework app (required for Fiber reconciler)
	fwApp := framework.NewApp()

	// Create DeclarativeNode WITH Fiber reconciler
	node := render.NewDeclarativeNodeFromFuncWithFiber(DemoApp)
    node.SetApp(fwApp)

	// Enable Fiber-first mode
	node.SetRenderMode(render.RenderModeFiberFirst)

	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("  Render Mode: %v\n", node.GetRenderMode())

	// Create buffer (60 wide, 90 tall)
	buf := paint.NewBuffer(60, 90)

	// Create paint context
	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 60, Height: 90},
		AvailableWidth:  60,
		AvailableHeight: 90,
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 60))
	fmt.Println("Rendering containers with native border properties...")
	fmt.Printf("%s\n\n", strings.Repeat("=", 60))

	// Render
	node.Paint(ctx, buf)

	// Output result with colors
	fmt.Print(buf.String())

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Border Component Features:")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  Native Border Properties:")
	fmt.Println("    - Border() sets style and label")
	fmt.Println("    - SingleBorder(), DoubleBorder(), etc.")
	fmt.Println("    - No wrapper API needed")
	fmt.Println("")
	fmt.Println("  Container Support:")
	fmt.Println("    - Stack: VStack, HStack")
	fmt.Println("    - Grid")
	fmt.Println("    - Wrap")
	fmt.Println("    - Absolute")
	fmt.Println("")
	fmt.Println("  Border Styles:")
	fmt.Println("    - Single: ┌─┐ │ └─┘")
	fmt.Println("    - Double: ╔═╗ ║ ╚═╝")
	fmt.Println("    - Rounded: ╭─╮ │ ╰─╯")
	fmt.Println("    - Dashed: +-+ | +-+")
	fmt.Println("")
	fmt.Println("  Other Features:")
	fmt.Println("    - Label: Title on top border")
	fmt.Println("    - Color: Border color (FgColor)")
	fmt.Println("")
	fmt.Println("  Layout Impact:")
	fmt.Println("    - Single/Rounded/Dashed: +2 width, +2 height")
	fmt.Println("    - Double: +4 width, +4 height")
	fmt.Println(strings.Repeat("=", 60))
}
