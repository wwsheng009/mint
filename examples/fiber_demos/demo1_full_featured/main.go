// Demo 1: Full-Featured Demo App
//
// This demo demonstrates the complete TUI engine architecture, covering:
// - Declarative components
// - State system (Hooks)
// - Layout system (Flex, VStack, HStack, Table)
// - Modal (Layer) - Using Layer system
// - Input with Focus management
// - Theme system with semantic colors
// - Button variants (Primary, Secondary, Danger, Success)
// - Scroll containers
// - VirtualList for large data
// - Event handling
// - Animation
//
// This is an integration acceptance test for the UI Runtime.
//
// Based on: framework/docs/ui/demo/demo1.md

package main

import (
	"fmt"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
)

// Intent Types
type CloseFiberDemoModalIntent struct{}
func (CloseFiberDemoModalIntent) IntentType() string { return "CloseFiberDemoModal" }
func (CloseFiberDemoModalIntent) StayPressed() bool  { return true }

func main() {
	// ============================================================
	// 直接转换 Fiber 树并比较
	// ============================================================
	fmt.Println("=== Demo1: VNode to Fiber Tree Conversion ===")
	fmt.Println()

	// Ensure default theme is loaded
	_ = theme.SetTheme("nord")

	// Build the demo app's VNode tree
	vnode := App()

	// Convert to Fiber tree
	fiber := ui.CreateFiberFromVNode(vnode)

	fmt.Println("Fiber tree created successfully")
	fmt.Printf("Root Fiber: NodeID=%d, Tag=%s, Type=%v\n",
		fiber.NodeID, fiber.Tag, fiber.Type)
	fmt.Printf("Children count: %d\n", len(fiber.GetChildFibers()))
	fmt.Println()

	// Run comparison
	result := CompareTrees(vnode, fiber)

	// Print results
	PrintComparisonResult(result)

	// Check information preservation
	fmt.Println("=== Fiber Information Preservation ===")
	CheckFiberPreservation(fiber, 0)

	// Print summary
	PrintSummary(result)
}

// App is the root component (static version without hooks)
func App() ui.VNode {
	// Static data instead of hooks
	count := 0
	input := ""
	items := []string{
		"Log line #0", "Log line #1", "Log line #2",
		"Log line #3", "Log line #4", "Log line #5",
	}

	// Build main content
	mainContent := ui.VStackBuilder(
		Header(count),
		MainBody(count, input, items),
		DebugPanel(),
	).Stretch().Build()

	return mainContent
}

// Header demonstrates layout with Bordered component
func Header(count int) ui.VNode {
	headerContent := ui.HStack(
		ui.NewTextBuilder("TUI Engine Demo").
			Style(style.FgBgBold(theme.Text(), theme.Primary())).
			Build(),
		ui.NewTextBuilder("              ").
			Style(style.FgBg(theme.Surface(), theme.Primary())).
			Build(),
		ui.NewTextBuilder(" ").
			Style(style.FgBg(theme.Surface(), theme.Primary())).
			Build(),
		ui.NewTextBuilder(fmt.Sprintf("Clicks: %d", count)).
			Style(style.FgBgBold(theme.BG(), theme.Primary())).
			Build(),
	)

	return ui.Bordered().
		Style(string(theme.Primary())).
		Child(headerContent).
		Build()
}

// MainBody uses VStack/HStack with Bordered components for layout
// Matches the design from framework/docs/ui/demo/demo1.md:
//
//	┌───────────┬──────────────────────────────────────────┐
//	│ Menu      │ [ Input box............................... ] │
//	├───────────┼──────────────────────────────────────────┤
//	│ Add Count │ Log line #0                               │
//	├───────────┼──────────────────────────────────────────┤
//	│ Quit      │ Log line #1                               │
//	├───────────┼──────────────────────────────────────────┤
//	│           │ Log line #2                               │
//	├───────────┼──────────────────────────────────────────┤
//	│           │ Log line #3                               │
//	├───────────┼──────────────────────────────────────────┤
//	│           │ Log line #4                               │
//	├───────────┼──────────────────────────────────────────┤
//	│           │ Log line #5 ...                            │
//	└───────────┴──────────────────────────────────────────┘
func MainBody(count int, input string, items []string) ui.VNode {
	// Left sidebar with menu buttons (no click handlers)
	sidebar := ui.VStackBuilder(
		ui.NewTextBuilder("Menu").
			Style(style.FgBoldUnderline(theme.Muted())).
			Build(),
		ui.NewTextBuilder("Add Count").
			Style(style.FgBold(theme.Primary())).
			Build(),
		ui.NewTextBuilder("Quit").
			Style(style.FgBold(theme.Error())).
			Build(),
	).Stretch().Build()

	// Right content area with input and log lines
	contentArea := ui.VStackBuilder(
		ui.NewTextBuilder("[ Input: " + input + " ]").
			Style(style.Foreground(theme.Text())).
			Build(),
		ui.NewTextBuilder("──────────────────────────────────────").
			Style(style.Foreground(theme.Border())).
			Build(),
		ui.NewTextBuilder(items[0]).
			Style(style.Foreground(theme.Muted())).
			Build(),
		ui.NewTextBuilder(items[1]).
			Style(style.Foreground(theme.Muted())).
			Build(),
		ui.NewTextBuilder(items[2]).
			Style(style.Foreground(theme.Muted())).
			Build(),
		ui.NewTextBuilder(items[3]).
			Style(style.Foreground(theme.Muted())).
			Build(),
		ui.NewTextBuilder(items[4]).
			Style(style.Foreground(theme.Muted())).
			Build(),
		ui.HStack(
			ui.NewTextBuilder(items[5]).
				Style(style.Foreground(theme.Muted())).
				Build(),
			ui.NewTextBuilder(" ...").
				Style(style.FgItalic(theme.Placeholder())).
				Build(),
		),
	).Stretch().Build()

	// Combine sidebar and content with borders
	// Uses theme BORDER color for borders
	return ui.HStackBuilder(
		ui.Flex(
			ui.Bordered().
				Style(string(theme.Border())).
				Child(sidebar).
				Build(),
			1, // Flex factor
		),
		ui.Flex(
			ui.Bordered().
				Style(string(theme.Border())).
				Child(contentArea).
				Build(),
			1, // Flex factor
		),
	).Gap(0).Build()
}

