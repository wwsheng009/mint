package heatmap

import (
	"strings"
	"testing"

	fwtheme "github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/charts/internal/palette"
)

func TestVNodeBuilder(t *testing.T) {
	vnode := NewBuilder([][]float64{{1, 2, 3}, {4, 5, 6}}).
		Title("Usage").
		RowLabels([]string{"API", "Worker"}).
		ColLabels([]string{"M", "T", "W"}).
		Viewport(NewViewport(0, 2, 1, 2)).
		ViewportScale().
		ShowSummary(true).
		MaxRowLabelWidth(4).
		CompactLegend().
		ColorMode(fwtheme.ColorMode16).
		ShowLegend(false).
		Key("heat-1").
		Build()

	heat, ok := vnode.(*VNode)
	if !ok {
		t.Fatal("Build() should return *VNode")
	}
	if heat.Key() != "heat-1" {
		t.Fatalf("Key() = %q, want heat-1", heat.Key())
	}
	if heat.Title() != "Usage" {
		t.Fatalf("Title() = %q, want Usage", heat.Title())
	}
	if got := heat.RowLabels(); len(got) != 2 || got[0] != "API" || got[1] != "Worker" {
		t.Fatalf("RowLabels() = %#v, want [API Worker]", got)
	}
	if got := heat.ColLabels(); len(got) != 3 || got[2] != "W" {
		t.Fatalf("ColLabels() = %#v, want [... W]", got)
	}
	if heat.ShowLegend() {
		t.Fatal("ShowLegend() = true, want false")
	}
	if !heat.ShowSummary() {
		t.Fatal("ShowSummary() = false, want true")
	}
	if heat.SummaryMode() != SummaryModeDetailed {
		t.Fatalf("SummaryMode() = %v, want %v", heat.SummaryMode(), SummaryModeDetailed)
	}
	if heat.LegendMode() != LegendModeCompact {
		t.Fatalf("LegendMode() = %v, want %v", heat.LegendMode(), LegendModeCompact)
	}
	if heat.ColorMode() != fwtheme.ColorMode16 {
		t.Fatalf("ColorMode() = %v, want %v", heat.ColorMode(), fwtheme.ColorMode16)
	}
	if heat.ScaleMode() != ScaleModeViewport {
		t.Fatalf("ScaleMode() = %v, want %v", heat.ScaleMode(), ScaleModeViewport)
	}
	if got := heat.Viewport(); got != NewViewport(0, 2, 1, 2) {
		t.Fatalf("Viewport() = %+v, want %+v", got, NewViewport(0, 2, 1, 2))
	}
	if heat.MaxRowLabelWidth() != 4 {
		t.Fatalf("MaxRowLabelWidth() = %d, want 4", heat.MaxRowLabelWidth())
	}
}

func TestInstanceMeasureAndPaint(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:      "Usage",
		propRowLabels:  []string{"API", "Worker"},
		propColLabels:  []string{"M", "T", "W"},
		propValues:     [][]float64{{1, 2, 3}, {4, 5, 6}},
		propShowAxis:   true,
		propShowLegend: true,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	if size.Height != 5 {
		t.Fatalf("Measure().Height = %d, want 5", size.Height)
	}

	buf := drawCmdsToBuffer(size.Width, size.Height, inst.Paint(0, 0))
	if got := strings.TrimRight(bufferRowText(buf, 0), " "); got != "Usage" {
		t.Fatalf("title row = %q, want Usage", got)
	}
	if got := strings.TrimRight(bufferRowText(buf, 1), " "); got != "Low ░ ▒ ▓ █ High" {
		t.Fatalf("legend row = %q, want %q", got, "Low ░ ▒ ▓ █ High")
	}
	if got := strings.TrimRight(bufferRowText(buf, 2), " "); got != "       M T W" && got != "      M T W" {
		t.Fatalf("column row = %q, want padded M T W", got)
	}
	if got := strings.TrimRight(bufferRowText(buf, 3), " "); !strings.Contains(got, "API") {
		t.Fatalf("row 1 = %q, want row label API", got)
	}
	if got := strings.TrimRight(bufferRowText(buf, 4), " "); !strings.Contains(got, "Worker") {
		t.Fatalf("row 2 = %q, want row label Worker", got)
	}

	render := strings.Join([]string{
		bufferRowText(buf, 3),
		bufferRowText(buf, 4),
	}, "\n")
	for _, glyph := range []rune{'░', '▒', '▓', '█'} {
		if strings.ContainsRune(render, glyph) {
			return
		}
	}
	t.Fatal("plot rows do not contain heatmap glyphs")
}

