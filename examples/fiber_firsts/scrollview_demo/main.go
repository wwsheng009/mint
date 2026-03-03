// Package main demonstrates the ScrollView component following the Fiber-first architecture.
// Shows various scroll features including auto-size, borders, indicators, and nested scrolling.
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
	"github.com/wwsheng009/mint/ui/components/scrollview"
	newtext "github.com/wwsheng009/mint/ui/components/text"
)

// buildLines creates n lines of text
func buildLines(n int) string {
	lines := make([]string, n)
	for i := 0; i < n; i++ {
		lines[i] = fmt.Sprintf("Line %3d: This is scrollable content", i+1)
	}
	return strings.Join(lines, "\n")
}

// buildLogLines creates log-style lines
func buildLogLines(n int) string {
	var lines []string
	levels := []string{"INFO", "DEBUG", "WARN", "ERROR"}
	for i := 0; i < n; i++ {
		level := levels[i%len(levels)]
		timestamp := fmt.Sprintf("10:%02d:%02d", i/60, i%60)
		lines = append(lines, fmt.Sprintf("[%s] [%s] Processing request #%d - status: OK", timestamp, level, i+1))
	}
	return strings.Join(lines, "\n")
}

// buildCodeSample creates a code sample
func buildCodeSample() string {
	return `package main

import "fmt"

func main() {
    // Hello World
    fmt.Println("Hello, World!")

    // Fibonacci
    for i := 0; i < 10; i++ {
        fmt.Printf("fib(%d) = %d\n", i, fib(i))
    }
}

func fib(n int) int {
    if n < 2 {
        return n
    }
    return fib(n-1) + fib(n-2)
}`
}

// buildMenuItems creates menu items
func buildMenuItems() string {
	items := []string{
		"File",
		"  New File",
		"  Open...",
		"  Save",
		"  Exit",
		"",
		"Edit",
		"  Undo",
		"  Redo",
		"  Cut",
		"  Copy",
		"  Paste",
		"",
		"View",
		"  Zoom In",
		"  Zoom Out",
		"",
		"Settings",
		"  Preferences",
		"  Theme",
	}
	return strings.Join(items, "\n")
}

// buildFileList creates a file list
func buildFileList() string {
	files := []string{
		"README.md",
		"main.go",
		"go.mod",
		"go.sum",
		"components/",
		"  button.go",
		"  input.go",
		"  scrollview.go",
		"  text.go",
		"runtime/",
		"  layout/",
		"  paint/",
		"examples/",
		"  scrollview_demo/",
		"internal/",
		"docs/",
		"Makefile",
		".gitignore",
	}
	return strings.Join(files, "\n")
}

