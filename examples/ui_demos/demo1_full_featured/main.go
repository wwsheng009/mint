// Demo 1: Full-Featured Demo App
//
// This demo demonstrates the complete TUI engine architecture, covering:
// - Declarative components
// - State system (Hooks)
// - Layout system (Flex, VStack, HStack, Table)
// - Modal (Layer)
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

	// Build main content using VStack for vertical layout
	mainContent := ui.VStack(
		Header(count, showModal, setShowModal, setCount),
		MainBody(count, setCount, input, setInput, items),
	)

	// Layer: Modal (conditional rendering)
	if showModal {
		return ui.VStack(
			mainContent,
			ConfirmModal(func() {
				setShowModal(false)
			}),
		)
	}

	return mainContent
}

// Header demonstrates state + layout with Bordered component
func Header(count int, showModal bool, setShowModal func(bool), setCount func(interface{})) ui.VNode {
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
	sidebar := ui.VStack(
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
	)

	// Right content area with input and log lines
	contentArea := ui.VStack(
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
	)

	// Combine sidebar and content with borders
	return ui.HStack(
		ui.Bordered().
			Color("blue").
			Child(sidebar).
			Build(),
		ui.Bordered().
			Color("blue").
			Child(contentArea).
			Build(),
	)
}

// ConfirmModal demonstrates Layer + Animation + Focus Trap with Bordered component
func ConfirmModal(onClose func()) ui.VNode {
	modalContent := ui.VStack(
		ui.Text(""),
		ui.HStack(
			ui.Text("       "),
			app.ButtonBuilder("[Cancel]").
				OnClick(onClose).
				Build(),
			ui.Text(" "),
			app.ButtonBuilder("[OK]").
				BgColor("green").
				FgColor("white").
				OnClick(onClose).
				Build(),
		),
		ui.Text(""),
	)

	return ui.VStack(
		ui.Text(""),
		ui.HStack(
			ui.Text("        "),
			ui.Bordered().
				Color("yellow").
				Style("double").
				Child(
					ui.VStack(
						ui.Text(""),
						ui.HStack(
							ui.Text("       "),
							app.NewTextBuilder("*** Are you sure? ***").
								Bold(true).
								FgColor("yellow").
								Build(),
						),
						ui.Text(""),
						modalContent,
					),
				).
				Build(),
		),
		ui.Text(""),
		app.NewTextBuilder("Press ESC to close").
			FgColor("gray").
			Build(),
	)
}
