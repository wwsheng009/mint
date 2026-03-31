package e2e

import (
	"testing"

	fwtheme "github.com/wwsheng009/mint/framework/theme"
	rtstyle "github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
	barchartcomp "github.com/wwsheng009/mint/ui/components/charts/barchart"
	bulletchartcomp "github.com/wwsheng009/mint/ui/components/charts/bulletchart"
	candlestickcomp "github.com/wwsheng009/mint/ui/components/charts/candlestick"
	heatmapcomp "github.com/wwsheng009/mint/ui/components/charts/heatmap"
	linechartcomp "github.com/wwsheng009/mint/ui/components/charts/linechart"
	scatterplotcomp "github.com/wwsheng009/mint/ui/components/charts/scatterplot"
	sparklinecomp "github.com/wwsheng009/mint/ui/components/charts/sparkline"
)

func newChartsStaticApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Charts E2E Fixture").Build(),
				linechartcomp.NewBuilder(nil).
					SetID("linechart-multi").
					Title("Latency Trend").
					Series(
						linechartcomp.Series{Name: "API", Data: []float64{1, 3, 2, 5, 4}},
						linechartcomp.Series{Name: "Worker", Data: []float64{2, 2, 4, 3, 5}},
					).
					Width(9).
					Height(3).
					ShowLegend(true).
					ShowGrid(true).
					ShowAxis(true).
					ShowPoints(true).
					Build(),
				barchartcomp.NewBuilder(nil).
					SetID("barchart-multi").
					Title("Throughput Bars").
					Labels([]string{"A", "B", "C"}).
					Series(
						barchartcomp.Series{Name: "Revenue", Values: []float64{3, 5, 2}},
						barchartcomp.Series{Name: "Cost", Values: []float64{2, 4, 1}},
					).
					Width(8).
					Height(3).
					ShowLegend(true).
					ShowAxis(true).
					Build(),
				barchartcomp.NewBuilder(nil).
					SetID("barchart-stacked").
					Title("Throughput Stacked").
					Labels([]string{"X", "Y", "Z"}).
					Series(
						barchartcomp.Series{Name: "North", Values: []float64{3, 2, 1}},
						barchartcomp.Series{Name: "South", Values: []float64{1, 3, 2}},
					).
					Stacked().
					Width(5).
					Height(3).
					ShowLegend(true).
					ShowAxis(true).
					Build(),
			})
	}
}

func newChartsHorizontalApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Charts Horizontal Fixture").Build(),
				barchartcomp.NewBuilder(nil).
					SetID("barchart-horizontal").
					Title("Throughput Horizontal").
					Labels([]string{"East", "West"}).
					Series(
						barchartcomp.Series{Name: "Online", Values: []float64{4, 2}},
						barchartcomp.Series{Name: "Retail", Values: []float64{3, 5}},
					).
					Horizontal().
					Width(18).
					Height(5).
					ShowLegend(true).
					ShowAxis(true).
					Build(),
			})
	}
}

func newChartsBarLabelFoldApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Charts Bar Label Fold Fixture").Build(),
				barchartcomp.NewBuilder([]float64{4, 3, 5}).
					SetID("barchart-fold-vertical").
					Title("Dense Vertical Labels").
					Labels([]string{"North America", "Latin America", "Asia Pacific"}).
					Width(5).
					Height(4).
					ShowAxis(true).
					Build(),
				barchartcomp.NewBuilder([]float64{10, 8}).
					SetID("barchart-fold-horizontal").
					Title("Dense Horizontal Labels").
					Labels([]string{"North America", "Latin America"}).
					Horizontal().
					Width(10).
					Height(3).
					ShowAxis(true).
					Build(),
			})
	}
}

func newChartsValueLabelApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Charts Value Label Fixture").Build(),
				barchartcomp.NewBuilder([]float64{12, 7, 15}).
					SetID("barchart-values-vertical").
					Title("Value Labels Vertical").
					Labels([]string{"A", "B", "C"}).
					Width(5).
					Height(4).
					ShowAxis(true).
					ShowValue(true).
					Build(),
				barchartcomp.NewBuilder([]float64{12, 9}).
					SetID("barchart-values-horizontal").
					Title("Value Labels Horizontal").
					Labels([]string{"East", "West"}).
					Horizontal().
					Width(22).
					Height(3).
					ShowAxis(true).
					ShowValue(true).
					Build(),
			})
	}
}

func newChartsHeatmapApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Charts Heatmap Fixture").Build(),
				heatmapcomp.NewBuilder([][]float64{
					{1, 2, 3, 4},
					{2, 4, 6, 8},
					{1, 3, 5, 7},
				}).
					SetID("heatmap-matrix").
					Title("Matrix Heatmap").
					RowLabels([]string{"API", "DB", "Queue"}).
					ColLabels([]string{"M", "T", "W", "T"}).
					ShowAxis(true).
					ShowLegend(true).
					Build(),
			})
	}
}

func newChartsHeatmapCompactLegendApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Charts Heatmap Compact Fixture").Build(),
				heatmapcomp.NewBuilder([][]float64{
					{1, 3, 5},
					{2, 4, 6},
				}).
					SetID("heatmap-compact-legend").
					Title("Compact Legend Heatmap").
					RowLabels([]string{"API", "DB"}).
					ColLabels([]string{"M", "T", "W"}).
					ShowAxis(true).
					ShowLegend(true).
					CompactLegend().
					ColorMode(ui.HeatmapColorMode16).
					Build(),
			})
	}
}

func newChartsHeatmapViewportApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Charts Heatmap Viewport Fixture").Build(),
				heatmapcomp.NewBuilder([][]float64{
					{1, 2, 3, 4, 5},
					{2, 3, 4, 5, 6},
					{3, 4, 5, 6, 7},
					{4, 5, 6, 7, 8},
				}).
					SetID("heatmap-viewport").
					Title("Viewport Heatmap").
					RowLabels([]string{"North America", "South America", "Europe", "Asia Pacific"}).
					ColLabels([]string{"Mon", "Tue", "Wed", "Thu", "Fri"}).
					Viewport(heatmapcomp.NewViewport(1, 2, 1, 3)).
					MaxRowLabelWidth(4).
					ShowAxis(true).
					ShowLegend(true).
					Build(),
			})
	}
}

func newChartsHeatmapViewportScaleApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Charts Heatmap Contrast Fixture").Build(),
				heatmapcomp.NewBuilder([][]float64{
					{0, 1000, 1000, 1000},
					{40, 41, 42, 43},
					{41, 42, 43, 44},
					{100, 120, 140, 160},
				}).
					SetID("heatmap-viewport-scale").
					Title("Viewport Contrast Heatmap").
					RowLabels([]string{"North America", "South America", "Europe", "Asia Pacific"}).
					ColLabels([]string{"Mon", "Tue", "Wed", "Thu"}).
					Viewport(heatmapcomp.NewViewport(1, 2, 1, 3)).
					ViewportScale().
					MaxRowLabelWidth(4).
					ShowAxis(true).
					ShowLegend(true).
					Build(),
			})
	}
}

func newChartsHeatmapAutoScaleApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Charts Heatmap Auto Scale Fixture").Build(),
				heatmapcomp.NewBuilder([][]float64{
					{0, 1000, 1000, 1000},
					{40, 41, 42, 43},
					{41, 42, 43, 44},
					{100, 120, 140, 160},
				}).
					SetID("heatmap-auto-scale").
					Title("Auto Contrast Heatmap").
					RowLabels([]string{"North America", "South America", "Europe", "Asia Pacific"}).
					ColLabels([]string{"Mon", "Tue", "Wed", "Thu"}).
					Viewport(heatmapcomp.NewViewport(1, 2, 1, 3)).
					AutoScale().
					MaxRowLabelWidth(4).
					ShowAxis(true).
					ShowLegend(true).
					Build(),
			})
	}
}

func newChartsHeatmapAutoThresholdApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Charts Heatmap Auto Threshold Fixture").Build(),
				heatmapcomp.NewBuilder([][]float64{
					{40, 41, 42, 1000},
					{41, 42, 43, 1000},
					{42, 43, 44, 1000},
					{43, 44, 45, 1000},
				}).
					SetID("heatmap-auto-threshold").
					Title("Auto Threshold Heatmap").
					RowLabels([]string{"North", "South", "East", "West"}).
					ColLabels([]string{"Mon", "Tue", "Wed", "Thu"}).
					Viewport(heatmapcomp.NewViewport(0, 4, 0, 3)).
					AutoScale().
					ShowAxis(true).
					ShowLegend(false).
					Build(),
			})
	}
}

func newChartsHeatmapSummaryApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Charts Heatmap Summary Fixture").Build(),
				heatmapcomp.NewBuilder([][]float64{
					{40, 41, 42, 43},
					{41, 42, 43, 44},
				}).
					SetID("heatmap-summary").
					Title("Summary Heatmap").
					RowLabels([]string{"South America", "Europe"}).
					ColLabels([]string{"Tue", "Wed", "Thu", "Fri"}).
					Viewport(heatmapcomp.NewViewport(0, 2, 0, 3)).
					ShowSummary(true).
					MaxRowLabelWidth(4).
					ShowAxis(true).
					ShowLegend(false).
					Build(),
			})
	}
}

func newChartsHeatmapCompactSummaryApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Charts Heatmap Compact Summary Fixture").Build(),
				heatmapcomp.NewBuilder([][]float64{
					{40, 41, 42, 43},
					{41, 42, 43, 44},
				}).
					SetID("heatmap-compact-summary").
					Title("Compact Summary Heatmap").
					RowLabels([]string{"South America", "Europe"}).
					ColLabels([]string{"Tue", "Wed", "Thu", "Fri"}).
					Viewport(heatmapcomp.NewViewport(0, 2, 0, 3)).
					CompactSummary().
					MaxRowLabelWidth(4).
					ShowAxis(true).
					ShowLegend(false).
					Build(),
			})
	}
}

func newChartsSparklineApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Charts Sparkline Fixture").Build(),
				sparklinecomp.NewBuilder([]float64{7, 8, 8, 9, 11, 12, 13, 12, 14, 16}).
					SetID("sparkline-requests").
					Title("Requests Sparkline").
					Width(12).
					Braille().
					Build(),
				sparklinecomp.NewBuilder([]float64{10, 10, 9, 9, 9, 8, 8, 8, 7, 7}).
					SetID("sparkline-budget").
					Title("Budget Sparkline").
					Width(12).
					Block().
					Build(),
				sparklinecomp.NewBuilder([]float64{2, 3, 5, 4, 6, 8, 7, 9, 8, 10}).
					SetID("sparkline-errors").
					Title("Errors Sparkline").
					Width(12).
					ASCII().
					Build(),
				sparklinecomp.NewBuilder([]float64{7, 8, 9, 10, 8, 7, 6, 5}).
					SetID("sparkline-auto-compact").
					Title("Auto Compact Sparkline").
					Width(4).
					Auto().
					Build(),
			})
	}
}

func newChartsSparklineEnhancedApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Charts Sparkline Enhanced Fixture").Build(),
				sparklinecomp.NewBuilder([]float64{2, 5, 3, 8, 4, 7, 1, 9}).
					SetID("sparkline-enhanced-latency").
					Title("Latency Sparkline").
					Width(8).
					Height(3).
					HighlightMinMax(true).
					InlineLabel("live").
					Build(),
			})
	}
}

func newChartsBulletChartApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Charts Bullet Fixture").Build(),
				bulletchartcomp.NewBuilder().
					SetID("bulletchart-latency").
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
					TargetMarkerRune('╻').
					BelowValueLabel().
					Build(),
				bulletchartcomp.NewBuilder().
					SetID("bulletchart-availability").
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
					TargetMarkerRune('┆').
					BelowValueLabel().
					Build(),
			})
	}
}

func newChartsBulletChartDirectionApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Charts Bullet Direction Fixture").Build(),
				bulletchartcomp.NewBuilder().
					SetID("bulletchart-throughput").
					Label("Throughput").
					Value(82).
					Target(75).
					Max(100).
					Width(22).
					HigherIsBetter().
					TargetMarkerRune('╻').
					BelowValueLabel().
					Build(),
				bulletchartcomp.NewBuilder().
					SetID("bulletchart-latency-direction").
					Label("Latency Ceiling").
					Value(173).
					Target(200).
					Max(250).
					Width(22).
					LowerIsBetter().
					TargetMarkerRune('┆').
					BelowValueLabel().
					Build(),
				bulletchartcomp.NewBuilder().
					SetID("bulletchart-error-rate-direction").
					Label("Error Rate").
					Value(0).
					Target(5).
					Max(100).
					Width(22).
					LowerIsBetter().
					TargetMarkerRune('¦').
					BelowValueLabel().
					Build(),
			})
	}
}

func newChartsScatterPlotApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Charts ScatterPlot Fixture").Build(),
				scatterplotcomp.NewBuilder(nil).
					SetID("scatterplot-density").
					Title("ScatterPlot Density").
					Series(
						scatterplotcomp.Series{
							Name:   "API",
							Points: []scatterplotcomp.Point{{X: 1, Y: 2}, {X: 2, Y: 5}, {X: 4, Y: 7}, {X: 7, Y: 9}},
						},
						scatterplotcomp.Series{
							Name:   "Worker",
							Points: []scatterplotcomp.Point{{X: 2, Y: 3}, {X: 3, Y: 4}, {X: 6, Y: 6}, {X: 8, Y: 8}},
							Glyph:  '◆',
						},
					).
					Width(11).
					Height(5).
					ShowLegend(true).
					ShowGrid(true).
					ShowAxis(true).
					Build(),
			})
	}
}

func newChartsLineContinuityApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Charts Line Continuity Fixture").Build(),
				linechartcomp.NewBuilder([]float64{1, 9, 2, 8, 3, 7, 4, 6, 5}).
					SetID("linechart-continuity").
					Title("Line Continuity").
					Width(5).
					Height(4).
					ShowLegend(false).
					ShowGrid(false).
					ShowAxis(true).
					ShowPoints(true).
					Build(),
			})
	}
}

func newChartsLineDenseAxisApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Charts Line Dense Axis Fixture").Build(),
				linechartcomp.NewBuilder([]float64{1, 9, 2, 8, 3, 7}).
					SetID("linechart-dense-axis").
					Title("Line Dense Axis").
					Labels([]string{"03/24", "03/25", "03/26", "03/27", "03/28", "03/29"}).
					DenseAxisLabels().
					Width(11).
					Height(4).
					ShowLegend(false).
					ShowGrid(true).
					ShowAxis(true).
					ShowPoints(true).
					Build(),
			})
	}
}

func newChartsLineAxisModesApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Charts Line Axis Modes Fixture").Build(),
				linechartcomp.NewBuilder([]float64{1, 9, 2, 8, 3, 7}).
					SetID("linechart-axis-dense").
					Title("Line Axis Dense").
					Labels([]string{"03/24", "03/25", "03/26", "03/27", "03/28", "03/29"}).
					DenseAxisLabels().
					Width(11).
					Height(4).
					ShowLegend(false).
					ShowGrid(true).
					ShowAxis(true).
					ShowPoints(true).
					Build(),
				linechartcomp.NewBuilder([]float64{1, 9, 2, 8, 3, 7}).
					SetID("linechart-axis-sparse").
					Title("Line Axis Sparse").
					Labels([]string{"03/24", "03/25", "03/26", "03/27", "03/28", "03/29"}).
					SparseAxisLabels().
					Width(11).
					Height(4).
					ShowLegend(false).
					ShowGrid(true).
					ShowAxis(true).
					ShowPoints(true).
					Build(),
				linechartcomp.NewBuilder([]float64{1, 9, 2, 8, 3, 7}).
					SetID("linechart-axis-auto").
					Title("Line Axis Auto").
					Labels([]string{"03/24", "03/25", "03/26", "03/27", "03/28", "03/29"}).
					AutoAxisLabels().
					Width(11).
					Height(4).
					ShowLegend(false).
					ShowGrid(true).
					ShowAxis(true).
					ShowPoints(true).
					Build(),
			})
	}
}

func newChartsScatterPlotDomainApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Charts ScatterPlot Domain Fixture").Build(),
				scatterplotcomp.NewBuilder(nil).
					SetID("scatterplot-domain").
					Title("ScatterPlot Domain").
					Series(
						scatterplotcomp.Series{
							Name:   "API",
							Points: []scatterplotcomp.Point{{X: 2, Y: 3}, {X: 4, Y: 6}, {X: 8, Y: 9}},
						},
						scatterplotcomp.Series{
							Name:   "Worker",
							Points: []scatterplotcomp.Point{{X: 3, Y: 4}, {X: 6, Y: 5}, {X: 7, Y: 8}},
							Glyph:  '◆',
						},
					).
					Domain(scatterplotcomp.NewDomain(0, 10, 0, 12)).
					Width(11).
					Height(5).
					ShowLegend(true).
					ShowGrid(true).
					ShowAxis(true).
					Build(),
			})
	}
}

func newChartsScatterPlotReferenceApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Charts ScatterPlot Reference Fixture").Build(),
				scatterplotcomp.NewBuilder(nil).
					SetID("scatterplot-reference").
					Title("ScatterPlot References").
					Series(
						scatterplotcomp.Series{
							Name:   "API",
							Points: []scatterplotcomp.Point{{X: 2, Y: 2}, {X: 8, Y: 8}},
						},
						scatterplotcomp.Series{
							Name:   "Worker",
							Points: []scatterplotcomp.Point{{X: 3, Y: 7}, {X: 7, Y: 3}},
							Glyph:  '◆',
						},
					).
					Domain(scatterplotcomp.NewDomain(0, 10, 0, 10)).
					XReferenceLineLabeled(5, "Target").
					YReferenceLineLabeled(6, "Floor").
					Width(11).
					Height(5).
					ShowLegend(true).
					ShowGrid(false).
					ShowAxis(true).
					Build(),
			})
	}
}

func newChartsScatterPlotViewportApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Charts ScatterPlot Viewport Fixture").Build(),
				scatterplotcomp.NewBuilder(nil).
					SetID("scatterplot-viewport").
					Title("ScatterPlot Viewport").
					Series(
						scatterplotcomp.Series{
							Name:   "API",
							Points: []scatterplotcomp.Point{{X: 1, Y: 2}, {X: 4, Y: 5}, {X: 8, Y: 9}},
						},
						scatterplotcomp.Series{
							Name:   "Worker",
							Points: []scatterplotcomp.Point{{X: 2, Y: 3}, {X: 6, Y: 7}, {X: 9, Y: 10}},
							Glyph:  '◆',
						},
					).
					Domain(scatterplotcomp.NewDomain(0, 10, 0, 12)).
					Viewport(scatterplotcomp.NewViewport(2, 8, 3, 9)).
					Width(11).
					Height(5).
					ShowLegend(true).
					ShowGrid(true).
					ShowAxis(true).
					Build(),
			})
	}
}

func newChartsScatterPlotCollisionApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Charts ScatterPlot Collision Fixture").Build(),
				scatterplotcomp.NewBuilder(nil).
					SetID("scatterplot-collision").
					Title("ScatterPlot Collision").
					Series(
						scatterplotcomp.Series{
							Name:   "API",
							Points: []scatterplotcomp.Point{{X: 2.0, Y: 2.0}, {X: 2.1, Y: 2.1}, {X: 6.0, Y: 7.0}},
						},
						scatterplotcomp.Series{
							Name:   "Worker",
							Points: []scatterplotcomp.Point{{X: 2.0, Y: 2.0}, {X: 6.1, Y: 7.1}},
							Glyph:  '◆',
						},
					).
					Domain(scatterplotcomp.NewDomain(0, 10, 0, 10)).
					Width(7).
					Height(5).
					ShowLegend(true).
					ShowGrid(false).
					ShowAxis(true).
					Build(),
			})
	}
}

func newChartsScatterPlotBandApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Charts ScatterPlot Band Fixture").Build(),
				scatterplotcomp.NewBuilder(nil).
					SetID("scatterplot-band").
					Title("ScatterPlot Bands").
					Series(
						scatterplotcomp.Series{
							Name:   "API",
							Points: []scatterplotcomp.Point{{X: 2, Y: 3}, {X: 6, Y: 7}},
						},
						scatterplotcomp.Series{
							Name:   "Worker",
							Points: []scatterplotcomp.Point{{X: 4, Y: 5}, {X: 8, Y: 8}},
							Glyph:  '◆',
						},
					).
					Domain(scatterplotcomp.NewDomain(0, 10, 0, 10)).
					XReferenceBandLabeled(2, 4, "Focus").
					YReferenceBandLabeled(6, 8, "Risk").
					Width(11).
					Height(5).
					ShowLegend(true).
					ShowGrid(false).
					ShowAxis(true).
					Build(),
			})
	}
}

func newChartsCandlestickApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Charts Candlestick Fixture").Build(),
				candlestickcomp.NewBuilder([]candlestickcomp.Candle{
					{Label: "M", Open: 100, High: 110, Low: 95, Close: 107},
					{Label: "T", Open: 107, High: 112, Low: 101, Close: 103},
					{Label: "W", Open: 103, High: 116, Low: 100, Close: 103},
					{Label: "T", Open: 103, High: 109, Low: 98, Close: 108},
				}).
					SetID("candlestick-trend").
					Title("Candlestick Trend").
					Width(7).
					Height(6).
					ShowLegend(true).
					ShowGrid(true).
					ShowAxis(true).
					Build(),
			})
	}
}

func newChartsCandlestickVolumeApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Charts Candlestick Volume Fixture").Build(),
				candlestickcomp.NewBuilder([]candlestickcomp.Candle{
					{Label: "M", Open: 100, High: 110, Low: 95, Close: 107, Volume: 1800},
					{Label: "T", Open: 107, High: 112, Low: 101, Close: 103, Volume: 1200},
					{Label: "W", Open: 103, High: 116, Low: 100, Close: 103, Volume: 900},
					{Label: "T", Open: 103, High: 109, Low: 98, Close: 108, Volume: 1500},
				}).
					SetID("candlestick-volume").
					Title("Candlestick Volume").
					Width(7).
					Height(5).
					ShowLegend(true).
					ShowGrid(true).
					ShowAxis(true).
					ShowVolume(true).
					VolumeHeight(3).
					Build(),
			})
	}
}

func newChartsCandlestickDenseAxisApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Charts Candlestick Dense Axis Fixture").Build(),
				candlestickcomp.NewBuilder([]candlestickcomp.Candle{
					{Label: "03/24", Open: 100, High: 110, Low: 95, Close: 107},
					{Label: "03/25", Open: 107, High: 112, Low: 101, Close: 103},
					{Label: "03/26", Open: 103, High: 116, Low: 100, Close: 103},
					{Label: "03/27", Open: 103, High: 109, Low: 98, Close: 108},
					{Label: "03/28", Open: 108, High: 114, Low: 104, Close: 111},
					{Label: "03/29", Open: 111, High: 116, Low: 109, Close: 115},
				}).
					SetID("candlestick-dense-axis").
					Title("Candlestick Dense Axis").
					Width(11).
					Height(5).
					ShowLegend(false).
					ShowGrid(true).
					ShowAxis(true).
					Build(),
			})
	}
}

func newChartsCandlestickStyledApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Charts Candlestick Styled Fixture").Build(),
				candlestickcomp.NewBuilder([]candlestickcomp.Candle{
					{Label: "M", Open: 100, High: 110, Low: 95, Close: 107, Volume: 1800},
					{Label: "T", Open: 107, High: 112, Low: 101, Close: 103, Volume: 1200},
					{Label: "W", Open: 103, High: 116, Low: 100, Close: 103, Volume: 900},
				}).
					SetID("candlestick-styled").
					Title("Candlestick Styled").
					Width(5).
					Height(5).
					ShowLegend(true).
					ShowGrid(false).
					ShowAxis(false).
					ShowVolume(true).
					VolumeHeight(2).
					UpStyle(rtstyle.NewStyle().Foreground(rtstyle.Yellow).Bold(true)).
					DownStyle(rtstyle.NewStyle().Foreground(rtstyle.Magenta).Underline(true)).
					FlatStyle(rtstyle.NewStyle().Foreground(rtstyle.Cyan).Italic(true)).
					WickStyle(rtstyle.NewStyle().Foreground(rtstyle.BrightBlack).Bold(true)).
					VolumeStyle(rtstyle.NewStyle().Foreground(rtstyle.BrightWhite).Reverse(true)).
					Build(),
			})
	}
}