// DemoApp renders ScrollView components using the Fiber-first pipeline
func DemoApp() rtui.VNode {
	return ui.NewVStack().
		SetWidth(70).
		SetGap(1).
		SetChildrenList([]rtui.VNode{
			// Title
			sectionTitle("════════════════ ScrollView Demo ════════════════"),

			// =====================================================
			// Section 1: Basic Scroll Views
			// =====================================================
			subTitle("1. Basic Scroll (width=35, height=5, offset=3)"),
			scrollview.NewBuilder().
				Child(newtext.New(buildLines(20))).
				Width(35).
				Height(5).
				ScrollOffset(3).
				ShowBorder(true).
				Build(),

			// =====================================================
			// Section 2: Log Viewer
			// =====================================================
			newtext.New(""),
			subTitle("2. Log Viewer (30 lines, height=6, offset=20)"),
			scrollview.NewBuilder().
				Child(newtext.New(buildLogLines(30))).
				Width(65).
				Height(6).
				ScrollOffset(20).
				ShowBorder(true).
				Build(),

			// =====================================================
			// Section 3: Code Viewer
			// =====================================================
			newtext.New(""),
			subTitle("3. Code Viewer (width=50, height=8)"),
			scrollview.NewBuilder().
				Child(newtext.New(buildCodeSample())).
				Width(50).
				Height(8).
				ScrollOffset(0).
				ShowBorder(true).
				Build(),

			// =====================================================
			// Section 4: File Explorer
			// =====================================================
			newtext.New(""),
			subTitle("4. File Explorer (width=28, height=7)"),
			scrollview.NewBuilder().
				Child(newtext.New(buildFileList())).
				Width(28).
				Height(7).
				ScrollOffset(0).
				ShowBorder(true).
				Build(),

			// =====================================================
			// Section 5: Menu List
			// =====================================================
			newtext.New(""),
			subTitle("5. Menu List (width=22, height=6, offset=5)"),
			scrollview.NewBuilder().
				Child(newtext.New(buildMenuItems())).
				Width(22).
				Height(6).
				ScrollOffset(5).
				ShowBorder(true).
				Build(),

			// =====================================================
			// Section 6: Side by Side - Scroll Indicators
			// =====================================================
			newtext.New(""),
			subTitle("6. Scroll Position Indicators (Top/Middle/Bottom)"),

			ui.NewHStack().
				SetGap(2).
				SetChildrenList([]rtui.VNode{
					scrollview.NewBuilder().
						Child(newtext.New(buildLines(10))).
						Width(18).
						Height(4).
						ScrollOffset(0).
						ShowBorder(true).
						Build(),

					scrollview.NewBuilder().
						Child(newtext.New(buildLines(10))).
						Width(18).
						Height(4).
						ScrollOffset(3).
						ShowBorder(true).
						Build(),

					scrollview.NewBuilder().
						Child(newtext.New(buildLines(10))).
						Width(18).
						Height(4).
						ScrollOffset(6).
						ShowBorder(true).
						Build(),
				}),

			// =====================================================
			// Section 7: Auto-Height Mode
			// =====================================================
			newtext.New(""),
			subTitle("7. Auto-Height Mode (height=0, shows all content)"),
			scrollview.NewBuilder().
				Child(newtext.New("Line 1\nLine 2\nLine 3")).
				Width(30).
				ShowBorder(true).
				Build(),

			// =====================================================
			// Section 8: Without Scroll Indicator
			// =====================================================
			newtext.New(""),
			subTitle("8. Without Scroll Indicator"),
			scrollview.NewBuilder().
				Child(newtext.New(buildLines(15))).
				Width(40).
				Height(5).
				ShowBorder(true).
				ShowIndicator(false).
				Build(),

			// Footer
			newtext.New(""),
			highlight("ScrollView: vertical scroll, page up/down, scroll indicators, auto-height"),
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
	fmt.Println("║   Fiber-First ScrollView Rendering Demo                          ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════╝")

	// Create framework app (required for Fiber reconciler)
	fwApp := framework.NewApp()

	// Create DeclarativeNode WITH Fiber reconciler
	node := render.NewDeclarativeNodeFromFuncWithFiber(DemoApp)
    node.SetApp(fwApp)

	// Enable Fiber-first mode
	node.SetRenderMode(render.RenderModeFiberFirst)

	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("  Render Mode: %v\n", node.GetRenderMode())

	// Create buffer (70 wide, 100 tall)
	buf := paint.NewBuffer(70, 100)

	// Create paint context
	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 70, Height: 100},
		AvailableWidth:  70,
		AvailableHeight: 100,
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 70))
	fmt.Println("Rendering ScrollView components...")
	fmt.Printf("%s\n\n", strings.Repeat("=", 70))

	// Render
	node.Paint(ctx, buf)

	// Output result
	utils.PrintBuffer(buf, 70, 100)

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("ScrollView Component Features:")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("  Scroll Controls:")
	fmt.Println("    - ScrollBy(delta): Scroll by n lines")
	fmt.Println("    - ScrollTo(offset): Scroll to absolute position")
	fmt.Println("    - PageUp/PageDown: Scroll by viewport height")
	fmt.Println("    - ScrollTop/ScrollBottom: Jump to start/end")
	fmt.Println("")
	fmt.Println("  Display Options:")
	fmt.Println("    - Width/Height: Fixed viewport size")
	fmt.Println("    - Auto-height: height=0 shows all content")
	fmt.Println("    - ShowBorder: Display border around viewport")
	fmt.Println("    - ShowIndicator: Show scroll position (↓↕↑)")
	fmt.Println("")
	fmt.Println("  Scroll Indicators:")
	fmt.Println("    ↓ : At top, can scroll down")
	fmt.Println("    ↕ : In middle, can scroll both ways")
	fmt.Println("    ↑ : At bottom, can scroll up")
	fmt.Println("")
	fmt.Println("  Actions Supported:")
	fmt.Println("    - ActionScroll")
	fmt.Println("    - ActionNavigateUp/Down")
	fmt.Println("    - ActionNavigatePageUp/PageDown")
	fmt.Println("    - ActionNavigateHome/End")
	fmt.Println(strings.Repeat("=", 70))
}
