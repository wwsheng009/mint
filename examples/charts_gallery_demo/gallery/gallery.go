package gallery

import (
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
	barchartcomp "github.com/wwsheng009/mint/ui/components/charts/barchart"
	bulletchartcomp "github.com/wwsheng009/mint/ui/components/charts/bulletchart"
	candlestickcomp "github.com/wwsheng009/mint/ui/components/charts/candlestick"
	heatmapcomp "github.com/wwsheng009/mint/ui/components/charts/heatmap"
	linechartcomp "github.com/wwsheng009/mint/ui/components/charts/linechart"
	scatterplotcomp "github.com/wwsheng009/mint/ui/components/charts/scatterplot"
	sparklinecomp "github.com/wwsheng009/mint/ui/components/charts/sparkline"
)

// Build returns the compact charts gallery view used by the example and e2e tests.
func Build() ui.VNode {
	return ui.NewVStack().
		SetGap(0).
		SetPadding(0, 1, 0, 1).
		SetChildrenList([]ui.VNode{
			headerPanel(),
			ui.HStackBuilder(
				ui.Flex(leftColumn(), 1),
				ui.Flex(rightColumn(), 1),
			).Gap(2).Stretch().Build(),
		})
}

func headerPanel() ui.VNode {
	content := ui.NewVStack().
		SetGap(0).
		SetChildrenList([]ui.VNode{
			ui.NewTextBuilder("Mint Charts Gallery").
				Bold(true).
				FgColor("cyan").
				Build(),
			ui.NewTextBuilder("Sparkline, bullet, line, bar, heatmap, tape, scatter.").
				FgColor("bright-black").
				Build(),
		})
	return panelBox("Overview", style.Cyan, 76, content)
}

func leftColumn() ui.VNode {
	return ui.NewVStack().
		SetGap(1).
		SetChildrenList([]ui.VNode{
			kpiPulsePanel(),
			bulletTargetsPanel(),
		})
}

func rightColumn() ui.VNode {
	return ui.NewVStack().
		SetGap(1).
		SetChildrenList([]ui.VNode{
			trafficTrendPanel(),
			throughputHotspotsPanel(),
		})
}

func kpiPulsePanel() ui.VNode {
	content := ui.NewVStack().
		SetGap(0).
		SetChildrenList([]ui.VNode{
			ui.NewHStack().
				SetGap(1).
				SetChildrenList([]ui.VNode{
					ui.Flex(metricCard("Requests", "18.4k", "7d +12.4%", []float64{7, 8, 8, 9, 11, 12, 13, 12, 14, 16}, style.Cyan), 1),
					ui.Flex(metricCard("Error Budget", "99.94%", "burn steady", []float64{10, 10, 9, 9, 9, 8, 8, 8, 7, 7}, style.Green), 1),
				}),
		})
	return panelBox("KPI Pulse", style.Cyan, 36, content)
}

func metricCard(title, value, note string, trend []float64, accent style.Color) ui.VNode {
	return ui.NewVStack().
		SetGap(0).
		SetBorderColor(accent).
		SingleBorder(title).
		SetChildrenList([]ui.VNode{
			ui.NewTextBuilder(value).
				Bold(true).
				FgColor("bright-white").
				Build(),
			sparklinecomp.NewBuilder(trend).
				Width(12).
				Braille().
				Style(style.NewStyle().Foreground(accent)).
				Build(),
			ui.NewTextBuilder(note).
				FgColor("bright-black").
				Build(),
		})
}

func bulletTargetsPanel() ui.VNode {
	content := ui.NewVStack().
		SetGap(0).
		SetChildrenList([]ui.VNode{
			bulletchartcomp.NewBuilder().
				Label("Latency").
				Value(173).
				Target(200).
				Max(250).
				Width(22).
				QualitativeRanges(
					bulletchartcomp.QualitativeRange{Limit: 100, Glyph: '░'},
					bulletchartcomp.QualitativeRange{Limit: 180, Glyph: '▒'},
					bulletchartcomp.QualitativeRange{Limit: 250, Glyph: '▓'},
				).
				BelowValueLabel().
				Build(),
			bulletchartcomp.NewBuilder().
				Label("Availability").
				Value(996).
				Target(999).
				Max(1000).
				Width(22).
				QualitativeRanges(
					bulletchartcomp.QualitativeRange{Limit: 970, Glyph: '░'},
					bulletchartcomp.QualitativeRange{Limit: 990, Glyph: '▒'},
					bulletchartcomp.QualitativeRange{Limit: 1000, Glyph: '▓'},
				).
				BelowValueLabel().
				Build(),
		})
	return panelBox("SLO Bullet Charts", style.Yellow, 36, content)
}

