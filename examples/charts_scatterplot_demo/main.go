package main

import (
	"github.com/wwsheng009/mint/ui"
	scatterplotcomp "github.com/wwsheng009/mint/ui/components/charts/scatterplot"
)

func main() {
	err := ui.Run(ScatterPlotDemo,
		ui.WithWidth(72),
		ui.WithHeight(20),
		ui.WithTitle("ScatterPlot Demo"),
	)
	if err != nil {
		panic(err)
	}
}

func ScatterPlotDemo() ui.VNode {
	return ui.NewVStack().
		SetGap(0).
		SetChildrenList([]ui.VNode{
			ui.NewTextBuilder("Charts ScatterPlot Demo").Build(),
			scatterplotcomp.NewBuilder(nil).
				SetID("scatterplot-demo-correlation").
				Title("Service Correlation").
				Series(
					scatterplotcomp.Series{
						Name: "API",
						Points: []scatterplotcomp.Point{
							{X: 2, Y: 3},
							{X: 4, Y: 6},
							{X: 8, Y: 9},
						},
					},
					scatterplotcomp.Series{
						Name:  "Worker",
						Glyph: '◆',
						Points: []scatterplotcomp.Point{
							{X: 3, Y: 4},
							{X: 6, Y: 5},
							{X: 7, Y: 8},
						},
					},
				).
				Domain(scatterplotcomp.NewDomain(0, 10, 0, 12)).
				XReferenceLineLabeled(5, "Target").
				YReferenceBandLabeled(6, 9, "Risk").
				Width(13).
				Height(6).
				ShowLegend(true).
				ShowGrid(true).
				ShowAxis(true).
				Build(),
		})
}
