package main

import (
	linechartdemo "github.com/wwsheng009/mint/examples/charts_linechart_demo/demo"
	"github.com/wwsheng009/mint/ui"
)

func main() {
	err := ui.Run(LineChartDemo,
		ui.WithWidth(72),
		ui.WithHeight(22),
		ui.WithTitle("LineChart Demo"),
	)
	if err != nil {
		panic(err)
	}
}

func LineChartDemo() ui.VNode {
	return linechartdemo.Build()
}