func TestE2EChartsLineAndBarRender(t *testing.T) {
	app, err := Run(newChartsStaticApp(), ui.WithSize(96, 40))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{
		"Charts E2E Fixture",
		"Latency Trend",
		"● API",
		"● Worker",
		"Throughput Bars",
		"█ Revenue",
		"█ Cost",
		" A  B  C",
		"Throughput Stacked",
		"█ North",
		"█ South",
		"X Y Z",
	} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestE2EChartsHorizontalBarRender(t *testing.T) {
	app, err := Run(newChartsHorizontalApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{
		"Charts Horizontal Fixture",
		"Throughput Horizontal",
		"█ Online",
		"█ Retail",
		"East",
		"West",
	} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestE2EChartsBarLabelFoldRender(t *testing.T) {
	app, err := Run(newChartsBarLabelFoldApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{
		"Charts Bar Label Fold Fixture",
		"Dense Vertical Labels",
		"N L A",
		"Dense Horizontal Labels",
		"NA",
		"LA",
	} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestE2EChartsValueLabelsRender(t *testing.T) {
	app, err := Run(newChartsValueLabelApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{
		"Charts Value Label Fixture",
		"Value Labels Vertical",
		"Values: 12 7 15",
		"Value Labels Horizontal",
		"12",
		"9",
	} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestE2EChartsHeatmapRender(t *testing.T) {
	app, err := Run(newChartsHeatmapApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{
		"Charts Heatmap Fixture",
		"Matrix Heatmap",
		"Low ░ ▒ ▓ █ High",
		"M T W T",
		"API",
		"DB",
		"Queue",
	} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestE2EChartsHeatmapCompactLegendRender(t *testing.T) {
	app, err := Run(newChartsHeatmapCompactLegendApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{
		"Charts Heatmap Compact Fixture",
		"Compact Legend Heatmap",
		"L ░▒▓█ H",
		"API",
		"DB",
	} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestE2EChartsHeatmapViewportRender(t *testing.T) {
	app, err := Run(newChartsHeatmapViewportApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{
		"Charts Heatmap Viewport Fixture",
		"Viewport Heatmap",
		"Low ░ ▒ ▓ █ High",
		"T W T",
		"Sou~",
		"Eur~",
	} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestE2EChartsHeatmapViewportScaleRender(t *testing.T) {
	app, err := Run(newChartsHeatmapViewportScaleApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{
		"Charts Heatmap Contrast Fixture",
		"Viewport Contrast Heatmap",
		"Sou~",
		"Eur~",
	} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestE2EChartsHeatmapAutoScaleRender(t *testing.T) {
	app, err := Run(newChartsHeatmapAutoScaleApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{
		"Charts Heatmap Auto Scale Fixture",
		"Auto Contrast Heatmap",
		"Sou~",
		"Eur~",
	} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestE2EChartsHeatmapAutoThresholdRender(t *testing.T) {
	app, err := Run(newChartsHeatmapAutoThresholdApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{
		"Charts Heatmap Auto Threshold Fixture",
		"Auto Threshold Heatmap",
		"M T W",
		"North",
		"South",
	} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestE2EChartsHeatmapSummaryRender(t *testing.T) {
	app, err := Run(newChartsHeatmapSummaryApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{
		"Charts Heatmap Summary Fixture",
		"Summary Heatmap",
		"range 40..43 avg 41.5",
		"Sou~",
		"Eur~",
	} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestE2EChartsHeatmapCompactSummaryRender(t *testing.T) {
	app, err := Run(newChartsHeatmapCompactSummaryApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{
		"Charts Heatmap Compact Summary Fixture",
		"Compact Summary Heatmap",
		"40..43 avg 41.5",
		"Sou~",
		"Eur~",
	} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestE2EChartsSparklineRender(t *testing.T) {
	app, err := Run(newChartsSparklineApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{
		"Charts Sparkline Fixture",
		"Requests Sparkline",
		"Budget Sparkline",
		"Errors Sparkline",
		"Auto Compact Sparkline",
		"⣀⣄⣄⣄⣆⣇⣧⣷⣧⣧⣷⣿",
		"██▆▆▆▆▃▃▃▃▁▁",
		".:==-+*+##*@",
		"+@+.",
	} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestE2EChartsSparklineEnhancedRender(t *testing.T) {
	app, err := Run(newChartsSparklineEnhancedApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{
		"Charts Sparkline Enhanced Fixture",
		"Latency Sparkline",
		"live",
	} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestE2EChartsBulletChartRender(t *testing.T) {
	app, err := Run(newChartsBulletChartApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{
		"Charts Bullet Fixture",
		"Latency: 173/250 target 200",
		"Availability: 996/1000 target 999",
		"╻",
		"┆",
		"▒",
		"▓",
	} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestE2EChartsBulletChartDirectionRender(t *testing.T) {
	app, err := Run(newChartsBulletChartDirectionApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{
		"Charts Bullet Direction Fixture",
		"Throughput: 82/100 target 75",
		"Latency Ceiling: 173/250 target 200",
		"Error Rate: 0/100 target 5",
		"╻",
		"┆",
		"¦",
		"▒",
		"▓",
	} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestE2EChartsBulletChartStyles(t *testing.T) {
	app, err := Run(newChartsBulletChartApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertStyle(ByText("╻"), StyleExpect{
		HasFG:   true,
		FG:      fwtheme.Warning(),
		HasBold: true,
		Bold:    true,
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("▒"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Secondary(),
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("▓"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Text(),
	}); err != nil {
		t.Fatal(err)
	}
}

func TestE2EChartsBulletChartDirectionStyles(t *testing.T) {
	app, err := Run(newChartsBulletChartDirectionApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertStyle(ByText("╻"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Success(),
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("╻"), StyleExpect{
		HasBold: true,
		Bold:    true,
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("┆"), StyleExpect{
		HasFG:   true,
		FG:      fwtheme.Error(),
		HasBold: true,
		Bold:    true,
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("▓"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Success(),
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("¦"), StyleExpect{
		HasFG:   true,
		FG:      fwtheme.Error(),
		HasBold: true,
		Bold:    true,
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("░"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Success(),
	}); err != nil {
		t.Fatal(err)
	}
}

func TestE2EChartsScatterPlotRender(t *testing.T) {
	app, err := Run(newChartsScatterPlotApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{
		"Charts ScatterPlot Fixture",
		"ScatterPlot Density",
		"● API",
		"◆ Worker",
		"x:1..8 y:2..9",
	} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestE2EChartsLineContinuityRender(t *testing.T) {
	app, err := Run(newChartsLineContinuityApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{
		"Charts Line Continuity Fixture",
		"Line Continuity",
		"─────",
	} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestE2EChartsLineDenseAxisRender(t *testing.T) {
	app, err := Run(newChartsLineDenseAxisApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{
		"Charts Line Dense Axis Fixture",
		"Line Dense Axis",
		"4 5 6 7 8 9",
		"───────────",
	} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestE2EChartsLineAxisModesRender(t *testing.T) {
	app, err := Run(newChartsLineAxisModesApp(), ui.WithSize(96, 40))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{
		"Charts Line Axis Modes Fixture",
		"Line Axis Dense",
		"Line Axis Sparse",
		"Line Axis Auto",
		"4 5 6 7 8 9",
		"4   6     9",
	} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestE2EChartsScatterPlotDomainRender(t *testing.T) {
	app, err := Run(newChartsScatterPlotDomainApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{
		"Charts ScatterPlot Domain Fixture",
		"ScatterPlot Domain",
		"● API",
		"◆ Worker",
		"x:0..10 y:0..12",
	} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestE2EChartsScatterPlotReferenceRender(t *testing.T) {
	app, err := Run(newChartsScatterPlotReferenceApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{
		"Charts ScatterPlot Reference Fixture",
		"ScatterPlot References",
		"● API",
		"◆ Worker",
		"│ x: Target",
		"─ y: Floor",
		"x:0..10 y:0..10",
		"│",
	} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestE2EChartsScatterPlotViewportRender(t *testing.T) {
	app, err := Run(newChartsScatterPlotViewportApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{
		"Charts ScatterPlot Viewport Fixture",
		"ScatterPlot Viewport",
		"● API",
		"◆ Worker",
		"x:2..8 y:3..9",
	} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestE2EChartsScatterPlotCollisionRender(t *testing.T) {
	app, err := Run(newChartsScatterPlotCollisionApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{
		"Charts ScatterPlot Collision Fixture",
		"ScatterPlot Collision",
		"● API",
		"◆ Worker",
		"3",
		"2",
		"x:0..10 y:0..10",
	} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestE2EChartsScatterPlotBandRender(t *testing.T) {
	app, err := Run(newChartsScatterPlotBandApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{
		"Charts ScatterPlot Band Fixture",
		"ScatterPlot Bands",
		"● API",
		"◆ Worker",
		"░ x: Focus",
		"░ y: Risk",
		"░",
		"▒",
		"x:0..10 y:0..10",
	} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestE2EChartsCandlestickRender(t *testing.T) {
	app, err := Run(newChartsCandlestickApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{
		"Charts Candlestick Fixture",
		"Candlestick Trend",
		"▓ Up",
		"█ Down",
		"■ Flat",
		"M T W T",
	} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestE2EChartsCandlestickVolumeRender(t *testing.T) {
	app, err := Run(newChartsCandlestickVolumeApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{
		"Charts Candlestick Volume Fixture",
		"Candlestick Volume",
		"▓ Up",
		"█ Down",
		"■ Flat",
		"▆ Volume",
		"M T W T",
	} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestE2EChartsCandlestickDenseAxisRender(t *testing.T) {
	app, err := Run(newChartsCandlestickDenseAxisApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{
		"Charts Candlestick Dense Axis Fixture",
		"Candlestick Dense Axis",
		"4 5 6 7 8 9",
	} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestE2EChartsCandlestickStyledRender(t *testing.T) {
	app, err := Run(newChartsCandlestickStyledApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{
		"Charts Candlestick Styled Fixture",
		"Candlestick Styled",
		"▓ Up",
		"█ Down",
		"■ Flat",
		"▆ Volume",
	} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestE2EChartsLegendStyles(t *testing.T) {
	app, err := Run(newChartsStaticApp(), ui.WithSize(96, 40))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertStyle(ByText("● API"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Primary(),
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("● Worker"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Accent(),
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("█ Revenue"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Primary(),
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("█ Cost"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Accent(),
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("█ North"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Primary(),
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("█ South"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Accent(),
	}); err != nil {
		t.Fatal(err)
	}
}

func TestE2EChartsHorizontalLegendStyles(t *testing.T) {
	app, err := Run(newChartsHorizontalApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertStyle(ByText("█ Online"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Primary(),
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("█ Retail"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Accent(),
	}); err != nil {
		t.Fatal(err)
	}
}

func TestE2EChartsBarLabelFoldStyles(t *testing.T) {
	app, err := Run(newChartsBarLabelFoldApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertStyle(ByText("N L A"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Muted(),
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("NA"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Muted(),
	}); err != nil {
		t.Fatal(err)
	}
}

func TestE2EChartsLineDenseAxisStyles(t *testing.T) {
	app, err := Run(newChartsLineDenseAxisApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertStyle(ByText("4 5 6 7 8 9"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Muted(),
	}); err != nil {
		t.Fatal(err)
	}
}

func TestE2EChartsLineAxisModesStyles(t *testing.T) {
	app, err := Run(newChartsLineAxisModesApp(), ui.WithSize(96, 40))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertStyle(ByText("4 5 6 7 8 9"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Muted(),
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("4   6     9"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Muted(),
	}); err != nil {
		t.Fatal(err)
	}
}

func TestE2EChartsScatterPlotLegendStyles(t *testing.T) {
	app, err := Run(newChartsScatterPlotApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertStyle(ByText("● API"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Primary(),
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("◆ Worker"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Accent(),
	}); err != nil {
		t.Fatal(err)
	}
}

func TestE2EChartsScatterPlotReferenceStyles(t *testing.T) {
	app, err := Run(newChartsScatterPlotReferenceApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertStyle(ByText("│"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Warning(),
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("│ x: Target"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Warning(),
	}); err != nil {
		t.Fatal(err)
	}
}

func TestE2EChartsScatterPlotCollisionStyles(t *testing.T) {
	app, err := Run(newChartsScatterPlotCollisionApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertStyle(ByText("3"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Warning(),
	}); err != nil {
		t.Fatal(err)
	}
}

func TestE2EChartsScatterPlotBandStyles(t *testing.T) {
	app, err := Run(newChartsScatterPlotBandApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertStyle(ByText("░"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Secondary(),
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("░ x: Focus"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Secondary(),
	}); err != nil {
		t.Fatal(err)
	}
}

func TestE2EChartsCandlestickLegendStyles(t *testing.T) {
	app, err := Run(newChartsCandlestickApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertStyle(ByText("▓ Up"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Success(),
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("█ Down"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Error(),
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("■ Flat"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Secondary(),
	}); err != nil {
		t.Fatal(err)
	}
}

func TestE2EChartsCandlestickVolumeLegendStyles(t *testing.T) {
	app, err := Run(newChartsCandlestickVolumeApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertStyle(ByText("▆ Volume"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Muted(),
	}); err != nil {
		t.Fatal(err)
	}
}

func TestE2EChartsCandlestickStyledStyles(t *testing.T) {
	app, err := Run(newChartsCandlestickStyledApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertStyle(ByText("▓ Up"), StyleExpect{
		HasFG:   true,
		FG:      rtstyle.Yellow,
		HasBold: true,
		Bold:    true,
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("█ Down"), StyleExpect{
		HasFG:        true,
		FG:           rtstyle.Magenta,
		HasUnderline: true,
		Underline:    true,
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("■ Flat"), StyleExpect{
		HasFG:     true,
		FG:        rtstyle.Cyan,
		HasItalic: true,
		Italic:    true,
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("▆ Volume"), StyleExpect{
		HasFG:      true,
		FG:         rtstyle.BrightWhite,
		HasReverse: true,
		Reverse:    true,
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("│"), StyleExpect{
		HasFG:   true,
		FG:      rtstyle.BrightBlack,
		HasBold: true,
		Bold:    true,
	}); err != nil {
		t.Fatal(err)
	}
}

func TestE2EChartsLineSnapshot(t *testing.T) {
	app, err := Run(newChartsStaticApp(), ui.WithSize(96, 40))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	defer func() {
		_, _ = app.SaveDiagnosticsOnFailure(t, "mint-e2e-charts-line-")
	}()

	assertRenderSnapshot(t, app, "charts_static_96x40.render.txt")
}

func TestE2EChartsHeatmapSnapshot(t *testing.T) {
	app, err := Run(newChartsHeatmapApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	defer func() {
		_, _ = app.SaveDiagnosticsOnFailure(t, "mint-e2e-charts-heatmap-")
	}()

	assertRenderSnapshot(t, app, "charts_heatmap_96x24.render.txt")
}

func TestE2EChartsHeatmapViewportSnapshot(t *testing.T) {
	app, err := Run(newChartsHeatmapViewportApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	defer func() {
		_, _ = app.SaveDiagnosticsOnFailure(t, "mint-e2e-charts-heatmap-viewport-")
	}()

	assertRenderSnapshot(t, app, "charts_heatmap_viewport_96x24.render.txt")
}

func TestE2EChartsHeatmapViewportScaleSnapshot(t *testing.T) {
	app, err := Run(newChartsHeatmapViewportScaleApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	defer func() {
		_, _ = app.SaveDiagnosticsOnFailure(t, "mint-e2e-charts-heatmap-viewport-scale-")
	}()

	assertRenderSnapshot(t, app, "charts_heatmap_viewport_scale_96x24.render.txt")
}

func TestE2EChartsHeatmapAutoScaleSnapshot(t *testing.T) {
	app, err := Run(newChartsHeatmapAutoScaleApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	defer func() {
		_, _ = app.SaveDiagnosticsOnFailure(t, "mint-e2e-charts-heatmap-auto-scale-")
	}()

	assertRenderSnapshot(t, app, "charts_heatmap_auto_scale_96x24.render.txt")
}

func TestE2EChartsHeatmapAutoThresholdSnapshot(t *testing.T) {
	app, err := Run(newChartsHeatmapAutoThresholdApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	defer func() {
		_, _ = app.SaveDiagnosticsOnFailure(t, "mint-e2e-charts-heatmap-auto-threshold-")
	}()

	assertRenderSnapshot(t, app, "charts_heatmap_auto_threshold_96x24.render.txt")
}

func TestE2EChartsHeatmapSummarySnapshot(t *testing.T) {
	app, err := Run(newChartsHeatmapSummaryApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	defer func() {
		_, _ = app.SaveDiagnosticsOnFailure(t, "mint-e2e-charts-heatmap-summary-")
	}()

	assertRenderSnapshot(t, app, "charts_heatmap_summary_96x24.render.txt")
}

func TestE2EChartsHeatmapCompactSummarySnapshot(t *testing.T) {
	app, err := Run(newChartsHeatmapCompactSummaryApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	defer func() {
		_, _ = app.SaveDiagnosticsOnFailure(t, "mint-e2e-charts-heatmap-compact-summary-")
	}()

	assertRenderSnapshot(t, app, "charts_heatmap_compact_summary_96x24.render.txt")
}

func TestE2EChartsHorizontalSnapshot(t *testing.T) {
	app, err := Run(newChartsHorizontalApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	defer func() {
		_, _ = app.SaveDiagnosticsOnFailure(t, "mint-e2e-charts-horizontal-")
	}()

	assertRenderSnapshot(t, app, "charts_horizontal_96x24.render.txt")
}

func TestE2EChartsBarLabelFoldSnapshot(t *testing.T) {
	app, err := Run(newChartsBarLabelFoldApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	defer func() {
		_, _ = app.SaveDiagnosticsOnFailure(t, "mint-e2e-charts-bar-label-fold-")
	}()

	assertRenderSnapshot(t, app, "charts_barchart_label_fold_96x24.render.txt")
}

func TestE2EChartsValueLabelsSnapshot(t *testing.T) {
	app, err := Run(newChartsValueLabelApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	defer func() {
		_, _ = app.SaveDiagnosticsOnFailure(t, "mint-e2e-charts-values-")
	}()

	assertRenderSnapshot(t, app, "charts_value_labels_96x24.render.txt")
}

func TestE2EChartsHeatmapCompactLegendSnapshot(t *testing.T) {
	app, err := Run(newChartsHeatmapCompactLegendApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	defer func() {
		_, _ = app.SaveDiagnosticsOnFailure(t, "mint-e2e-charts-heatmap-compact-")
	}()

	assertRenderSnapshot(t, app, "charts_heatmap_compact_96x24.render.txt")
}

func TestE2EChartsSparklineSnapshot(t *testing.T) {
	app, err := Run(newChartsSparklineApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	defer func() {
		_, _ = app.SaveDiagnosticsOnFailure(t, "mint-e2e-charts-sparkline-")
	}()

	assertRenderSnapshot(t, app, "charts_sparkline_96x24.render.txt")
}

func TestE2EChartsSparklineEnhancedSnapshot(t *testing.T) {
	app, err := Run(newChartsSparklineEnhancedApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	defer func() {
		_, _ = app.SaveDiagnosticsOnFailure(t, "mint-e2e-charts-sparkline-enhanced-")
	}()

	assertRenderSnapshot(t, app, "charts_sparkline_enhanced_96x24.render.txt")
}

func TestE2EChartsBulletChartSnapshot(t *testing.T) {
	app, err := Run(newChartsBulletChartApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	defer func() {
		_, _ = app.SaveDiagnosticsOnFailure(t, "mint-e2e-charts-bullet-")
	}()

	assertRenderSnapshot(t, app, "charts_bulletchart_96x24.render.txt")
}

func TestE2EChartsBulletChartDirectionSnapshot(t *testing.T) {
	app, err := Run(newChartsBulletChartDirectionApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	defer func() {
		_, _ = app.SaveDiagnosticsOnFailure(t, "mint-e2e-charts-bullet-direction-")
	}()

	assertRenderSnapshot(t, app, "charts_bulletchart_direction_96x24.render.txt")
}

func TestE2EChartsScatterPlotSnapshot(t *testing.T) {
	app, err := Run(newChartsScatterPlotApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	defer func() {
		_, _ = app.SaveDiagnosticsOnFailure(t, "mint-e2e-charts-scatterplot-")
	}()

	assertRenderSnapshot(t, app, "charts_scatterplot_96x24.render.txt")
}

func TestE2EChartsLineContinuitySnapshot(t *testing.T) {
	app, err := Run(newChartsLineContinuityApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	defer func() {
		_, _ = app.SaveDiagnosticsOnFailure(t, "mint-e2e-charts-line-continuity-")
	}()

	assertRenderSnapshot(t, app, "charts_line_continuity_96x24.render.txt")
}

func TestE2EChartsLineDenseAxisSnapshot(t *testing.T) {
	app, err := Run(newChartsLineDenseAxisApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	defer func() {
		_, _ = app.SaveDiagnosticsOnFailure(t, "mint-e2e-charts-line-dense-axis-")
	}()

	assertRenderSnapshot(t, app, "charts_line_dense_axis_96x24.render.txt")
}

func TestE2EChartsLineAxisModesSnapshot(t *testing.T) {
	app, err := Run(newChartsLineAxisModesApp(), ui.WithSize(96, 40))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	defer func() {
		_, _ = app.SaveDiagnosticsOnFailure(t, "mint-e2e-charts-line-axis-modes-")
	}()

	assertRenderSnapshot(t, app, "charts_line_axis_modes_96x40.render.txt")
}

func TestE2EChartsScatterPlotDomainSnapshot(t *testing.T) {
	app, err := Run(newChartsScatterPlotDomainApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	defer func() {
		_, _ = app.SaveDiagnosticsOnFailure(t, "mint-e2e-charts-scatterplot-domain-")
	}()

	assertRenderSnapshot(t, app, "charts_scatterplot_domain_96x24.render.txt")
}

func TestE2EChartsScatterPlotReferenceSnapshot(t *testing.T) {
	app, err := Run(newChartsScatterPlotReferenceApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	defer func() {
		_, _ = app.SaveDiagnosticsOnFailure(t, "mint-e2e-charts-scatterplot-reference-")
	}()

	assertRenderSnapshot(t, app, "charts_scatterplot_reference_96x24.render.txt")
}

func TestE2EChartsScatterPlotViewportSnapshot(t *testing.T) {
	app, err := Run(newChartsScatterPlotViewportApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	defer func() {
		_, _ = app.SaveDiagnosticsOnFailure(t, "mint-e2e-charts-scatterplot-viewport-")
	}()

	assertRenderSnapshot(t, app, "charts_scatterplot_viewport_96x24.render.txt")
}

func TestE2EChartsScatterPlotCollisionSnapshot(t *testing.T) {
	app, err := Run(newChartsScatterPlotCollisionApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	defer func() {
		_, _ = app.SaveDiagnosticsOnFailure(t, "mint-e2e-charts-scatterplot-collision-")
	}()

	assertRenderSnapshot(t, app, "charts_scatterplot_collision_96x24.render.txt")
}

func TestE2EChartsScatterPlotBandSnapshot(t *testing.T) {
	app, err := Run(newChartsScatterPlotBandApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	defer func() {
		_, _ = app.SaveDiagnosticsOnFailure(t, "mint-e2e-charts-scatterplot-band-")
	}()

	assertRenderSnapshot(t, app, "charts_scatterplot_band_96x24.render.txt")
}

func TestE2EChartsCandlestickSnapshot(t *testing.T) {
	app, err := Run(newChartsCandlestickApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	defer func() {
		_, _ = app.SaveDiagnosticsOnFailure(t, "mint-e2e-charts-candlestick-")
	}()

	assertRenderSnapshot(t, app, "charts_candlestick_96x24.render.txt")
}

func TestE2EChartsCandlestickVolumeSnapshot(t *testing.T) {
	app, err := Run(newChartsCandlestickVolumeApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	defer func() {
		_, _ = app.SaveDiagnosticsOnFailure(t, "mint-e2e-charts-candlestick-volume-")
	}()

	assertRenderSnapshot(t, app, "charts_candlestick_volume_96x24.render.txt")
}

func TestE2EChartsCandlestickDenseAxisSnapshot(t *testing.T) {
	app, err := Run(newChartsCandlestickDenseAxisApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	defer func() {
		_, _ = app.SaveDiagnosticsOnFailure(t, "mint-e2e-charts-candlestick-dense-axis-")
	}()

	assertRenderSnapshot(t, app, "charts_candlestick_dense_axis_96x24.render.txt")
}

func TestE2EChartsCandlestickStyledSnapshot(t *testing.T) {
	app, err := Run(newChartsCandlestickStyledApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	defer func() {
		_, _ = app.SaveDiagnosticsOnFailure(t, "mint-e2e-charts-candlestick-styled-")
	}()

	assertRenderSnapshot(t, app, "charts_candlestick_styled_96x24.render.txt")
}