func TestInstanceSummaryRow(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:       "Summary",
		propValues:      [][]float64{{1, 2, 3}, {4, 5, 6}},
		propShowAxis:    false,
		propShowLegend:  false,
		propShowSummary: true,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	buf := drawCmdsToBuffer(size.Width, size.Height, inst.Paint(0, 0))
	if got := strings.TrimRight(bufferRowText(buf, 1), " "); got != "range 1..6 avg 3.5" {
		t.Fatalf("summary row = %q, want %q", got, "range 1..6 avg 3.5")
	}
}

func TestInstanceCompactSummaryRow(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:       "Summary",
		propValues:      [][]float64{{1, 2, 3}, {4, 5, 6}},
		propShowAxis:    false,
		propShowLegend:  false,
		propSummaryMode: SummaryModeCompact,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	buf := drawCmdsToBuffer(size.Width, size.Height, inst.Paint(0, 0))
	if got := strings.TrimRight(bufferRowText(buf, 1), " "); got != "1..6 avg 3.5" {
		t.Fatalf("compact summary row = %q, want %q", got, "1..6 avg 3.5")
	}
}

func TestInstanceNoDataPaint(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:      "Empty",
		propValues:     [][]float64{},
		propShowLegend: false,
		propShowAxis:   false,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	buf := drawCmdsToBuffer(size.Width, size.Height, inst.Paint(0, 0))
	if got := strings.TrimRight(bufferRowText(buf, 1), " "); got != "No data" {
		t.Fatalf("empty row = %q, want No data", got)
	}
}

func TestInstanceColorModeFallbacks(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propValues:     [][]float64{{1, 5}, {7, 9}},
		propShowAxis:   false,
		propShowLegend: false,
		propColorMode:  fwtheme.ColorMode16,
	})

	cmds := inst.Paint(0, 0)
	hasNamedColor := false
	for _, cmd := range cmds {
		if cmd.Style.FG == "bright-black" || cmd.Style.FG == "bright-cyan" || cmd.Style.FG == "bright-yellow" || cmd.Style.FG == "bright-red" {
			hasNamedColor = true
			break
		}
	}
	if !hasNamedColor {
		t.Fatal("expected named fallback colors in 16-color mode")
	}

	inst = NewInstance(rtui.Props{
		propValues:     [][]float64{{1, 5}, {7, 9}},
		propShowAxis:   false,
		propShowLegend: false,
		propColorMode:  fwtheme.ColorModeNone,
	})
	cmds = inst.Paint(0, 0)
	for _, cmd := range cmds {
		if cmd.Style.FG != "" {
			t.Fatalf("expected no foreground colors in no-color mode, got %q", cmd.Style.FG)
		}
	}
}

func TestInstanceLegendGlyphStyles(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propValues:     [][]float64{{1, 5}, {7, 9}},
		propShowAxis:   false,
		propShowLegend: true,
		propColorMode:  fwtheme.ColorMode16,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	buf := drawCmdsToBuffer(size.Width, size.Height, inst.Paint(0, 0))

	legendGlyphXs := []int{4, 6, 8, 10}
	legendRatios := []float64{0.00, 0.33, 0.66, 1.00}
	for i, x := range legendGlyphXs {
		cell := buf.Cells[0][x]
		if cell.Cluster == "" || cell.Cluster == " " {
			t.Fatalf("legend glyph %d missing at x=%d", i, x)
		}
		want := palette.HeatmapColor(legendRatios[i], fwtheme.ColorMode16)
		if cell.Style.FG != want {
			t.Fatalf("legend glyph %d fg = %q, want %q", i, cell.Style.FG, want)
		}
	}

	inst = NewInstance(rtui.Props{
		propValues:     [][]float64{{1, 5}, {7, 9}},
		propShowAxis:   false,
		propShowLegend: true,
		propColorMode:  fwtheme.ColorModeNone,
	})
	size = inst.Measure(layout.UnboundedConstraints())
	buf = drawCmdsToBuffer(size.Width, size.Height, inst.Paint(0, 0))
	for _, x := range legendGlyphXs {
		if got := buf.Cells[0][x].Style.FG; got != "" {
			t.Fatalf("legend glyph fg in none mode = %q, want empty", got)
		}
	}
}

