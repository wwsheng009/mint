package ui

import (
	fwtheme "github.com/wwsheng009/mint/framework/theme"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/charts/barchart"
	"github.com/wwsheng009/mint/ui/components/charts/bulletchart"
	"github.com/wwsheng009/mint/ui/components/charts/candlestick"
	"github.com/wwsheng009/mint/ui/components/charts/heatmap"
	"github.com/wwsheng009/mint/ui/components/charts/linechart"
	"github.com/wwsheng009/mint/ui/components/charts/scatterplot"
	"github.com/wwsheng009/mint/ui/components/charts/sparkline"
)

// Chart builder factories.

func NewSparklineBuilder(data []float64) *sparkline.Builder {
	return sparkline.NewBuilder(data)
}

func NewBulletChartBuilder() *bulletchart.Builder {
	return bulletchart.NewBuilder()
}

func NewBarChartBuilder(values []float64) *barchart.Builder {
	return barchart.NewBuilder(values)
}

func NewLineChartBuilder(data []float64) *linechart.Builder {
	return linechart.NewBuilder(data)
}

func NewHeatmapBuilder(values [][]float64) *heatmap.Builder {
	return heatmap.NewBuilder(values)
}

func NewScatterPlotBuilder(points []scatterplot.Point) *scatterplot.Builder {
	return scatterplot.NewBuilder(points)
}

func NewCandlestickBuilder(candles []candlestick.Candle) *candlestick.Builder {
	return candlestick.NewBuilder(candles)
}

// Chart shortcuts.

func Sparkline(data []float64) rtui.VNode {
	return sparkline.NewBuilder(data).Build()
}

func ASCIISparkline(data []float64) rtui.VNode {
	return sparkline.NewBuilder(data).
		ASCII().
		Build()
}

func BulletChart(value, target, max int) rtui.VNode {
	return bulletchart.NewBuilder().
		Value(value).
		Target(target).
		Max(max).
		Build()
}

func BarChart(labels []string, values []float64) rtui.VNode {
	return barchart.NewBuilder(values).
		Labels(labels).
		Build()
}

func ASCIIBarChart(labels []string, values []float64) rtui.VNode {
	return barchart.NewBuilder(values).
		Labels(labels).
		ASCII().
		Build()
}

// HorizontalASCIIBarChart creates a keyed horizontal ASCII bar chart with value labels.
func HorizontalASCIIBarChart(key string, labels []string, values []float64, width, height int) rtui.VNode {
	builder := barchart.NewBuilder(values).
		Key(key).
		Labels(labels).
		ASCII().
		Horizontal().
		ShowValue(true)
	if width > 0 {
		builder.Width(width)
	}
	if height > 0 {
		builder.Height(height)
	}
	return builder.Build()
}

// HorizontalASCIIBarChartPanel creates a titled table-style panel containing a
// keyed horizontal ASCII bar chart with value labels.
func HorizontalASCIIBarChartPanel(title, key string, labels []string, values []float64, panelWidth, chartWidth, chartHeight int) rtui.VNode {
	return TablePanel(title, HorizontalASCIIBarChart(key, labels, values, chartWidth, chartHeight), panelWidth)
}

func LineChart(data []float64) rtui.VNode {
	return linechart.NewBuilder(data).Build()
}

func Heatmap(values [][]float64) rtui.VNode {
	return heatmap.NewBuilder(values).Build()
}

func ScatterPlot(points []scatterplot.Point) rtui.VNode {
	return scatterplot.NewBuilder(points).Build()
}

func Candlestick(candles []candlestick.Candle) rtui.VNode {
	return candlestick.NewBuilder(candles).Build()
}

// Chart type re-exports.

type SparklineRenderMode = sparkline.RenderMode
type BulletChartQualitativeRange = bulletchart.QualitativeRange
type BulletChartValueLabelMode = bulletchart.ValueLabelMode
type BulletChartDirection = bulletchart.Direction
type LineChartSeries = linechart.Series
type BarChartSeries = barchart.Series
type BarChartMode = barchart.Mode
type BarChartOrientation = barchart.Orientation
type BarChartRenderMode = barchart.RenderMode
type ScatterPlotPoint = scatterplot.Point
type ScatterPlotSeries = scatterplot.Series
type ScatterPlotDomain = scatterplot.Domain
type ScatterPlotViewport = scatterplot.Viewport
type ScatterPlotReferenceLine = scatterplot.ReferenceLine
type ScatterPlotReferenceBand = scatterplot.ReferenceBand
type CandlestickCandle = candlestick.Candle
type HeatmapColorMode = fwtheme.ColorMode
type HeatmapLegendMode = heatmap.LegendMode
type HeatmapSummaryMode = heatmap.SummaryMode
type HeatmapScaleMode = heatmap.ScaleMode
type HeatmapViewport = heatmap.Viewport

const (
	SparklineRenderModeAuto    = sparkline.RenderModeAuto
	SparklineRenderModeBraille = sparkline.RenderModeBraille
	SparklineRenderModeBlock   = sparkline.RenderModeBlock
	SparklineRenderModeASCII   = sparkline.RenderModeASCII

	BulletChartValueLabelModeAuto   = bulletchart.ValueLabelModeAuto
	BulletChartValueLabelModeInline = bulletchart.ValueLabelModeInline
	BulletChartValueLabelModeBelow  = bulletchart.ValueLabelModeBelow

	BulletChartDirectionNeutral      = bulletchart.DirectionNeutral
	BulletChartDirectionHigherBetter = bulletchart.DirectionHigherBetter
	BulletChartDirectionLowerBetter  = bulletchart.DirectionLowerBetter

	BarChartModeGrouped           = barchart.ModeGrouped
	BarChartModeStacked           = barchart.ModeStacked
	BarChartOrientationVertical   = barchart.OrientationVertical
	BarChartOrientationHorizontal = barchart.OrientationHorizontal
	BarChartRenderModeBlock       = barchart.RenderModeBlock
	BarChartRenderModeASCII       = barchart.RenderModeASCII

	HeatmapColorModeTrueColor = fwtheme.ColorModeTrueColor
	HeatmapColorMode256       = fwtheme.ColorMode256
	HeatmapColorMode16        = fwtheme.ColorMode16
	HeatmapColorModeNone      = fwtheme.ColorModeNone

	HeatmapLegendModeFull    = heatmap.LegendModeFull
	HeatmapLegendModeCompact = heatmap.LegendModeCompact

	HeatmapSummaryModeNone     = heatmap.SummaryModeNone
	HeatmapSummaryModeCompact  = heatmap.SummaryModeCompact
	HeatmapSummaryModeDetailed = heatmap.SummaryModeDetailed

	HeatmapScaleModeGlobal   = heatmap.ScaleModeGlobal
	HeatmapScaleModeViewport = heatmap.ScaleModeViewport
	HeatmapScaleModeAuto     = heatmap.ScaleModeAuto
)
