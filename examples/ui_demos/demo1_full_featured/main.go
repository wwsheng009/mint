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

func main() {
	// Ensure default theme is loaded
	_ = theme.SetTheme("nord")

	err := ui.Run(App,
		ui.WithWidth(80),
		ui.WithHeight(24),
		ui.WithTitle("Mint TUI - Full Featured Demo"),
	)
	if err != nil {
		panic(err)
	}
}

// App is the root component
func App() ui.VNode {
	count, setCount, _ := ui.UseStateInt(0)
	showModal, setShowModal := ui.UseStateBool(false)
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
	).Stretch().Build()

	// If modal is open, render both main content and modal
	// The LayerManager will separate them into different layers
	if showModal {
		return ui.VStack(
			mainContent,
			// Modal layer - automatically centered and overlays main content
			ConfirmModal(func() {
				setShowModal(false)
			}),
		)
	}

	// Otherwise render just main content
	return mainContent
}

// Header demonstrates state + layout with Bordered component
// Uses theme colors: PRIMARY for header background, TEXT for text
func Header(count int, setShowModal func(bool), setCount func(interface{})) ui.VNode {
	headerContent := ui.HStack(
		app.NewTextBuilder("TUI Engine Demo").
			Style(style.Style{}.Foreground(theme.Text()).Background(theme.Primary()).Bold(true)).
			Build(),
		app.NewTextBuilder("              ").
			Style(style.Style{}.Foreground(theme.Surface()).Background(theme.Primary())).
			Build(),
		app.ButtonBuilder("[Open Modal]").
			Variant(app.ButtonVariantPrimary). // 使用 Primary variant，默认就有 PRIMARY 背景
			OnClick(func() {
				setShowModal(true)
			}).
			FocusStyle(app.FocusStyleBracket). // 恢复 Bracket 样式
			Build(),
		app.NewTextBuilder(" ").
			Style(style.Style{}.Foreground(theme.Surface()).Background(theme.Primary())).
			Build(),
		app.NewTextBuilder(fmt.Sprintf("Clicks: %d", count)).
			Style(style.Style{}.Foreground(theme.BG()).Background(theme.Primary()).Bold(true)).
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
			Style(style.Style{}.Foreground(theme.Muted()).Bold(true).Underline(true)).
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
			Width(30). // Set explicit width to match divider
			OnChange(setInput).
			Build(),
		app.NewTextBuilder("──────────────────────────────").
			Style(style.Style{}.Foreground(theme.Border())).
			Build(),
		app.NewTextBuilder(items[0]).
			Style(style.Style{}.Foreground(theme.Muted())).
			Build(),
		app.NewTextBuilder(items[1]).
			Style(style.Style{}.Foreground(theme.Muted())).
			Build(),
		app.NewTextBuilder(items[2]).
			Style(style.Style{}.Foreground(theme.Muted())).
			Build(),
		app.NewTextBuilder(items[3]).
			Style(style.Style{}.Foreground(theme.Muted())).
			Build(),
		app.NewTextBuilder(items[4]).
			Style(style.Style{}.Foreground(theme.Muted())).
			Build(),
		ui.HStack(
			app.NewTextBuilder(items[5]).
				Style(style.Style{}.Foreground(theme.Muted())).
				Build(),
			app.NewTextBuilder(" ...").
				Style(style.Style{}.Foreground(theme.Placeholder()).Italic(true)).
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
				// Centered title - use HStack with AlignCenter
				// Uses theme WARNING color for title
				ui.HStackBuilder(
					app.NewTextBuilder("*** Are you sure? ***").
						Style(style.Style{}.Foreground(theme.Warning()).Bold(true)).
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
						Style(style.Style{}.Foreground(theme.Placeholder())).
						Build(),
				).Align(ui.AlignCenter).Build(),
				ui.Text(""),
			).Build(),
		).
		Build()

	return ui.Modal(modalBox).
		OnClose(onClose).
		CloseOnESC(true).
		CloseOnBackdropClick(true).
		Build()
}
