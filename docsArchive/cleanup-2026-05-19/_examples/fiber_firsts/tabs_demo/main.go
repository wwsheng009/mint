// Fiber-First Tabs Component Demo
// Demonstrates the new Tabs component following the Fiber-first architecture
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
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/ui/components/tabs"
	newtext "github.com/wwsheng009/mint/ui/components/text"
)

// DemoApp creates the demo UI
func DemoApp() rtui.VNode {
	return ui.NewVStack().
		SetWidth(72).
		SetGap(1).
		SetChildrenList([]rtui.VNode{
			sectionTitle("Fiber-First Tabs Component Demo"),
			newtext.New(""),
			newtext.New("Enhanced tabs now support:"),
			newtext.New("  • controlled selection via ActiveTab / ActiveTabID"),
			newtext.New("  • icons, badges, hotkeys, hidden tabs"),
			newtext.New("  • wrap, custom divider, loop navigation"),
			newtext.New("  • top / bottom / left / right layouts"),
			newtext.New(""),

			subTitle("1. Metadata + controlled selection"),
			tabs.NewBuilder().
				ComponentID("workspace").
				AddTabItem(tabs.Item("home", "Home").WithIcon("H").WithHotkey('h')).
				AddTabItem(tabs.Item("search", "Search").WithIcon("S").WithHotkey('s')).
				AddTabItem(tabs.Item("alerts", "Alerts").WithIcon("!").WithBadge("12").WithHotkey('a')).
				AddTabItem(tabs.Item("locked", "Locked").WithIcon("X").WithDisabled(true)).
				ActiveTabID("alerts").
				ShowHotkeys(true).
				Width(68).
				Top().
				ActiveTabStyle(style.NewStyle().Foreground(style.Cyan).Bold(true)).
				DisabledTabStyle(style.NewStyle().Foreground(style.BrightBlack)).
				Build(),
			newtext.New("  H/S/A choose tabs, disabled tab stays visible but cannot activate.").Foreground("bright-black"),
			newtext.New(""),

			subTitle("2. Wrapped tabs + custom divider"),
			tabs.NewBuilder().
				AddTabItem(tabs.Item("files", "Files").WithHotkey('1')).
				AddTabItem(tabs.Item("outline", "Outline").WithHotkey('2')).
				AddTabItem(tabs.Item("search", "Search").WithHotkey('3')).
				AddTabItem(tabs.Item("source", "Source Control").WithHotkey('4')).
				AddTabItem(tabs.Item("ports", "Ports").WithHotkey('5')).
				AddTabItem(tabs.Item("timeline", "Timeline").WithHotkey('6')).
				ActiveTab(3).
				ShowHotkeys(true).
				WrapTabs(true).
				Divider(" / ").
				Width(32).
				Top().
				Build(),
			newtext.New("  WrapTabs uses local hit areas, so clicking works after wrapping.").Foreground("bright-black"),
			newtext.New(""),

			subTitle("3. Left / right / bottom positions"),
			ui.NewHStack().
				SetGap(2).
				SetChildrenList([]rtui.VNode{
					tabs.NewBuilder().
						AddTab("nav", "Nav").
						AddTab("repo", "Repo").
						AddTab("test", "Test").
						ActiveTab(1).
						Width(18).
						Height(4).
						Left().
						Build(),
					tabs.NewBuilder().
						AddTab("logs", "Logs").
						AddTab("tasks", "Tasks").
						AddTab("debug", "Debug").
						ActiveTab(0).
						Width(18).
						Height(4).
						Bottom().
						Build(),
					tabs.NewBuilder().
						AddTab("git", "Git").
						AddTab("ci", "CI").
						AddTab("perf", "Perf").
						ActiveTab(2).
						Width(18).
						Height(4).
						Right().
						Build(),
				}),
			newtext.New("  Bottom/Right now honor component size instead of always painting at row 0.").Foreground("bright-black"),
			newtext.New(""),

			subTitle("4. Hidden tabs + loop navigation"),
			tabs.NewBuilder().
				AddTabItem(tabs.Item("overview", "Overview").WithIcon("O")).
				AddTabItem(tabs.Item("internals", "Internals").WithIcon("I").WithHidden(true)).
				AddTabItem(tabs.Item("metrics", "Metrics").WithIcon("M").WithBadge("99+")).
				AddTabItem(tabs.Item("trace", "Trace").WithIcon("T")).
				LoopNavigation(true).
				ActiveTabID("metrics").
				Width(68).
				Top().
				Build(),
			newtext.New(""),

			highlight("Tabs: metadata, controlled state, wrapping, intents, and richer navigation"),
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
	node := render.NewDeclarativeNodeFromFuncWithFiber(DemoApp)
	node.SetApp(fwApp)

	// Enable Fiber-first mode
	node.SetRenderMode(render.RenderModeFiberFirst)

	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("  Render Mode: %v\n", node.GetRenderMode())
	fmt.Printf("  Fiber-First Enabled: %v\n", node.IsFiberFirstEnabled())

	// Create buffer
	buf := paint.NewBuffer(72, 34)

	// Create paint context
	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 72, Height: 34},
		AvailableWidth:  72,
		AvailableHeight: 34,
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 70))
	fmt.Println("Rendering Tabs with Fiber-first pipeline...")
	fmt.Printf("%s\n\n", strings.Repeat("=", 70))

	// Render
	node.Paint(ctx, buf)

	// Output result
	utils.PrintBuffer(buf, 72, 34)

	// Feature summary
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("Tabs Architecture:")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("  ✓ VNode: Pure description (no state, no closures, no paint)")
	fmt.Println("  ✓ Instance: Runtime state management (activeTab)")
	fmt.Println("  ✓ Intent: Replaces closure-based callbacks")
	fmt.Println("  ✓ Metadata: Icon, Badge, Hotkey, Hidden, Disabled")
	fmt.Println("  ✓ Positions: Top, Bottom, Left, Right")
	fmt.Println("  ✓ Keyboard: arrows, Home/End, Ctrl+Tab, hotkeys, digits")
	fmt.Println("  ✓ Mouse: local hit testing works with wrapped and offset tabs")
	fmt.Println("  ✓ Styles: separate normal / active / disabled tab styles")
	fmt.Println("  ✓ Builder: controlled active tab, wrap, divider, loop navigation")
	fmt.Println(strings.Repeat("=", 70))
}
