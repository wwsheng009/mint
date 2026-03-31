package main

import (
	"github.com/wwsheng009/mint/ui"
	heatmapcomp "github.com/wwsheng009/mint/ui/components/charts/heatmap"
)

func main() {
	err := ui.Run(HeatmapDemo,
		ui.WithWidth(72),
		ui.WithHeight(20),
		ui.WithTitle("Heatmap Demo"),
	)
	if err != nil {
		panic(err)
	}
}

func HeatmapDemo() ui.VNode {
	return ui.NewVStack().
		SetGap(0).
		SetChildrenList([]ui.VNode{
			ui.NewTextBuilder("Charts Heatmap Demo").Build(),
			heatmapcomp.NewBuilder([][]float64{
				{1, 2, 3, 4, 5},
				{2, 4, 6, 8, 9},
				{1, 3, 5, 7, 8},
				{2, 2, 4, 6, 7},
			}).
				SetID("heatmap-demo-windowed").
				Title("Regional Hotspots").
				RowLabels([]string{
					"North America",
					"South America",
					"Europe",
					"Asia Pacific",
				}).
				ColLabels([]string{"Mon", "Tue", "Wed", "Thu", "Fri"}).
				RowWindow(1, 3).
				ColWindow(1, 3).
				MaxRowLabelWidth(4).
				CompactLegend().
				ColorMode(ui.HeatmapColorMode16).
				ShowAxis(true).
				ShowLegend(true).
				Build(),
		})
}
