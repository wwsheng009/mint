package main

import (
	"github.com/wwsheng009/mint/ui"
	bulletchartcomp "github.com/wwsheng009/mint/ui/components/charts/bulletchart"
)

func main() {
	err := ui.Run(BulletChartDemo,
		ui.WithWidth(72),
		ui.WithHeight(20),
		ui.WithTitle("Charts BulletChart Demo"),
	)
	if err != nil {
		panic(err)
	}
}

func BulletChartDemo() ui.VNode {
	return ui.NewVStack().
		SetGap(0).
		SetChildrenList([]ui.VNode{
			ui.NewTextBuilder("Charts BulletChart Demo").Build(),
			bulletchartcomp.NewBuilder().
				SetID("bulletchart-demo-throughput").
				Label("Throughput").
				Value(82).
				Target(75).
				Max(100).
				Width(22).
				HigherIsBetter().
				BelowValueLabel().
				Build(),
			bulletchartcomp.NewBuilder().
				SetID("bulletchart-demo-latency").
				Label("Latency Ceiling").
				Value(173).
				Target(200).
				Max(250).
				Width(22).
				LowerIsBetter().
				BelowValueLabel().
				Build(),
			bulletchartcomp.NewBuilder().
				SetID("bulletchart-demo-error-rate").
				Label("Error Rate").
				Value(0).
				Target(5).
				Max(100).
				Width(22).
				NeutralDirection().
				BelowValueLabel().
				Build(),
		})
}
