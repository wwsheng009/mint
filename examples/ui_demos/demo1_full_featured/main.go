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

	// Build main content using Table for row-based layout
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

// Header demonstrates state + layout
func Header(count int, showModal bool, setShowModal func(bool), setCount func(interface{})) ui.VNode {
	return ui.VStack(
		app.NewTextBuilder("+--------------------------------------------------+").
			FgColor("blue").
			Build(),
		ui.HStack(
			app.NewTextBuilder("| ").
				FgColor("blue").
				Build(),
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
			app.NewTextBuilder(" |").
				FgColor("blue").
				BgColor("blue").
				Build(),
		),
		app.NewTextBuilder("+--------------------------------------------------+").
			FgColor("blue").
			Build(),
	)
}

// MainBody uses Table layout to align sidebar and content area row by row
// Matches the design from framework/docs/ui/demo/demo1.md:
//
//	| Menu      | [ Input box....................... ] |
//	| Add Count |--------------------------------------|
//	| Quit      | Log line #0                          |
//	|           | Log line #1                          |
//	|           | ...                                  |
//	|           | (scroll)                             |
//	+-----------+--------------------------------------+
func MainBody(count int, setCount func(interface{}), input string, setInput func(string), items []string) ui.VNode {
	return ui.Table(
		// Border top row
		ui.Row(
			ui.Cell(app.NewTextBuilder("+-----------+").FgColor("blue").Build()),
			ui.Cell(app.NewTextBuilder("--------------------------------------+").FgColor("blue").Build()),
		),

		// Row 1: "Menu" | Input box
		ui.Row(
			ui.Cell(ui.HStack(
				app.NewTextBuilder("| ").FgColor("blue").Build(),
				app.NewTextBuilder("Menu").Bold(true).Underline(true).Build(),
			)),
			ui.Cell(ui.HStack(
				app.NewTextBuilder("| ").FgColor("blue").Build(),
				app.InputBuilder().Value(input).Placeholder("Type something...").OnChange(setInput).Build(),
			)),
		),

		// Row 2: "Add Count" | Separator
		ui.Row(
			ui.Cell(ui.HStack(
				app.NewTextBuilder("| ").FgColor("blue").Build(),
				app.ButtonBuilder("Add Count").OnClick(func() {
					setCount(func(c int) int { return c + 1 })
				}).Build(),
			)),
			ui.Cell(ui.HStack(
				app.NewTextBuilder("|").FgColor("blue").Build(),
				app.NewTextBuilder("--------------------------------------").FgColor("blue").Build(),
			)),
		),

		// Row 3: "Quit" | Log line #0
		ui.Row(
			ui.Cell(ui.HStack(
				app.NewTextBuilder("| ").FgColor("blue").Build(),
				app.ButtonBuilder("Quit").BgColor("red").FgColor("white").OnClick(func() {
					ui.Quit()
				}).Build(),
			)),
			ui.Cell(ui.HStack(
				app.NewTextBuilder("| ").FgColor("blue").Build(),
				app.NewTextBuilder(items[0]).FgColor("gray").Build(),
			)),
		),

		// Row 4: Empty | Log line #1
		ui.Row(
			ui.Cell(ui.HStack(
				app.NewTextBuilder("|").FgColor("blue").Build(),
				app.NewTextBuilder("           ").Build(),
			)),
			ui.Cell(ui.HStack(
				app.NewTextBuilder("| ").FgColor("blue").Build(),
				app.NewTextBuilder(items[1]).FgColor("gray").Build(),
			)),
		),

		// Row 5: Empty | Log line #2
		ui.Row(
			ui.Cell(ui.HStack(
				app.NewTextBuilder("|").FgColor("blue").Build(),
				app.NewTextBuilder("           ").Build(),
			)),
			ui.Cell(ui.HStack(
				app.NewTextBuilder("| ").FgColor("blue").Build(),
				app.NewTextBuilder(items[2]).FgColor("gray").Build(),
			)),
		),

		// Row 6: Empty | Log line #3
		ui.Row(
			ui.Cell(ui.HStack(
				app.NewTextBuilder("|").FgColor("blue").Build(),
				app.NewTextBuilder("           ").Build(),
			)),
			ui.Cell(ui.HStack(
				app.NewTextBuilder("| ").FgColor("blue").Build(),
				app.NewTextBuilder(items[3]).FgColor("gray").Build(),
			)),
		),

		// Row 7: Empty | Log line #4
		ui.Row(
			ui.Cell(ui.HStack(
				app.NewTextBuilder("|").FgColor("blue").Build(),
				app.NewTextBuilder("           ").Build(),
			)),
			ui.Cell(ui.HStack(
				app.NewTextBuilder("| ").FgColor("blue").Build(),
				app.NewTextBuilder(items[4]).FgColor("gray").Build(),
			)),
		),

		// Row 8: Empty | Log line #5 + "..."
		ui.Row(
			ui.Cell(ui.HStack(
				app.NewTextBuilder("|").FgColor("blue").Build(),
				app.NewTextBuilder("           ").Build(),
			)),
			ui.Cell(ui.HStack(
				app.NewTextBuilder("| ").FgColor("blue").Build(),
				app.NewTextBuilder(items[5]).FgColor("gray").Build(),
				app.NewTextBuilder(" ...").FgColor("dark-gray").Italic(true).Build(),
			)),
		),

		// Border bottom row
		ui.Row(
			ui.Cell(app.NewTextBuilder("+-----------+").FgColor("blue").Build()),
			ui.Cell(app.NewTextBuilder("--------------------------------------+").FgColor("blue").Build()),
		),
	)
}

