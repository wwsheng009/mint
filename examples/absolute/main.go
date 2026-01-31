package main

import (
	"fmt"

	"github.com/wwsheng009/mint/ui"
)

func main() {
	ui.Run(func() ui.VNode {
		// Simple badge example
		count, setCount, _ := ui.UseStateInt(0)

		return ui.VStack(
			ui.NewTextBuilder("Absolute Positioning Demo").Bold(true).FgColor("cyan").Build(),
			ui.Text(""),
			ui.Text("Button with notification badge:"),
			ui.Text(""),
			ui.HStack(
				ui.ButtonBuilder("  Messages  ").
					OnClick(func() {
						setCount(count + 1)
					}).
					Build(),
				// Badge positioned absolutely relative to parent
				ui.AbsoluteBuilder(
					ui.NewTextBuilder("New!").
						FgColor("red").
						Bold(true).
						Build(),
				).
					Left(ui.AbsolutePosition(15)).
					Top(ui.AbsolutePosition(0)).
					Build(),
			),
			ui.Text(""),
			ui.NewTextBuilder("Stacked Elements").FgColor("yellow").Build(),
			ui.Text(""),
			ui.VStack(
				ui.Text("Background layer"),
				ui.HStack(
					ui.Text("Middle layer"),
					ui.AbsoluteBuilder(
						ui.NewTextBuilder("OVERLAY").FgColor("white").BgColor("red").Build(),
					).
						Left(ui.AbsolutePosition(20)).
						Top(ui.AbsolutePosition(0)).
						ZIndex(10).
						Build(),
				),
			),
			ui.Text(""),
			ui.NewTextBuilder(fmt.Sprintf("Click count: %d", count)).Build(),
		)
	},
		ui.WithWidth(50),
		ui.WithHeight(20),
		ui.WithTitle("Absolute Demo"),
	)
}
