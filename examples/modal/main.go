package main

import (
	"github.com/wwsheng009/mint/ui"
)

func main() {
	ui.Run(func() ui.VNode {
		// Use int state: 0 = closed, 1 = open
		_, setState, getState := ui.UseStateInt(0)

		state := getState()

		// If modal is open, show modal content
		if state == 1 {
			return ui.VStack(
				ui.NewTextBuilder("┌───────────────────────────────────────┐").FgColor("cyan").Build(),
				ui.NewTextBuilder("│           MODAL IS OPEN               │").FgColor("cyan").Build(),
				ui.NewTextBuilder("│                                       │").FgColor("cyan").Build(),
				ui.NewTextBuilder("│  Do you want to proceed?              │").FgColor("white").Build(),
				ui.NewTextBuilder("│                                       │").FgColor("cyan").Build(),
				ui.HStack(
					ui.NewTextBuilder("│  ").FgColor("cyan").Build(),
					ui.ButtonBuilder(" Yes ").
						OnClick(func() {
							setState(0)
						}).
						Build(),
					ui.NewTextBuilder("  ").FgColor("cyan").Build(),
					ui.ButtonBuilder(" No ").
						OnClick(func() {
							setState(0)
						}).
						Build(),
					ui.NewTextBuilder("               │").FgColor("cyan").Build(),
				),
				ui.NewTextBuilder("│                                       │").FgColor("cyan").Build(),
				ui.NewTextBuilder("└───────────────────────────────────────┘").FgColor("cyan").Build(),
				ui.Text(""),
				ui.NewTextBuilder("Press Tab to focus, Enter to close").FgColor("gray").Build(),
			)
		}

		// Modal is closed - show main content
		return ui.VStack(
			ui.NewTextBuilder("Modal Demo").Bold(true).FgColor("cyan").Build(),
			ui.Text(""),
			ui.NewTextBuilder("Click the button below to open a modal dialog").FgColor("gray").Build(),
			ui.Text(""),
			ui.ButtonBuilder("  Show Modal  ").
				OnClick(func() {
					setState(1)
				}).
				Build(),
			ui.Text(""),
			ui.NewTextBuilder("Tab/Arrows: focus | Enter/Space: click").FgColor("gray").Build(),
		)
	},
		ui.WithWidth(50),
		ui.WithHeight(20),
		ui.WithTitle("Modal Demo"),
	)
}
