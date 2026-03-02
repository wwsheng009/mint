// Package main demonstrates the Fiber-first Panel component.
// Panel is a high-level container that manages borders, headers, and content layout.
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
	"github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/ui/components/panel"
	newtext "github.com/wwsheng009/mint/ui/components/text"
)

// DemoApp creates the demo UI
func DemoApp() rtui.VNode {
	return ui.NewVStack().
		SetWidth(70).
		SetGap(1).
		SetChildrenList([]rtui.VNode{
			// Title
			sectionTitle("Panel Component Demo"),
			newtext.New(""),

			// =====================================================
			// Section 1: Basic Panel with Title
			// =====================================================
			subTitle("1. Basic Panel with Rounded Border"),
			panel.NewBuilder().
				Title("Basic Panel").
				Content(newtext.New("This is the main content area of the panel.")).
				Width(40).
				Height(5).
				Rounded().
				Build(),
			newtext.New(""),

			// =====================================================
			// Section 2: Panel with Header and Footer
			// =====================================================
			subTitle("2. Panel with Header and Footer"),
			panel.NewBuilder().
				Title("Complete Panel").
				Header(newtext.New("━ Header Line ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")).
				Content(newtext.New("Content goes here.\nMultiple lines are supported.\nPanel handles layout automatically.")).
				Footer(newtext.New("━ Footer Line ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")).
				Width(50).
				Height(8).
				BorderColor("cyan").
				Build(),
			newtext.New(""),

			// =====================================================
			// Section 3: Border Styles
			// =====================================================
			subTitle("3. Border Styles (Double, Single, Rounded)"),
			ui.NewHStack().
				SetGap(2).
				SetChildrenList([]rtui.VNode{
					panel.NewBuilder().
						Title("Double").
						Content(newtext.New("═══ Double Border ═══\nStyle uses ╔╗╚╝ chars")).
						Width(20).
						Height(5).
						Double().
						BorderColor("yellow").
						Build(),
					panel.NewBuilder().
						Title("Single").
						Content(newtext.New("Single-line border\nUses ┌┐└┘ chars")).
						Width(20).
						Height(5).
						Single().
						BorderColor("green").
						Build(),
					panel.NewBuilder().
						Title("Rounded").
						Content(newtext.New("Rounded corners\nUses ╭╮╰╯ chars")).
						Width(20).
						Height(5).
						Rounded().
						BorderColor("blue").
						Build(),
				}),
			newtext.New(""),

			// =====================================================
			// Section 4: No Border
			// =====================================================
			subTitle("4. No Border Mode"),
			panel.NewBuilder().
				Title("No Border").
				Content(newtext.New("This panel has no border.\nJust pure content area.\nFor minimalist designs.")).
				Width(40).
				Height(5).
				NoBorder().
				Build(),
			newtext.New(""),

			// =====================================================
			// Section 5: Custom Border Label
			// =====================================================
			subTitle("5. Custom Border Label"),
			panel.NewBuilder().
				Label(" ⚡ Custom Label ⚡ ").
				Content(newtext.New("The label appears in the top border.\nYou can customize it independently from the title.")).
				Width(45).
				Height(5).
				Rounded().
				BorderColor("magenta").
				Build(),
			newtext.New(""),

			// =====================================================
			// Section 6: Flex Panel
			// =====================================================
			subTitle("6. Flex Panel (expands to fill space)"),
			panel.NewBuilder().
				Title("Flex Panel").
				Content(newtext.New("This panel has flex=1.\nIt will expand to fill available space in a stack.\nUseful for responsive layouts.")).
				Width(50).
				Flex(1).
				BorderColor("blue").
				Build(),
			newtext.New(""),

			// =====================================================
			// Section 7: Text Wrapping
			// =====================================================
			subTitle("7. Text Wrapping in Panel"),
			panel.NewBuilder().
				Title("Text Wrap Demo").
				Content(newtext.New("This is a very long text that should wrap to multiple lines when displayed inside a panel. The wrap feature automatically breaks text at word boundaries while preserving readability.").SetWrap(true)).
				Width(40).
				Height(8).
				BorderColor("green").
				Build(),
			newtext.New(""),
			ui.NewHStack().
				SetGap(3).
				SetChildrenList([]rtui.VNode{
					panel.NewBuilder().
						Title("No Wrap").
						Content(newtext.New("This text is too long and will be truncated.").SetWrap(false)).
						Width(20).
						Height(3).
						BorderColor("red").
						Build(),
					panel.NewBuilder().
						Title("With Wrap").
						Content(newtext.New("This text is too long and will be wrapped to multiple lines.").SetWrap(true)).
						Width(20).
						BorderColor("green").
						Build(),
				}),
			newtext.New(""),

			// Footer
			highlight("Panel: borders, headers, footers, flexible layout, multiple border styles"),
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

	fmt.Println("╔══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║   Fiber-First Panel Component Demo                               ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════╝")

	// Create framework app (required for Fiber reconciler)
	fwApp := framework.NewApp()

	// Create DeclarativeNode WITH Fiber reconciler
	node := render.NewDeclarativeNodeFromFuncWithFiber(DemoApp, fwApp)

	// Enable Fiber-first mode
	node.SetRenderMode(render.RenderModeFiberFirst)

	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("  Render Mode: %v\n", node.GetRenderMode())

	// Create buffer (70 wide, 90 tall)
	buf := paint.NewBuffer(70, 90)

	// Create paint context
	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 70, Height: 90},
		AvailableWidth:  70,
		AvailableHeight: 90,
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 70))
	fmt.Println("Rendering Panel components...")
	fmt.Printf("%s\n\n", strings.Repeat("=", 70))

	// Render
	node.Paint(ctx, buf)

	// Output result
	utils.PrintBuffer(buf, 70, 90)

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("Panel Component Features:")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("  Layout Options:")
	fmt.Println("    - Width/Height: Fixed dimensions")
	fmt.Println("    - Height=0: Auto height based on content")
	fmt.Println("    - Flex: Flex factor for expansion")
	fmt.Println("    - Padding: Inner padding")
	fmt.Println("")
	fmt.Println("  Border Styles:")
	fmt.Println("    - Rounded: ╭╮╰╯ (default for titled panels)")
	fmt.Println("    - Single:  ┌┐└┘")
	fmt.Println("    - Double:  ╔╗╚╝")
	fmt.Println("    - None:    No border")
	fmt.Println("")
	fmt.Println("  Content Areas:")
	fmt.Println("    - Header: Optional header component")
	fmt.Println("    - Content: Main content (required)")
	fmt.Println("    - Footer: Optional footer component")
	fmt.Println("")
	fmt.Println("  Styling:")
	fmt.Println("    - BorderColor: Color of the border")
	fmt.Println("    - Title: Sets title (appears in border label)")
	fmt.Println("    - Label: Custom border label")
	fmt.Println("")
	fmt.Println("  Convenience Functions:")
	fmt.Println("    - panel.Of(content)")
	fmt.Println("    - panel.OfSize(content, w, h)")
	fmt.Println("    - panel.Titled(title, content)")
	fmt.Println("    - panel.Bordered(content, w, h)")
	fmt.Println(strings.Repeat("=", 70))
}