func TestInstanceCompactLegend(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propValues:     [][]float64{{1, 2}, {3, 4}},
		propShowAxis:   false,
		propShowLegend: true,
		propLegendMode: LegendModeCompact,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	buf := drawCmdsToBuffer(size.Width, size.Height, inst.Paint(0, 0))
	if got := strings.TrimRight(bufferRowText(buf, 0), " "); got != "L ░▒▓█ H" {
		t.Fatalf("compact legend row = %q, want %q", got, "L ░▒▓█ H")
	}
}

func TestInstanceViewportAndLabelClipping(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:            "Viewport",
		propRowLabels:        []string{"North America", "South America", "Europe"},
		propColLabels:        []string{"Mon", "Tue", "Wed", "Thu"},
		propValues:           [][]float64{{1, 2, 3, 4}, {5, 6, 7, 8}, {2, 4, 6, 8}},
		propShowAxis:         true,
		propShowLegend:       false,
		propViewport:         NewViewport(1, 2, 1, 2),
		propMaxRowLabelWidth: 4,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	buf := drawCmdsToBuffer(size.Width, size.Height, inst.Paint(0, 0))
	if got := strings.TrimRight(bufferRowText(buf, 1), " "); got != "     T W" && got != "    T W" {
		t.Fatalf("column row = %q, want visible window T W", got)
	}
	row1 := strings.TrimRight(bufferRowText(buf, 2), " ")
	row2 := strings.TrimRight(bufferRowText(buf, 3), " ")
	if !strings.Contains(row1, "Sou~") {
		t.Fatalf("row 1 = %q, want clipped label Sou~", row1)
	}
	if !strings.Contains(row2, "Euro") && !strings.Contains(row2, "Eur~") {
		t.Fatalf("row 2 = %q, want clipped label for Europe", row2)
	}
	if strings.Contains(row1, "North") || strings.Contains(row2, "Thu") {
		t.Fatalf("viewport leaked hidden labels: row1=%q row2=%q", row1, row2)
	}
}

func TestInstanceViewportScaleUsesVisibleRange(t *testing.T) {
	global := NewInstance(rtui.Props{
		propValues:     [][]float64{{0, 1000, 1000}, {41, 42, 43}, {42, 43, 44}},
		propShowAxis:   false,
		propShowLegend: false,
		propViewport:   NewViewport(1, 2, 0, 3),
		propScaleMode:  ScaleModeGlobal,
	})

	size := global.Measure(layout.UnboundedConstraints())
	buf := drawCmdsToBuffer(size.Width, size.Height, global.Paint(0, 0))
	globalRows := strings.Join([]string{
		strings.TrimRight(bufferRowText(buf, 0), " "),
		strings.TrimRight(bufferRowText(buf, 1), " "),
	}, "\n")
	if strings.ContainsAny(globalRows, "▒▓█") {
		t.Fatalf("global scale rows = %q, want only low-density glyphs in viewport", globalRows)
	}

	local := NewInstance(rtui.Props{
		propValues:     [][]float64{{0, 1000, 1000}, {41, 42, 43}, {42, 43, 44}},
		propShowAxis:   false,
		propShowLegend: false,
		propViewport:   NewViewport(1, 2, 0, 3),
		propScaleMode:  ScaleModeViewport,
	})

	size = local.Measure(layout.UnboundedConstraints())
	buf = drawCmdsToBuffer(size.Width, size.Height, local.Paint(0, 0))
	localRows := strings.Join([]string{
		strings.TrimRight(bufferRowText(buf, 0), " "),
		strings.TrimRight(bufferRowText(buf, 1), " "),
	}, "\n")
	if !strings.ContainsAny(localRows, "▒▓█") {
		t.Fatalf("viewport scale rows = %q, want expanded visible contrast", localRows)
	}
}

