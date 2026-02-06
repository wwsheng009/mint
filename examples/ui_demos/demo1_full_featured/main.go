// Demo 1: Full-Featured Demo App
//
// This demo demonstrates the complete TUI engine architecture, covering:
// - Declarative components
// - State system (Hooks)
// - Layout system (Flex, VStack, HStack, Table)
// - Modal (Layer) - NEW: Using Layer system
// - Input with Focus management
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
	"github.com/wwsheng009/mint/ui"
)

func main() {
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
func Header(count int, setShowModal func(bool), setCount func(interface{})) ui.VNode {
	headerContent := ui.HStack(
		app.NewTextBuilder("TUI Engine Demo").
			Bold(true).
			FgColor("white").
			BgColor("blue").
			Build(),
		app.NewTextBuilder("              ").
			BgColor("blue").
			Build(),
		app.ButtonBuilder("[Open Modal]").
			OnClick(func() {
				setShowModal(true)
			}).
			Build(),
		app.NewTextBuilder(" ").
			BgColor("blue").
			Build(),
		app.NewTextBuilder(fmt.Sprintf("Clicks: %d", count)).
			BgColor("blue").
			FgColor("yellow").
			Build(),
	)

	return ui.Bordered().
		Color("blue").
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
	sidebar := ui.VStackBuilder(
		app.NewTextBuilder("Menu").
			Bold(true).
			Underline(true).
			Build(),
		app.ButtonBuilder("Add Count").
			OnClick(func() {
				setCount(func(c int) int { return c + 1 })
			}).
			Build(),
		app.ButtonBuilder("Quit").
			BgColor("red").
			FgColor("white").
			OnClick(func() {
				ui.Quit()
			}).
			Build(),
	).Stretch().Build()

	// Right content area with input and log lines
	// Use VStackBuilder with Stretch to make all items fill the width
	contentArea := ui.VStackBuilder(
		app.InputBuilder().
			Value(input).
			Placeholder("Type something...").
			Width(30). // Set explicit width to match divider
			OnChange(setInput).
			Build(),
		app.NewTextBuilder("──────────────────────────────").
			FgColor("blue").
			Build(),
		app.NewTextBuilder(items[0]).
			FgColor("gray").
			Build(),
		app.NewTextBuilder(items[1]).
			FgColor("gray").
			Build(),
		app.NewTextBuilder(items[2]).
			FgColor("gray").
			Build(),
		app.NewTextBuilder(items[3]).
			FgColor("gray").
			Build(),
		app.NewTextBuilder(items[4]).
			FgColor("gray").
			Build(),
		ui.HStack(
			app.NewTextBuilder(items[5]).
				FgColor("gray").
				Build(),
			app.NewTextBuilder(" ...").
				FgColor("dark-gray").
				Italic(true).
				Build(),
		),
	).Stretch().Build()

	// Combine sidebar and content with borders
	// Apply Flex to both Bordered nodes so they stretch to match heights
	// Use gap=0 so they fill the full width evenly
	return ui.HStackBuilder(
		ui.Flex(
			ui.Bordered().
				Color("blue").
				Child(sidebar).
				Build(),
			1, // Flex factor
		),
		ui.Flex(
			ui.Bordered().
				Color("blue").
				Child(contentArea).
				Build(),
			1, // Flex factor
		),
	).Gap(0).Build()
}

// ConfirmModal demonstrates Layer + Focus Trap with overlay rendering
// Uses the new Layer system for automatic centering and backdrop
func ConfirmModal(onClose func()) ui.VNode {
	// Modal content - the actual dialog box with border
	// Fixed size modal for consistent centering
	modalBox := ui.Bordered().
		Color("yellow").
		Width(40).  // Fixed width for the modal
		Child(
			ui.VStackBuilder(
				ui.Text(""),
				// Centered title
				ui.HStack(
					ui.Text(""),
					app.NewTextBuilder("*** Are you sure? ***").
						Bold(true).
						FgColor("yellow").
						Build(),
					ui.Text(""),
				),
				ui.Text(""),
				// Centered buttons
				ui.HStack(
					ui.Text(""),
					app.ButtonBuilder("[ Cancel ]").
						OnClick(onClose).
						Build(),
					ui.Text(" "),
					app.ButtonBuilder("[ OK ]").
						BgColor("green").
						FgColor("white").
						OnClick(onClose).
						Build(),
					ui.Text(""),
				),
				ui.Text(""),
				// Centered footer text
				ui.HStack(
					ui.Text(""),
					app.NewTextBuilder("Press ESC to close").
						FgColor("gray").
						Build(),
					ui.Text(""),
				),
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
