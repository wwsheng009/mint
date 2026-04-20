package main

import (
	linechartprototype "github.com/wwsheng009/mint/examples/charts_linechart_image_prototype/demo"
	"github.com/wwsheng009/mint/ui"
)

func main() {
	err := ui.Run(LineChartImagePrototype,
		ui.WithWidth(104),
		ui.WithHeight(24),
		ui.WithTitle("LineChart Image Prototype (Chart Pixel Backend Paused)"),
	)
	if err != nil {
		panic(err)
	}
}

func LineChartImagePrototype() ui.VNode {
	return linechartprototype.Build()
}
