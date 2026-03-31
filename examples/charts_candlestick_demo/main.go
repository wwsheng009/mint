package main

import (
	"github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/ui/components/charts/candlestick"
)

func main() {
	err := ui.Run(CandlestickDemo,
		ui.WithWidth(72),
		ui.WithHeight(20),
		ui.WithTitle("Charts Candlestick Demo"),
	)
	if err != nil {
		panic(err)
	}
}

func CandlestickDemo() ui.VNode {
	return ui.NewVStack().
		SetGap(0).
		SetChildrenList([]ui.VNode{
			ui.NewTextBuilder("Charts Candlestick Demo").Build(),
			candlestick.NewBuilder([]candlestick.Candle{
				{Label: "M", Open: 100, High: 110, Low: 96, Close: 107, Volume: 1800},
				{Label: "T", Open: 107, High: 112, Low: 101, Close: 103, Volume: 1200},
				{Label: "W", Open: 103, High: 116, Low: 99, Close: 111, Volume: 1500},
				{Label: "T", Open: 111, High: 118, Low: 108, Close: 109, Volume: 900},
			}).
				Title("Daily Tape").
				Width(9).
				Height(6).
				ShowLegend(true).
				ShowGrid(true).
				ShowAxis(true).
				ShowVolume(true).
				VolumeHeight(3).
				Build(),
		})
}