func TestInstanceAutoScaleUsesViewportOnlyForPartialWindow(t *testing.T) {
	autoPartial := NewInstance(rtui.Props{
		propValues:     [][]float64{{0, 1000, 1000}, {41, 42, 43}, {42, 43, 44}},
		propShowAxis:   false,
		propShowLegend: false,
		propViewport:   NewViewport(1, 2, 0, 3),
		propScaleMode:  ScaleModeAuto,
	})

	size := autoPartial.Measure(layout.UnboundedConstraints())
	buf := drawCmdsToBuffer(size.Width, size.Height, autoPartial.Paint(0, 0))
	autoPartialRows := strings.Join([]string{
		strings.TrimRight(bufferRowText(buf, 0), " "),
		strings.TrimRight(bufferRowText(buf, 1), " "),
	}, "\n")
	if !strings.ContainsAny(autoPartialRows, "▒▓█") {
		t.Fatalf("auto scale with partial viewport = %q, want expanded visible contrast", autoPartialRows)
	}

	autoFull := NewInstance(rtui.Props{
		propValues:     [][]float64{{0, 1000, 1000}, {41, 42, 43}, {42, 43, 44}},
		propShowAxis:   false,
		propShowLegend: false,
		propScaleMode:  ScaleModeAuto,
	})

	size = autoFull.Measure(layout.UnboundedConstraints())
	buf = drawCmdsToBuffer(size.Width, size.Height, autoFull.Paint(0, 0))
	autoFullRows := strings.Join([]string{
		strings.TrimRight(bufferRowText(buf, 0), " "),
		strings.TrimRight(bufferRowText(buf, 1), " "),
		strings.TrimRight(bufferRowText(buf, 2), " "),
	}, "\n")
	if !strings.ContainsRune(autoFullRows, '█') {
		t.Fatalf("auto scale without viewport = %q, want global full-matrix scaling", autoFullRows)
	}
}

func TestInstanceAutoScaleKeepsNearFullWindowGlobal(t *testing.T) {
	autoNearFull := NewInstance(rtui.Props{
		propValues: [][]float64{
			{40, 41, 42, 1000},
			{41, 42, 43, 1000},
			{42, 43, 44, 1000},
			{43, 44, 45, 1000},
		},
		propShowAxis:   false,
		propShowLegend: false,
		propViewport:   NewViewport(0, 4, 0, 3),
		propScaleMode:  ScaleModeAuto,
	})

	size := autoNearFull.Measure(layout.UnboundedConstraints())
	buf := drawCmdsToBuffer(size.Width, size.Height, autoNearFull.Paint(0, 0))
	rows := strings.Join([]string{
		strings.TrimRight(bufferRowText(buf, 0), " "),
		strings.TrimRight(bufferRowText(buf, 1), " "),
		strings.TrimRight(bufferRowText(buf, 2), " "),
		strings.TrimRight(bufferRowText(buf, 3), " "),
	}, "\n")
	if strings.ContainsAny(rows, "▒▓█") {
		t.Fatalf("auto scale near-full viewport = %q, want global scaling to stay active", rows)
	}
}

func drawCmdsToBuffer(width, height int, cmds []paint.DrawCmd) *paint.Buffer {
	buf := paint.NewBuffer(width, height)
	for _, cmd := range cmds {
		buf.SetString(cmd.X, cmd.Y, cmd.Text, cmd.Style)
	}
	return buf
}

func bufferRowText(buf *paint.Buffer, y int) string {
	if y < 0 || y >= buf.Height {
		return ""
	}
	var builder strings.Builder
	for x := 0; x < buf.Width; x++ {
		cell := buf.Cells[y][x]
		if cell.IsContinuation {
			continue
		}
		cluster := cell.Cluster
		if cluster == "" {
			cluster = " "
		}
		builder.WriteString(cluster)
	}
	return builder.String()
}
