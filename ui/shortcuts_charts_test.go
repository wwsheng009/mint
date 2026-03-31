package ui

import "testing"

func TestChartBuilderFactories(t *testing.T) {
	if vnode := NewSparklineBuilder([]float64{1, 2, 3}).Title("Trend").Build(); vnode == nil {
		t.Fatal("NewSparklineBuilder().Build() returned nil")
	}
	if vnode := NewBulletChartBuilder().Value(60).Target(75).Max(100).QualitativeRanges(
		BulletChartQualitativeRange{Limit: 40, Glyph: '░'},
		BulletChartQualitativeRange{Limit: 70, Glyph: '▒'},
		BulletChartQualitativeRange{Limit: 100, Glyph: '▓'},
	).BelowValueLabel().TargetMarkerRune('╻').Build(); vnode == nil {
		t.Fatal("NewBulletChartBuilder().Build() returned nil")
	}
	if vnode := NewBarChartBuilder([]float64{1, 2, 3}).Labels([]string{"A", "B", "C"}).Horizontal().Build(); vnode == nil {
		t.Fatal("NewBarChartBuilder().Build() returned nil")
	}
	if vnode := NewLineChartBuilder([]float64{1, 2, 3}).Title("Trend").Build(); vnode == nil {
		t.Fatal("NewLineChartBuilder().Build() returned nil")
	}
	if vnode := NewHeatmapBuilder([][]float64{{1, 2}, {3, 4}}).Title("Heat").Build(); vnode == nil {
		t.Fatal("NewHeatmapBuilder().Build() returned nil")
	}
	if vnode := NewScatterPlotBuilder([]ScatterPlotPoint{{X: 1, Y: 2}, {X: 3, Y: 4}}).Title("Scatter").XDomain(0, 10).YDomain(0, 12).XViewport(2, 8).YViewport(1, 9).XReferenceLineLabeled(5, "Target").YReferenceLineLabeled(6, "Floor").XReferenceBandLabeled(2, 4, "Focus").YReferenceBandLabeled(3, 6, "Risk").Build(); vnode == nil {
		t.Fatal("NewScatterPlotBuilder().Build() returned nil")
	}
	if vnode := NewCandlestickBuilder([]CandlestickCandle{{Label: "M", Open: 1, High: 3, Low: 0, Close: 2}}).Title("Candles").Build(); vnode == nil {
		t.Fatal("NewCandlestickBuilder().Build() returned nil")
	}
}

func TestChartShortcuts(t *testing.T) {
	if vnode := Sparkline([]float64{1, 2, 3}); vnode.Tag() != "sparkline" {
		t.Fatalf("Sparkline().Tag() = %q, want sparkline", vnode.Tag())
	}
	if vnode := BulletChart(60, 75, 100); vnode.Tag() != "bulletchart" {
		t.Fatalf("BulletChart().Tag() = %q, want bulletchart", vnode.Tag())
	}
	if vnode := BarChart([]string{"A", "B"}, []float64{1, 2}); vnode.Tag() != "barchart" {
		t.Fatalf("BarChart().Tag() = %q, want barchart", vnode.Tag())
	}
	if vnode := LineChart([]float64{1, 2, 3}); vnode.Tag() != "linechart" {
		t.Fatalf("LineChart().Tag() = %q, want linechart", vnode.Tag())
	}
	if vnode := Heatmap([][]float64{{1, 2}, {3, 4}}); vnode.Tag() != "heatmap" {
		t.Fatalf("Heatmap().Tag() = %q, want heatmap", vnode.Tag())
	}
	if vnode := ScatterPlot([]ScatterPlotPoint{{X: 1, Y: 2}, {X: 3, Y: 4}}); vnode.Tag() != "scatterplot" {
		t.Fatalf("ScatterPlot().Tag() = %q, want scatterplot", vnode.Tag())
	}
	if vnode := Candlestick([]CandlestickCandle{{Label: "M", Open: 1, High: 3, Low: 0, Close: 2}}); vnode.Tag() != "candlestick" {
		t.Fatalf("Candlestick().Tag() = %q, want candlestick", vnode.Tag())
	}
}

func TestHeatmapColorModeAliases(t *testing.T) {
	if HeatmapColorModeTrueColor != 0 {
		t.Fatalf("HeatmapColorModeTrueColor = %d, want 0", HeatmapColorModeTrueColor)
	}
	if HeatmapColorMode256 != 1 {
		t.Fatalf("HeatmapColorMode256 = %d, want 1", HeatmapColorMode256)
	}
	if HeatmapColorMode16 != 2 {
		t.Fatalf("HeatmapColorMode16 = %d, want 2", HeatmapColorMode16)
	}
	if HeatmapColorModeNone != 3 {
		t.Fatalf("HeatmapColorModeNone = %d, want 3", HeatmapColorModeNone)
	}
}

