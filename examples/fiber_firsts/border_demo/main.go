// Fiber-first Border Component Demo
// Demonstrates container with native border properties (Migration to new API)
//
// ═══ Migration Guide ═══
// Old API (wrapping):    border.New().Label("Title").SetChild(content)
// New API (container):   stack.NewVStack().SingleBorder("Title").SetChildrenList([content])
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

	"github.com/wwsheng009/mint/examples/utils"
	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/internal/render"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	newborder "github.com/wwsheng009/mint/ui/components/border"
	newbutton "github.com/wwsheng009/mint/ui/components/button"
	newstack "github.com/wwsheng009/mint/ui/components/stack"
	newtext "github.com/wwsheng009/mint/ui/components/text"
	"github.com/wwsheng009/mint/ui/components/wrap"
)

// DemoApp renders Border containers using the Fiber-first components
func DemoApp() rtui.VNode {
	return newstack.NewVStack().
		SetGap(1).
		SetChildrenList([]rtui.VNode{
			// =====================================================
			// Section 1: AUTO-MEASURE Feature (NEW!)
			// =====================================================
			sectionTitle("═══ 1. Auto-Measure (NEW!) ═══"),

			subTitle("1.1 No size - auto measures child"),
			highlight("Border auto-sizes to fit child content!"),
			newborder.New().
				Rounded().
				SetBorderColor("green").
				SetChild(newtext.New("Hello World")), // No width/height!

			subTitle("1.2 Auto height with explicit width=30"),
			newborder.New().
				SetWidth(30).
				// No height - auto measured
				SetChild(newtext.New("This text is wrapped in a 30-char wide border")),

			subTitle("1.3 Auto width with explicit height=2"),
			newborder.New().
				SetHeight(2).
				// No width - auto measured
				SetChild(newtext.New("Short")),

			// =====================================================
			// Section 2: Auto-Measure with Complex Children
			// =====================================================
			sectionTitle("═══ 2. Auto-Measure Complex Children ═══"),

			subTitle("2.1 VStack child (auto-sized)"),
			newborder.New().
				Rounded().
				Label(" Menu ").
				SetBorderColor("cyan").
				// No size - auto measures VStack
				SetChild(
					newstack.NewVStack().
						SetGap(0).
						SetChildrenList([]rtui.VNode{
							newtext.New("1. New File"),
							newtext.New("2. Open File"),
							newtext.New("3. Save"),
							newtext.New("4. Exit"),
						}),
				),

			subTitle("2.2 HStack child (auto-sized)"),
			newborder.New().
				SetBorderColor("yellow").
				// No size - auto measures HStack
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

			subTitle("2.3 Wrap child (auto-sized)"),
			newborder.New().
				Rounded().
				Label(" Tags ").
				SetBorderColor("magenta").
				// No size - auto measures Wrap
				SetChild(
					wrap.New().
						SetWidth(40).
						SetGap(1).
						SetChildrenList([]rtui.VNode{
							newtext.New("#golang"),
							newtext.New("#tui"),
							newtext.New("#terminal"),
							newtext.New("#fiber"),
							newtext.New("#ui"),
						}),
				),

			subTitle("2.4 Button children (auto-sized)"),
			newborder.New().
				SetBorderColor("blue").
				// No size - auto measures HStack with buttons
				SetChild(
					newstack.NewHStack().
						SetGap(1).
						SetChildrenList([]rtui.VNode{
							newbutton.New("OK").SetVariant(newbutton.VariantPrimary),
							newbutton.New("Cancel"),
							newbutton.New("Apply"),
						}),
				),

			// =====================================================
			// Section 3: Explicit vs Auto Comparison
			// =====================================================
			sectionTitle("═══ 3. Explicit vs Auto ═══"),

			subTitle("3.1 Explicit size (25x2)"),
			newborder.New().
				SetWidth(25).
				SetHeight(2).
				SetBorderColor("blue").
				SetChild(newtext.New("Fixed size border")),

			subTitle("3.2 Auto size (same content)"),
			newborder.New().
				// No size - fits content exactly
				SetBorderColor("green").
				SetChild(newtext.New("Fixed size border")),

			subTitle("3.3 Explicit with extra space"),
			newborder.New().
				SetWidth(40).
				SetHeight(3).
				SetBorderColor("yellow").
				SetChild(newtext.New("Content in larger box")),

			// =====================================================
			// Section 4: Border Styles
			// =====================================================
			sectionTitle("═══ 4. Border Styles ═══"),

			subTitle("4.1 Single (default)"),
			newborder.New().
				SetChild(newtext.New("Single line border")),

			subTitle("4.2 Double"),
			newborder.New().
				Double().
				SetChild(newtext.New("Double line border")),

			subTitle("4.3 Rounded"),
			newborder.New().
				Rounded().
				SetChild(newtext.New("Rounded corners")),

			subTitle("4.4 Dashed"),
			newborder.New().
				Dashed().
				SetChild(newtext.New("Dashed line border")),

			// =====================================================
			// Section 5: Border with Label
			// =====================================================
			sectionTitle("═══ 5. Border with Label ═══"),

			subTitle("5.1 Label (auto-sized)"),
			newborder.New().
				Rounded().
				Label(" Settings ").
				SetBorderColor("cyan").
				SetChild(
					newstack.NewVStack().
						SetGap(0).
						SetChildrenList([]rtui.VNode{
							newtext.New("Theme: Dark"),
							newtext.New("Font: Mono"),
						}),
				),

			subTitle("5.2 Long label (auto-sized)"),
			newborder.New().
				Double().
				Label(" Configuration Panel ").
				SetBorderColor("yellow").
				SetChild(newtext.New("Settings content here")),

			// =====================================================
			// Section 6: Border Colors
			// =====================================================
			sectionTitle("═══ 6. Border Colors ═══"),

			subTitle("6.1 Green"),
			newborder.New().Rounded().Color("green").SetChild(newtext.New("Success message")),

			subTitle("6.2 Red"),
			newborder.New().Color("red").SetChild(newtext.New("Error message")),

			subTitle("6.3 Yellow"),
			newborder.New().Color("yellow").SetChild(newtext.New("Warning message")),

			subTitle("6.4 Cyan"),
			newborder.New().Rounded().Color("cyan").SetChild(newtext.New("Info message")),

			// =====================================================
			// Section 7: Nested Borders (Auto-Sized)
			// =====================================================
			sectionTitle("═══ 7. Nested Borders ═══"),

			subTitle("7.1 Nested auto-sized"),
			newborder.New().
				Double().
				Label(" Outer ").
				SetBorderColor("blue").
				// No size - auto measures inner border
				SetChild(
					newborder.New().
						Rounded().
						Label(" Inner ").
						SetBorderColor("green").
						// No size - auto measures content
						SetChild(newtext.New("Deeply nested content")),
				),

			// =====================================================
			// Section 8: Real World Examples
			// =====================================================
			sectionTitle("═══ 8. Real World ═══"),

			subTitle("8.1 Dialog (auto-sized to content)"),
			newborder.New().
				Double().
				Label(" Confirm ").
				SetBorderColor("yellow").
				SetChild(
					newstack.NewVStack().
						SetGap(1).
						SetChildrenList([]rtui.VNode{
							newtext.New("Delete this file?"),
							newstack.NewHStack().
								SetGap(2).
								SetChildrenList([]rtui.VNode{
									newbutton.New("Yes").SetVariant(newbutton.VariantDanger),
									newbutton.New("No"),
								}),
						}),
				),

			subTitle("8.2 Info Panel"),
			newborder.New().
				Rounded().
				Label(" Status ").
				Color("green").
				SetChild(
					newstack.NewVStack().
						SetGap(0).
						SetChildrenList([]rtui.VNode{
							newtext.New("Server: Running"),
							newtext.New("Memory: 256MB"),
							newtext.New("CPU: 12%"),
						}),
				),

			subTitle("8.3 Error Box"),
			newborder.New().
				Color("red").
				Label(" ERROR ").
				SetChild(newtext.New("Connection failed: timeout")),
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

// highlight creates a highlighted note
func highlight(text string) rtui.VNode {
	return newtext.New("    >>> " + text + " <<<").
		Foreground("yellow")
}

func main() {
	os.Setenv("MINT_USE_FIBER", "true")
	os.Setenv("MINT_FIBER_FIRST", "true")

	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║   Fiber-First Border Rendering Demo                        ║")
	fmt.Println("║   (Border with AUTO-MEASURE feature)                       ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")

	// Create framework app (required for Fiber reconciler)
	fwApp := framework.NewApp()

	// Create DeclarativeNode WITH Fiber reconciler
	node := render.NewDeclarativeNodeFromFuncWithFiber(DemoApp, fwApp)

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
	fmt.Println("Rendering Border containers with AUTO-MEASURE...")
	fmt.Printf("%s\n\n", strings.Repeat("=", 60))

	// Render
	node.Paint(ctx, buf)

	// Output result
	utils.PrintBuffer(buf, 60, 90)

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Border Component Features:")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  AUTO-MEASURE (NEW!):")
	fmt.Println("    - No width/height needed - auto measures child")
	fmt.Println("    - Set only width -> auto height")
	fmt.Println("    - Set only height -> auto width")
	fmt.Println("")
	fmt.Println("  Border Styles:")
	fmt.Println("    - Single: ┌─┐ │ └─┘")
	fmt.Println("    - Double: ╔═╗ ║ ╚═╝")
	fmt.Println("    - Rounded: ╭─╮ │ ╰─╯")
	fmt.Println("    - Dashed: +-+ | +-+")
	fmt.Println("")
	fmt.Println("  Other Features:")
	fmt.Println("    - Label: Title on top border")
	fmt.Println("    - Color: Border color (blue, green, red, yellow...)")
	fmt.Println("")
	fmt.Println("  Layout Impact:")
	fmt.Println("    - Single/Rounded/Dashed: +2 width, +2 height")
	fmt.Println("    - Double: +4 width, +4 height")
	fmt.Println(strings.Repeat("=", 60))
}
