// Fiber-First Modal Component Demo
// Demonstrates the new Modal component following the Fiber-first architecture
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
	newstack "github.com/wwsheng009/mint/ui/components/stack"
	newtext "github.com/wwsheng009/mint/ui/components/text"
	"github.com/wwsheng009/mint/ui/components/modal"
)

// DemoApp creates the demo UI
func DemoApp() rtui.VNode {
	return newstack.New(newstack.Column).
		SetWidth(70).
		SetGap(1).
		SetChildrenList([]rtui.VNode{
			// Title
			sectionTitle("Fiber-First Modal Component Demo"),
			newtext.New(""),
			newtext.New("Modal dialog with Fiber-first architecture:"),
			newtext.New("  • Pure descriptive VNode"),
			newtext.New("  • Intent-based event handling"),
			newtext.New("  • Layer system integration"),
			newtext.New("  • Multiple border styles"),
			newtext.New("  • ESC key / click-outside to close"),
			newtext.New(""),

			// =====================================================
			// Section 1: Basic Open Modal
			// =====================================================
			subTitle("1. Basic Open Modal with Title"),
			modal.NewBuilder().
				Title("Modal Dialog").
				Content(newtext.New("This is a basic modal dialog.\nModal is rendered on LayerOverlay.\nClose with ESC or click outside.")).
				Width(45).
				Height(10).
				Rounded().
				Open(true).
				Build(),
			newtext.New(""),

			// =====================================================
			// Section 2: Border Styles
			// =====================================================
			subTitle("2. Border Styles (Rounded, Double, Single)"),
			newstack.New(newstack.Row).
				SetGap(2).
				SetChildrenList([]rtui.VNode{
					modal.NewBuilder().
						Title("Rounded").
						Content(newtext.New("╭╮╰╯ borders\nStyle uses rounded chars")).
						Width(20).
						Height(7).
						Rounded().
						Open(true).
						Build(),
					modal.NewBuilder().
						Title("Double").
						Content(newtext.New("╔╗╚╝ borders\nStyle uses double chars")).
						Width(20).
						Height(7).
						Double().
						Open(true).
						Build(),
					modal.NewBuilder().
						Title("Single").
						Content(newtext.New("┌┐└┘ borders\nStyle uses single chars")).
						Width(20).
						Height(7).
						Single().
						Open(true).
						Build(),
				}),
			newtext.New(""),

			// =====================================================
			// Section 3: Modal with Footer
			// =====================================================
			subTitle("3. Modal with Custom Footer"),
			modal.NewBuilder().
				Title("Confirmation Dialog").
				Content(newtext.New("Do you want to proceed?\nThis action cannot be undone.")).
				Footer(newtext.New("  [Esc/C] Cancel     [Enter] Confirm  ")).
				Width(40).
				Height(8).
				Rounded().
				Open(true).
				Build(),
			newtext.New(""),

			// =====================================================
			// Section 4: Alert Modal
			// =====================================================
			subTitle("4. Alert Modal (Convenience Function)"),
			modal.Alert("Alert", "Important notification\nPlease review this message."),
			newtext.New(""),

			// =====================================================
			// Section 5: Different Sizes
			// =====================================================
			subTitle("5. Different Modal Sizes"),
			newstack.New(newstack.Row).
				SetGap(2).
				SetChildrenList([]rtui.VNode{
					modal.NewBuilder().
						Title("Small").
						Content(newtext.New("Small modal")).
						Size(25, 6).
						Rounded().
						Open(true).
						Build(),
					modal.NewBuilder().
						Title("Medium").
						Content(newtext.New("Medium modal\nMore content\nAvailable.")).
						Size(30, 8).
						Rounded().
						Open(true).
						Build(),
					modal.NewBuilder().
						Title("Large").
						Content(newtext.New("Large modal\nWith more content\nLines for display.\nCan show complex UI.")).
						Size(35, 10).
						Rounded().
						Open(true).
						Build(),
				}),
			newtext.New(""),

			// =====================================================
			// Section 6: Non-Closeable Modal
			// =====================================================
			subTitle("6. Non-Closeable Modal (modal lock)"),
			modal.NewBuilder().
				Title("Important Message").
				Content(newtext.New("This modal cannot be closed.\nUsed for critical alerts.\nMust handle via application logic.")).
				Width(45).
				Height(7).
				Rounded().
				Closeable(false).
				Open(true).
				Build(),
			newtext.New(""),

			// =====================================================
			// Section 7: Dashed Border Style
			// =====================================================
			subTitle("7. Dashed Border Style"),
			modal.NewBuilder().
				Title("Dashed Border").
				Content(newtext.New("Modal with dashed border style.\nAlternative visual design.\nGood for information dialogs.")).
				Width(40).
				Height(8).
				Dashed().
				Open(true).
				Build(),
			newtext.New(""),

			// Footer
			highlight("Modal: Fiber-first, Layer system, Intent events, ESC/click-outside close"),
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
	fmt.Println("║   Fiber-First Modal Component Demo                      ║")
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

	// Create buffer (increase height to fit all content)
	buf := paint.NewBuffer(70, 120)

	// Create paint context
	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 70, Height: 120},
		AvailableWidth:  70,
		AvailableHeight: 120,
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 70))
	fmt.Println("Rendering Modal with Fiber-first pipeline...")
	fmt.Printf("%s\n\n", strings.Repeat("=", 70))

	// Render
	node.Paint(ctx, buf)

	// Output result
	printBuffer(buf, 70, 120)

	// Feature summary
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("Modal Architecture:")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("  ✓ VNode: Pure description (no state, no closures, no paint)")
	fmt.Println("  ✓ Instance: Runtime state management")
	fmt.Println("  ✓ Intent: Replaces closure-based callbacks")
	fmt.Println("  ✓ Layer: Renders on LayerOverlay for z-ordering")
	fmt.Println("  ✓ Events: ESC key and click-outside to close")
	fmt.Println("  ✓ Styles: Single, Double, Rounded, Dashed borders")
	fmt.Println("  ✓ Sizes: Configurable width and height")
	fmt.Println("  ✓ Closeable: Can be set to non-closeable (modal lock)")
	fmt.Println("  ✓ Builder: Fluent API with convenience methods")
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
