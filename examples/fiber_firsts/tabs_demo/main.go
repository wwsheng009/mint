// Fiber-First Tabs Component Demo
// Demonstrates the new Tabs component following the Fiber-first architecture
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
	"github.com/wwsheng009/mint/ui/components/tabs"
)

// DemoApp creates the demo UI
func DemoApp() rtui.VNode {
	return newstack.New(newstack.Column).
		SetWidth(70).
		SetGap(1).
		SetChildrenList([]rtui.VNode{
			// Title
			sectionTitle("Fiber-First Tabs Component Demo"),
			newtext.New(""),
			newtext.New("Tabs navigation with Fiber-first architecture:"),
			newtext.New("  • Pure descriptive VNode"),
			newtext.New("  • Intent-based event handling"),
			newtext.New("  • Multiple tab positions (Top/Bottom/Left/Right)"),
			newtext.New("  • Keyboard navigation (arrows, Home, End)"),
			newtext.New("  • Click-to-switch tabs"),
			newtext.New("  • Disabled tabs support"),
			newtext.New(""),

			// =====================================================
			// Section 1: Basic Tabs
			// =====================================================
			subTitle("1. Basic Tabs (Top Position)"),
			tabs.NewBuilder().
				AddTab("home", "Home").
				AddTab("about", "About").
				AddTab("contact", "Contact").
				AddTab("settings", "Settings").
				Width(60).
				Height(10).
				Top().
				Build(),
			newtext.New(""),

			// =====================================================
			// Section 2: Tabs with Disabled Tab
			// =====================================================
			subTitle("2. Tabs with Disabled Tab"),
			tabs.NewBuilder().
				AddTab("active1", "Active").
				AddTabWithOptions("disabled", "Disabled", true).
				AddTab("active2", "Active").
				Width(50).
				Height(8).
				Top().
				Build(),
			newtext.New(""),

			// =====================================================
			// Section 3: Tab Positions
			// =====================================================
			subTitle("3. Different Tab Positions"),
			newstack.New(newstack.Row).
				SetGap(2).
				SetChildrenList([]rtui.VNode{
					tabs.NewBuilder().
						AddTab("t1", "Top").
						AddTab("t2", "Pos").
						Width(25).
						Height(6).
						Top().
						Build(),
					tabs.NewBuilder().
						AddTab("b1", "Btm").
						AddTab("b2", "Pos").
						Width(25).
						Height(6).
						Bottom().
						Build(),
				}),
			newtext.New(""),

			// =====================================================
			// Section 4: Multiple Tabs
			// =====================================================
			subTitle("4. Many Tabs"),
			tabs.NewBuilder().
				AddTab("tab1", "File").
				AddTab("tab2", "Edit").
				AddTab("tab3", "View").
				AddTab("tab4", "Go").
				AddTab("tab5", "Run").
				AddTab("tab6", "Tools").
				AddTab("tab7", "Help").
				AddTab("tab8", "Debug").
				Width(65).
				Height(8).
				Top().
				Build(),
			newtext.New(""),

			// =====================================================
			// Section 5: Compact Tabs
			// =====================================================
			subTitle("5. Compact Short Name Tabs"),
			tabs.NewBuilder().
				AddTab("h", "H").
				AddTab("a", "A").
				AddTab("c", "C").
				Width(30).
				Height(5).
				Top().
				Build(),
			newtext.New(""),

			// Footer
			highlight("Tabs: Navigation, keyboard support, disabled states, flexible positioning"),
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
	fmt.Println("║   Fiber-First Tabs Component Demo                       ║")
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
	buf := paint.NewBuffer(70, 40)

	// Create paint context
	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 70, Height: 40},
		AvailableWidth:  70,
		AvailableHeight: 40,
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 70))
	fmt.Println("Rendering Tabs with Fiber-first pipeline...")
	fmt.Printf("%s\n\n", strings.Repeat("=", 70))

	// Render
	node.Paint(ctx, buf)

	// Output result
	printBuffer(buf, 70, 40)

	// Feature summary
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("Tabs Architecture:")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("  ✓ VNode: Pure description (no state, no closures, no paint)")
	fmt.Println("  ✓ Instance: Runtime state management (activeTab)")
	fmt.Println("  ✓ Intent: Replaces closure-based callbacks")
	fmt.Println("  ✓ Positions: Top, Bottom, Left, Right")
	fmt.Println("  ✓ Keyboard: Arrow keys, Home, End for navigation")
	fmt.Println("  ✓ Mouse: Click tabs to switch")
	fmt.Println("  ✓ Disabled: Support for disabled tabs")
	fmt.Println("  ✓ Styles: Separate styles for normal and active tabs")
	fmt.Println("  ✓ Builder: Fluent API with AddTab, Position, Width, Height")
	fmt.Println(strings.Repeat("=", 70))
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