func TestHeatmapLegendModeAliases(t *testing.T) {
	if HeatmapLegendModeFull != 0 {
		t.Fatalf("HeatmapLegendModeFull = %d, want 0", HeatmapLegendModeFull)
	}
	if HeatmapLegendModeCompact != 1 {
		t.Fatalf("HeatmapLegendModeCompact = %d, want 1", HeatmapLegendModeCompact)
	}
}

func TestHeatmapSummaryModeAliases(t *testing.T) {
	if HeatmapSummaryModeNone != 0 {
		t.Fatalf("HeatmapSummaryModeNone = %d, want 0", HeatmapSummaryModeNone)
	}
	if HeatmapSummaryModeCompact != 1 {
		t.Fatalf("HeatmapSummaryModeCompact = %d, want 1", HeatmapSummaryModeCompact)
	}
	if HeatmapSummaryModeDetailed != 2 {
		t.Fatalf("HeatmapSummaryModeDetailed = %d, want 2", HeatmapSummaryModeDetailed)
	}
}

func TestHeatmapScaleModeAliases(t *testing.T) {
	if HeatmapScaleModeGlobal != 0 {
		t.Fatalf("HeatmapScaleModeGlobal = %d, want 0", HeatmapScaleModeGlobal)
	}
	if HeatmapScaleModeViewport != 1 {
		t.Fatalf("HeatmapScaleModeViewport = %d, want 1", HeatmapScaleModeViewport)
	}
	if HeatmapScaleModeAuto != 2 {
		t.Fatalf("HeatmapScaleModeAuto = %d, want 2", HeatmapScaleModeAuto)
	}
}

func TestHeatmapViewportAlias(t *testing.T) {
	viewport := HeatmapViewport{RowStart: 1, RowCount: 3, ColStart: 2, ColCount: 4}
	if viewport.RowStart != 1 || viewport.RowCount != 3 || viewport.ColStart != 2 || viewport.ColCount != 4 {
		t.Fatalf("HeatmapViewport = %+v, want RowStart=1 RowCount=3 ColStart=2 ColCount=4", viewport)
	}
}

func TestScatterPlotDomainAlias(t *testing.T) {
	domain := ScatterPlotDomain{MinX: 0, MaxX: 10, MinY: 0, MaxY: 12, HasX: true, HasY: true}
	if !domain.HasX || !domain.HasY {
		t.Fatalf("ScatterPlotDomain flags = %+v, want HasX/HasY true", domain)
	}
}

func TestScatterPlotViewportAlias(t *testing.T) {
	viewport := ScatterPlotViewport{MinX: 2, MaxX: 8, MinY: 1, MaxY: 9, HasX: true, HasY: true}
	if !viewport.HasX || !viewport.HasY {
		t.Fatalf("ScatterPlotViewport flags = %+v, want HasX/HasY true", viewport)
	}
}

func TestScatterPlotReferenceBandAlias(t *testing.T) {
	band := ScatterPlotReferenceBand{Min: 2, Max: 4, Label: "Focus"}
	if band.Min != 2 || band.Max != 4 || band.Label != "Focus" {
		t.Fatalf("ScatterPlotReferenceBand = %+v, want Min=2 Max=4 Label=Focus", band)
	}
}

func TestBulletChartAliases(t *testing.T) {
	r := BulletChartQualitativeRange{Limit: 70, Glyph: '▒'}
	if r.Limit != 70 || r.Glyph != '▒' {
		t.Fatalf("BulletChartQualitativeRange = %+v, want Limit=70 Glyph=▒", r)
	}
	if BulletChartValueLabelModeBelow != 2 {
		t.Fatalf("BulletChartValueLabelModeBelow = %d, want 2", BulletChartValueLabelModeBelow)
	}
	if BulletChartDirectionNeutral != 0 {
		t.Fatalf("BulletChartDirectionNeutral = %d, want 0", BulletChartDirectionNeutral)
	}
	if BulletChartDirectionHigherBetter != 1 {
		t.Fatalf("BulletChartDirectionHigherBetter = %d, want 1", BulletChartDirectionHigherBetter)
	}
	if BulletChartDirectionLowerBetter != 2 {
		t.Fatalf("BulletChartDirectionLowerBetter = %d, want 2", BulletChartDirectionLowerBetter)
	}
}

func TestScatterPlotReferenceLineAlias(t *testing.T) {
	line := ScatterPlotReferenceLine{Value: 5, Label: "Target"}
	if line.Value != 5 || line.Label != "Target" {
		t.Fatalf("ScatterPlotReferenceLine = %+v, want Value=5 Label=Target", line)
	}
}
