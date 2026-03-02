package main

import (
	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/ui"
)

func main() {
	ui.Run(func() ui.VNode {
		// Simple 2x2 grid with equal columns
		return app.VStack(
			ui.NewTextBuilder("Grid Layout Demo").Bold(true).FgColor("cyan").Build(),
			ui.Text(""),
			app.GridBuilder().
				Columns(app.Fixed(20), app.Fixed(20), app.Flex{Factor: 1}).
				Rows(app.Auto{}, app.Auto{}, app.Auto{}).
				Cell(0, 0,
					ui.NewTextBuilder("Name:").FgColor("gray").Build(),
				).
				Cell(0, 1,
					ui.NewTextBuilder("Age:").FgColor("gray").Build(),
				).
				Cell(0, 2,
					ui.NewTextBuilder("Status:").FgColor("gray").Build(),
				).
				Cell(1, 0,
					ui.Text("John Doe"),
				).
				Cell(1, 1,
					ui.Text("30"),
				).
				Cell(1, 2,
					ui.NewTextBuilder("Active").FgColor("green").Build(),
				).
				Gap(1, 0).
				Padding(1, 1, 1, 1).
				Build(),
			ui.Text(""),
			ui.NewTextBuilder("Dashboard Layout Example").FgColor("yellow").Build(),
			ui.Text(""),
			app.GridBuilder().
				Columns(app.Fixed(15), app.Flex{Factor: 2}, app.Fixed(15)).
				Rows(app.Auto{}, app.Auto{}).
				CellSpan(0, 0, 2, 1, // CPU - spans 2 rows, 1 col
					app.VStack(
						ui.NewTextBuilder("CPU").FgColor("cyan").Build(),
						ui.NewTextBuilder("45%").FgColor("green").Build(),
					),
				).
				Cell(0, 1, // Memory - top right
					app.VStack(
						ui.NewTextBuilder("Memory").FgColor("cyan").Build(),
						ui.NewTextBuilder("2.1GB / 8GB").FgColor("yellow").Build(),
					),
				).
				Cell(1, 1, // Disk - bottom right
					app.VStack(
						ui.NewTextBuilder("Disk").FgColor("cyan").Build(),
						ui.NewTextBuilder("120GB / 500GB").FgColor("blue").Build(),
					),
				).
				Gap(1, 0).
				Padding(1, 1, 1, 1).
				Build(),
		)
	},
		ui.WithWidth(60),
		ui.WithHeight(20),
		ui.WithTitle("Grid Demo"),
	)
}
