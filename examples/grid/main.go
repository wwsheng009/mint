package main

import (
	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/ui"
)

func main() {
	ui.Run(func() ui.VNode {
		// Simple 2x2 grid with equal columns
		return app.VStack(
			app.NewTextBuilder("Grid Layout Demo").Bold(true).FgColor("cyan").Build(),
			app.Text(""),
			app.GridBuilder().
				Columns(app.Fixed(20), app.Fixed(20), app.Flex{Factor: 1}).
				Rows(app.Auto{}, app.Auto{}, app.Auto{}).
				Cell(0, 0,
					app.NewTextBuilder("Name:").FgColor("gray").Build(),
				).
				Cell(0, 1,
					app.NewTextBuilder("Age:").FgColor("gray").Build(),
				).
				Cell(0, 2,
					app.NewTextBuilder("Status:").FgColor("gray").Build(),
				).
				Cell(1, 0,
					app.Text("John Doe"),
				).
				Cell(1, 1,
					app.Text("30"),
				).
				Cell(1, 2,
					app.NewTextBuilder("Active").FgColor("green").Build(),
				).
				Gap(1, 0).
				Padding(1, 1, 1, 1).
				Build(),
			app.Text(""),
			app.NewTextBuilder("Dashboard Layout Example").FgColor("yellow").Build(),
			app.Text(""),
			app.GridBuilder().
				Columns(app.Fixed(15), app.Flex{Factor: 2}, app.Fixed(15)).
				Rows(app.Auto{}, app.Auto{}).
				CellSpan(0, 0, 2, 1, // CPU - spans 2 rows, 1 col
					app.VStack(
						app.NewTextBuilder("CPU").FgColor("cyan").Build(),
						app.NewTextBuilder("45%").FgColor("green").Build(),
					),
				).
				Cell(0, 1, // Memory - top right
					app.VStack(
						app.NewTextBuilder("Memory").FgColor("cyan").Build(),
						app.NewTextBuilder("2.1GB / 8GB").FgColor("yellow").Build(),
					),
				).
				Cell(1, 1, // Disk - bottom right
					app.VStack(
						app.NewTextBuilder("Disk").FgColor("cyan").Build(),
						app.NewTextBuilder("120GB / 500GB").FgColor("blue").Build(),
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