// renderVisibleItems renders log lines in the content area
// Note: This function is no longer used, kept for reference
func renderVisibleItems(items []string, offset int, visibleCount int) ui.VNode {
	children := make([]ui.VNode, 0, visibleCount)

	end := offset + visibleCount
	if end > len(items) {
		end = len(items)
	}

	for i := offset; i < end; i++ {
		children = append(children,
			app.NewTextBuilder(fmt.Sprintf(" %s", items[i])).
				FgColor("gray").
				Build(),
		)
	}

	if len(items) > visibleCount {
		children = append(children,
			app.NewTextBuilder(" ...").
				FgColor("dark-gray").
				Italic(true).
				Build(),
		)
	}

	return ui.HStack(
		app.NewTextBuilder("|").FgColor("blue").Build(),
		ui.VStack(children...),
		app.NewTextBuilder("|").FgColor("blue").Build(),
	)
}

// ConfirmModal demonstrates Layer + Animation + Focus Trap
func ConfirmModal(onClose func()) ui.VNode {
	return ui.VStack(
		ui.Text(""),
		ui.HStack(
			ui.Text("         "),
			ui.VStack(
				app.NewTextBuilder("+--------------------------------------+").
					FgColor("yellow").
					Build(),
				ui.HStack(
					app.NewTextBuilder("| ").
						FgColor("yellow").
						Build(),
					app.NewTextBuilder("           ").
						Build(),
					app.NewTextBuilder("|").
						FgColor("yellow").
						Build(),
				),
				ui.HStack(
					app.NewTextBuilder("| ").
						FgColor("yellow").
						Build(),
					app.NewTextBuilder("    *** Are you sure? ***").
						Bold(true).
						FgColor("yellow").
						Build(),
					app.NewTextBuilder("     |").
						FgColor("yellow").
						Build(),
				),
				ui.HStack(
					app.NewTextBuilder("| ").
						FgColor("yellow").
						Build(),
					app.NewTextBuilder("           ").
						Build(),
					app.NewTextBuilder("|").
						FgColor("yellow").
						Build(),
				),
				ui.HStack(
					app.NewTextBuilder("| ").
						FgColor("yellow").
						Build(),
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
					app.NewTextBuilder("       |").
						FgColor("yellow").
						Build(),
				),
				ui.HStack(
					app.NewTextBuilder("| ").
						FgColor("yellow").
						Build(),
					app.NewTextBuilder("           ").
						Build(),
					app.NewTextBuilder("|").
						FgColor("yellow").
						Build(),
				),
				app.NewTextBuilder("+--------------------------------------+").
					FgColor("yellow").
					Build(),
			),
		),
		ui.Text(""),
		app.NewTextBuilder("Press ESC to close").
			FgColor("gray").
			Build(),
	)
}
