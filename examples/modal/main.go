package main

import (
	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/ui"
)

func main() {
	ui.Run(func() ui.VNode {
		// Use int state: 0 = closed, 1 = open
		_, setState, getState := ui.UseStateInt(0)

		state := getState()

		// If modal is open, show modal content
		if state == 1 {
			return app.VStack(
				app.NewTextBuilder("┌───────────────────────────────────────┐").FgColor("cyan").Build(),
				app.NewTextBuilder("│           MODAL IS OPEN               │").FgColor("cyan").Build(),
				app.NewTextBuilder("│                                       │").FgColor("cyan").Build(),
				app.NewTextBuilder("│  Do you want to proceed?              │").FgColor("white").Build(),
				app.NewTextBuilder("│                                       │").FgColor("cyan").Build(),
				app.HStack(
					app.NewTextBuilder("│  ").FgColor("cyan").Build(),
					app.ButtonBuilder(" Yes ").
						OnClick(func() {
							setState(0)
						}).
						Build(),
					app.NewTextBuilder("  ").FgColor("cyan").Build(),
					app.ButtonBuilder(" No ").
						OnClick(func() {
							setState(0)
						}).
						Build(),
					app.NewTextBuilder("               │").FgColor("cyan").Build(),
				),
				app.NewTextBuilder("│                                       │").FgColor("cyan").Build(),
				app.NewTextBuilder("└───────────────────────────────────────┘").FgColor("cyan").Build(),
				app.Text(""),
				app.NewTextBuilder("Press Tab to focus, Enter to close").FgColor("gray").Build(),
			)
		}

		// Modal is closed - show main content
		return app.VStack(
			app.NewTextBuilder("Modal Demo").Bold(true).FgColor("cyan").Build(),
			app.Text(""),
			app.NewTextBuilder("Click the button below to open a modal dialog").FgColor("gray").Build(),
			app.Text(""),
			app.ButtonBuilder("  Show Modal  ").
				OnClick(func() {
					setState(1)
				}).
				Build(),
			app.Text(""),
			app.NewTextBuilder("Tab/Arrows: focus | Enter/Space: click").FgColor("gray").Build(),
		)
	},
		ui.WithWidth(50),
		ui.WithHeight(20),
		ui.WithTitle("Modal Demo"),
	)
}
