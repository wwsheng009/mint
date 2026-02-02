package main

import (
	"fmt"

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/ui"
)

func main() {
	ui.Run(func() ui.VNode {
		// Simple badge example
		count, setCount, _ := ui.UseStateInt(0)

		return app.VStack(
			app.NewTextBuilder("Absolute Positioning Demo").Bold(true).FgColor("cyan").Build(),
			app.Text(""),
			app.Text("Button with notification badge:"),
			app.Text(""),
			app.HStack(
				app.ButtonBuilder("  Messages  ").
					OnClick(func() {
						setCount(count + 1)
					}).
					Build(),
				// Badge positioned absolutely relative to parent
				app.AbsoluteBuilder(
					app.NewTextBuilder("New!").
						FgColor("red").
						Bold(true).
						Build(),
				).
					Left(app.AbsolutePosition(15)).
					Top(app.AbsolutePosition(0)).
					Build(),
			),
			app.Text(""),
			app.NewTextBuilder("Stacked Elements").FgColor("yellow").Build(),
			app.Text(""),
			app.VStack(
				app.Text("Background layer"),
				app.HStack(
					app.Text("Middle layer"),
					app.AbsoluteBuilder(
						app.NewTextBuilder("OVERLAY").FgColor("white").BgColor("red").Build(),
					).
						Left(app.AbsolutePosition(20)).
						Top(app.AbsolutePosition(0)).
						ZIndex(10).
						Build(),
				),
			),
			app.Text(""),
			app.NewTextBuilder(fmt.Sprintf("Click count: %d", count)).Build(),
		)
	},
		ui.WithWidth(50),
		ui.WithHeight(20),
		ui.WithTitle("Absolute Demo"),
	)
}