// ConfirmModal demonstrates Layer + Focus Trap with overlay rendering
// Uses the new Layer system for automatic centering and backdrop
// Uses theme colors: WARNING for modal border, SUCCESS for OK button
func ConfirmModal(onClose func()) ui.VNode {
	// Modal content - the actual dialog box with border
	// Uses theme WARNING color for modal border to indicate caution
	modalBox := ui.Bordered().
		Style(string(theme.Warning())).
		Width(40). // Fixed width for the modal
		Child(
			ui.VStackBuilder(
				ui.Text(""),
				// DEBUG: Line number to verify position
				ui.Text("=== MODAL START ==="),
				// Centered title - use HStack with AlignCenter
				// Uses theme WARNING color for title
				ui.HStackBuilder(
					ui.NewTextBuilder("*** Are you sure? ***").
						Style(style.FgBold(theme.Warning())).
						Build(),
				).Align(ui.AlignCenter).Build(),
				ui.Text(""),
				// Centered buttons - use HStack with AlignCenter
				// Uses theme colors: Secondary for Cancel, Success for OK
				ui.HStackBuilder(
					app.ButtonBuilder("[ Cancel ]").
						Variant(app.ButtonVariantSecondary).
						OnPress(CloseFiberDemoModalIntent{}).
						FocusStyle(app.FocusStyleBracket).
						Build(),
					ui.Text(" "),
					app.ButtonBuilder("[ OK ]").
						Variant(app.ButtonVariantSuccess).
						OnPress(CloseFiberDemoModalIntent{}).
						FocusStyle(app.FocusStyleBracket).
						Build(),
				).Align(ui.AlignCenter).Build(),
				ui.Text(""),
				// Centered footer text
				// Uses theme PLACEHOLDER color for hint text
				ui.HStackBuilder(
					ui.NewTextBuilder("Press ESC to close").
						Style(style.Foreground(theme.Placeholder())).
						Build(),
				).Align(ui.AlignCenter).Build(),
				ui.Text(""),
				// DEBUG: End marker
				ui.Text("=== MODAL END ==="),
			).Build(),
		).
		Build()

	// Register handler for CloseFiberDemoModalIntent (only when actually running)
	// Note: This is a Fiber conversion test, not a real TUI app
	ui.On(CloseFiberDemoModalIntent{}, func() {
		if onClose != nil {
			onClose()
		}
	})

	return ui.Modal(modalBox).
		CloseOnESC(true).
		CloseOnBackdropClick(true).
		Build()
}

// =============================================================================
// Debug Helper Functions
// =============================================================================

// DebugPanel 显示屏幕配置和调试信息
func DebugPanel() ui.VNode {
	infoLines := []string{
		"┌─ SCREEN/INFO PANEL ─────────────────────────────────────────────┐",
		"│ Buffer Size: 80x24 (configured via ui.WithWidth/Height)        │",
		"│ Debug Log: demo1_debug.log (check for HitTest details)         │",
		"│                                                                │",
		"│ MODAL BUTTON HITEST VERIFICATION:                              │",
		"│ 1. Modal opens automatically (AUTO_OPEN_MODAL=true)            │",
		"│ 2. Click modal buttons - they increment the counter           │",
		"│ 3. Check demo1_debug.log for:                                  │",
		"│    - Mouse position (X, Y)                                     │",
		"│    - HitTest results (button bounds)                           │",
		"│    - Multiple button overlap detection                         │",
		"│    - Modal centering calculations                              │",
		"│                                                                │",
		"│ EXPECTED BEHAVIOR:                                             │",
		"│ - Modal centered in buffer: Y position depends on buffer size│",
		"│ - If actual terminal > 24 lines, check logs for actual size   │",
		"└────────────────────────────────────────────────────────────────┘",
	}

	return ui.VStackBuilder(
		ui.Text(infoLines[0]),
		ui.Text(infoLines[1]),
		ui.Text(infoLines[2]),
		ui.Text(infoLines[3]),
		ui.Text(infoLines[4]),
		ui.Text(infoLines[5]),
		ui.Text(infoLines[6]),
		ui.Text(infoLines[7]),
		ui.Text(infoLines[8]),
		ui.Text(infoLines[9]),
		ui.Text(infoLines[10]),
		ui.Text(infoLines[11]),
		ui.Text(infoLines[12]),
		ui.Text(infoLines[13]),
	).
		Build()
}
