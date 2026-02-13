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
	"os"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
)

func main() {
	// Ensure default theme is loaded
	_ = theme.SetTheme("nord")

	// ============================================================
	// 调试环境变量 - 用于 HitTest 和 Modal 点击测试
	// ============================================================
	// 启用文件日志（记录 HitTest、鼠标位置、Modal 居中等）
	os.Setenv("TUI_DEBUG_LOG", "demo1_debug.log")

	// 自动打开 Modal，方便直接测试按钮点击
	os.Setenv("AUTO_OPEN_MODAL", "true")

	// 启用以下环境变量可获取更详细的调试信息：
	os.Setenv("TUI_DEBUG_HITMAP", "true")   // HitMap 构建详情
	// os.Setenv("TUI_DEBUG_LAYER", "true")    // Layer 系统调试
	// os.Setenv("TUI_DEBUG_RENDER", "true") // 渲染管线调试
	os.Setenv("TUI_DEBUG_UI", "true")       // UI 通用调试

	// ============================================================
	// 运行应用
	// ============================================================
	err := ui.Run(App,
		ui.WithWidth(80),
		ui.WithHeight(24),
		ui.WithTitle("Mint TUI - Full Featured Demo"),
	)
	if err != nil {
		panic(err)
	}

	// 运行结束后提示日志位置
	log.UILogger.Debug("=== Debug session ended ===")
	log.UILogger.Debug("Log file: demo1_debug.log")
	log.UILogger.Debug("Check for:")
	log.UILogger.Debug("  - [MOUSE] mouse position and HitTest results")
	log.UILogger.Debug("  - [LAYER] modal centering and position")
	log.UILogger.Debug("  - [HITMAP] button bounds entries")
}

// App is the root component
func App() ui.VNode {
	count, setCount, _ := ui.UseStateInt(0)

	// 环境变量控制：默认打开 modal 用于调试
	autoOpenModal := os.Getenv("AUTO_OPEN_MODAL") == "true"
	showModal, setShowModal := ui.UseStateBool(autoOpenModal)

	input, setInput := ui.UseStateString("")

	// Generate large list for VirtualList
	items := make([]string, 100)
	for i := range items {
		items[i] = fmt.Sprintf("Log line #%d", i)
	}

	// NEW: Render both main content AND modal (when open)
	// The Layer system handles proper z-ordering and centering
	mainContent := ui.VStackBuilder(
		Header(count, setShowModal, setCount),
		MainBody(count, setCount, input, setInput, items),
		DebugPanel(),
	).Stretch().Build()

	// If modal is open, render both main content and modal
	// The LayerManager will separate them into different layers
	if showModal {
		modalVNode := ConfirmModal(func() {
			setShowModal(false)
		})

		result := ui.VStack(
			mainContent,
			// Modal layer - automatically centered and overlays main content
			modalVNode,
		)

		return result
	}

	// Otherwise render just main content
	return mainContent
}

// Header demonstrates state + layout with Bordered component
// Uses theme colors: PRIMARY for header background, TEXT for text
func Header(count int, setShowModal func(bool), setCount func(interface{})) ui.VNode {
	headerContent := ui.HStack(
		app.NewTextBuilder("TUI Engine Demo").
			Style(style.FgBgBold(theme.Text(), theme.Primary())).
			Build(),
		app.NewTextBuilder("              ").
			Style(style.FgBg(theme.Surface(), theme.Primary())).
			Build(),
		app.ButtonBuilder("[Open Modal]").
			Variant(app.ButtonVariantPrimary). // 使用 Primary variant，默认就有 PRIMARY 背景
			OnClick(func() {
				setShowModal(true)
			}).
			FocusStyle(app.FocusStyleBracket). // 恢复 Bracket 样式
			Build(),
		app.NewTextBuilder(" ").
			Style(style.FgBg(theme.Surface(), theme.Primary())).
			Build(),
		app.NewTextBuilder(fmt.Sprintf("Clicks: %d", count)).
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
func MainBody(count int, setCount func(interface{}), input string, setInput func(string), items []string) ui.VNode {
	// Left sidebar with menu buttons
	// Uses theme colors: MUTED for menu label, Primary variant for Add Count, Danger variant for Quit
	sidebar := ui.VStackBuilder(
		app.NewTextBuilder("Menu").
			Style(style.FgBoldUnderline(theme.Muted())).
			Build(),
		app.ButtonBuilder("Add Count").
			Variant(app.ButtonVariantPrimary).
			OnClick(func() {
				setCount(func(c int) int { return c + 1 })
			}).
			FocusStyle(app.FocusStyleBracket).
			Build(),
		app.ButtonBuilder("Quit").
			Variant(app.ButtonVariantDanger).
			FocusStyle(app.FocusStyleBracket).
			OnClick(func() {
				ui.Quit()
			}).
			Build(),
	).Stretch().Build()

	// Right content area with input and log lines
	// Uses theme colors: TEXT for labels, MUTED for log lines, BORDER for divider
	contentArea := ui.VStackBuilder(
		app.InputBuilder().
			Value(input).
			Placeholder("Type something...").
			Width(30). // Input width (less than panel width)
			OnChange(setInput).
			Build(),
		app.NewTextBuilder("──────────────────────────────────────").
			Style(style.Foreground(theme.Border())).
			Build(),
		app.NewTextBuilder(items[0]).
			Style(style.Foreground(theme.Muted())).
			Build(),
		app.NewTextBuilder(items[1]).
			Style(style.Foreground(theme.Muted())).
			Build(),
		app.NewTextBuilder(items[2]).
			Style(style.Foreground(theme.Muted())).
			Build(),
		app.NewTextBuilder(items[3]).
			Style(style.Foreground(theme.Muted())).
			Build(),
		app.NewTextBuilder(items[4]).
			Style(style.Foreground(theme.Muted())).
			Build(),
		ui.HStack(
			app.NewTextBuilder(items[5]).
				Style(style.Foreground(theme.Muted())).
				Build(),
			app.NewTextBuilder(" ...").
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
					app.NewTextBuilder("*** Are you sure? ***").
						Style(style.FgBold(theme.Warning())).
						Build(),
				).Align(ui.AlignCenter).Build(),
				ui.Text(""),
				// Centered buttons - use HStack with AlignCenter
				// Uses theme colors: Secondary for Cancel, Success for OK
				ui.HStackBuilder(
					app.ButtonBuilder("[ Cancel ]").
						Variant(app.ButtonVariantSecondary).
						OnClick(onClose).
						FocusStyle(app.FocusStyleBracket).
						Build(),
					ui.Text(" "),
					app.ButtonBuilder("[ OK ]").
						Variant(app.ButtonVariantSuccess).
						FocusStyle(app.FocusStyleBracket).
						OnClick(onClose).
						Build(),
				).Align(ui.AlignCenter).Build(),
				ui.Text(""),
				// Centered footer text
				// Uses theme PLACEHOLDER color for hint text
				ui.HStackBuilder(
					app.NewTextBuilder("Press ESC to close").
						Style(style.Foreground(theme.Placeholder())).
						Build(),
				).Align(ui.AlignCenter).Build(),
				ui.Text(""),
				// DEBUG: End marker
				ui.Text("=== MODAL END ==="),
			).Build(),
		).
		Build()

	return ui.Modal(modalBox).
		OnClose(onClose).
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