func trafficTrendPanel() ui.VNode {
	content := ui.NewVStack().
		SetGap(0).
		SetChildrenList([]ui.VNode{
			linechartcomp.NewBuilder(nil).
				Series(
					linechartcomp.Series{Name: "API", Data: []float64{32, 40, 38, 48, 45, 52, 50, 58}},
					linechartcomp.Series{Name: "Worker", Data: []float64{28, 29, 35, 34, 39, 41, 44, 46}},
				).
				Width(30).
				Height(2).
				ShowLegend(true).
				ShowGrid(true).
				ShowAxis(false).
				ShowPoints(true).
				Build(),
			ui.HStackBuilder(
				ui.Flex(tapeMiniPanel(), 1),
				ui.Flex(scatterMiniPanel(), 1),
			).Gap(1).Stretch().Build(),
		})
	return panelBox("Traffic + Tape", style.Blue, 36, content)
}

func tapeMiniPanel() ui.VNode {
	return ui.NewVStack().
		SetGap(0).
		SetChildrenList([]ui.VNode{
			ui.NewTextBuilder("Tape").
				FgColor("bright-black").
				Build(),
			candlestickcomp.NewBuilder([]candlestickcomp.Candle{
				{Label: "M", Open: 100, High: 110, Low: 95, Close: 107},
				{Label: "T", Open: 107, High: 112, Low: 101, Close: 103},
				{Label: "W", Open: 103, High: 116, Low: 100, Close: 103},
				{Label: "T", Open: 103, High: 109, Low: 98, Close: 108},
			}).
				Width(7).
				Height(3).
				ShowLegend(false).
				ShowGrid(false).
				ShowAxis(false).
				Build(),
		})
}

func scatterMiniPanel() ui.VNode {
	return ui.NewVStack().
		SetGap(0).
		SetChildrenList([]ui.VNode{
			ui.NewTextBuilder("Scatter").
				FgColor("bright-black").
				Build(),
			scatterplotcomp.NewBuilder(nil).
				Series(
					scatterplotcomp.Series{
						Name:   "API",
						Points: []scatterplotcomp.Point{{X: 1, Y: 2}, {X: 2, Y: 4}, {X: 4, Y: 5}},
					},
					scatterplotcomp.Series{
						Name:   "Worker",
						Points: []scatterplotcomp.Point{{X: 2, Y: 3}, {X: 3, Y: 2}, {X: 5, Y: 4}},
						Glyph:  '◆',
					},
				).
				Width(11).
				Height(3).
				ShowLegend(false).
				ShowGrid(false).
				ShowAxis(false).
				Build(),
		})
}

func throughputHotspotsPanel() ui.VNode {
	content := ui.HStackBuilder(
		ui.Flex(hotspotMiniPanel(), 1),
		ui.Flex(regionalThroughputChart(), 1),
	).Gap(1).Stretch().Build()
	return panelBox("Throughput + Hotspots", style.BrightBlue, 36, content)
}

func hotspotMiniPanel() ui.VNode {
	return ui.NewVStack().
		SetGap(0).
		SetChildrenList([]ui.VNode{
			ui.NewTextBuilder("Hotspots").
				Bold(true).
				Build(),
			heatmapcomp.NewBuilder([][]float64{
				{1, 3, 5, 7},
				{2, 4, 6, 8},
				{3, 5, 7, 9},
			}).
				RowLabels([]string{"API", "DB", "Cache"}).
				ColLabels([]string{"M", "T", "W", "T"}).
				RowWindow(0, 2).
				ColWindow(0, 3).
				ViewportScale().
				ShowSummary(true).
				ShowLegend(false).
				ShowAxis(true).
				MaxRowLabelWidth(4).
				Build(),
		})
}

func regionalThroughputChart() ui.VNode {
	return barchartcomp.NewBuilder(nil).
		Labels([]string{"NA", "EU"}).
		Series(
			barchartcomp.Series{Name: "Ingress", Values: []float64{12, 9}},
			barchartcomp.Series{Name: "Egress", Values: []float64{10, 11}},
		).
		Stacked().
		Horizontal().
		Width(14).
		Height(5).
		ShowLegend(true).
		ShowAxis(false).
		ShowValue(true).
		Build()
}

func panelBox(title string, borderColor style.Color, contentWidth int, content ui.VNode) ui.VNode {
	return ui.NewPanelBuilder().
		Title(title).
		ContentWidth(contentWidth).
		AutoHeight().
		Content(content).
		Rounded().
		BorderColor(borderColor).
		Build()
}
