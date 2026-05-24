package barchart

import (
	"strings"
	"testing"

	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/charts/internal/palette"
)

func TestVNodeBuilder(t *testing.T) {
	vnode := NewBuilder([]float64{3, 5, 2}).
		Title("Bars").
		Labels([]string{"A", "B", "C"}).
		Series(
			Series{Name: "Revenue", Values: []float64{3, 5, 2}},
			Series{Name: "Cost", Values: []float64{2, 4, 1}},
		).
		Stacked().
		Horizontal().
		Width(8).
		Height(4).
		ShowLegend(true).
		ShowValue(true).
		ASCII().
		Key("bar-1").
		Build()

	bar, ok := vnode.(*VNode)
	if !ok {
		t.Fatal("Build() should return *VNode")
	}
	if bar.Key() != "bar-1" {
		t.Fatalf("Key() = %q, want bar-1", bar.Key())
	}
	if bar.Title() != "Bars" {
		t.Fatalf("Title() = %q, want Bars", bar.Title())
	}
	if bar.Height() != 4 {
		t.Fatalf("Height() = %d, want 4", bar.Height())
	}
	if len(bar.Series()) != 2 {
		t.Fatalf("Series() len = %d, want 2", len(bar.Series()))
	}
	if bar.Mode() != ModeStacked {
		t.Fatalf("Mode() = %v, want ModeStacked", bar.Mode())
	}
	if bar.Orientation() != OrientationHorizontal {
		t.Fatalf("Orientation() = %v, want OrientationHorizontal", bar.Orientation())
	}
	if !bar.ShowLegend() {
		t.Fatal("ShowLegend() = false, want true")
	}
	if !bar.ShowValue() {
		t.Fatal("ShowValue() = false, want true")
	}
	if bar.RenderMode() != RenderModeASCII {
		t.Fatalf("RenderMode() = %v, want RenderModeASCII", bar.RenderMode())
	}
}

func TestInstanceMeasureAndPaint(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:    "Bars",
		propLabels:   []string{"A", "B", "C"},
		propValues:   []float64{3, 5, 2},
		propWidth:    5,
		propHeight:   4,
		propShowAxis: true,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	if size.Width < 5 {
		t.Fatalf("Measure().Width = %d, want >= 5", size.Width)
	}
	if size.Height != 7 {
		t.Fatalf("Measure().Height = %d, want 7", size.Height)
	}

	buf := drawCmdsToBuffer(size.Width, size.Height, inst.Paint(0, 0))
	if got := strings.TrimRight(bufferRowText(buf, 0), " "); got != "Bars" {
		t.Fatalf("title row = %q, want Bars", got)
	}
	if got := strings.TrimRight(bufferRowText(buf, 5), " "); got != "─────" {
		t.Fatalf("axis row = %q, want %q", got, "─────")
	}
	if got := strings.TrimRight(bufferRowText(buf, 6), " "); got != "A B C" {
		t.Fatalf("label row = %q, want %q", got, "A B C")
	}
}

func TestInstanceWidthSampling(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propLabels: []string{"A", "B", "C", "D", "E"},
		propValues: []float64{1, 2, 3, 4, 5},
		propWidth:  3,
		propHeight: 3,
	})

	seriesList := inst.visibleSeries()
	if len(seriesList) != 1 {
		t.Fatalf("visible series len = %d, want 1", len(seriesList))
	}
	if len(seriesList[0].Values) != 2 {
		t.Fatalf("visible bars = %d, want 2", len(seriesList[0].Values))
	}
	labels := inst.visibleLabels(2)
	if labels[0] != "A" || labels[1] != "E" {
		t.Fatalf("labels = %#v, want [A E]", labels)
	}
}

func TestVerticalLabelRowFoldsDenseLabels(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propLabels: []string{"North America", "Latin America", "Asia Pacific"},
		propValues: []float64{4, 3, 5},
		propWidth:  5,
		propHeight: 4,
	})

	row := inst.verticalLabelRow(5, inst.visibleLabels(3), inst.groupCenters(3))
	if got := strings.TrimRight(row, " "); got != "N L A" {
		t.Fatalf("verticalLabelRow() = %q, want %q", got, "N L A")
	}
}

func TestInstanceMultiSeriesPaint(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:      "Bars",
		propLabels:     []string{"A", "B", "C"},
		propSeries:     []Series{{Name: "Revenue", Values: []float64{3, 5, 2}}, {Name: "Cost", Values: []float64{2, 4, 1}}},
		propMode:       ModeGrouped,
		propWidth:      8,
		propHeight:     4,
		propShowAxis:   true,
		propShowLegend: true,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	buf := drawCmdsToBuffer(size.Width, size.Height, inst.Paint(0, 0))
	if got := strings.TrimRight(bufferRowText(buf, 1), " "); got != "█ Revenue" {
		t.Fatalf("legend row 1 = %q, want %q", got, "█ Revenue")
	}
	if got := strings.TrimRight(bufferRowText(buf, 2), " "); got != "█ Cost" {
		t.Fatalf("legend row 2 = %q, want %q", got, "█ Cost")
	}

	hasSeries0 := false
	hasSeries1 := false
	for y := 3; y < 3+inst.plotHeight(); y++ {
		for x := 0; x < buf.Width; x++ {
			cell := buf.Cells[y][x]
			if cell.IsContinuation || cell.Cluster == "" || cell.Cluster == " " {
				continue
			}
			if cell.Style.FG == palette.SeriesColor(0) {
				hasSeries0 = true
			}
			if cell.Style.FG == palette.SeriesColor(1) {
				hasSeries1 = true
			}
		}
	}
	if !hasSeries0 || !hasSeries1 {
		t.Fatalf("multi-series colors missing: series0=%v series1=%v", hasSeries0, hasSeries1)
	}
}

func TestInstanceStackedPaint(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:      "Stacked Bars",
		propLabels:     []string{"X", "Y", "Z"},
		propSeries:     []Series{{Name: "North", Values: []float64{3, 2, 1}}, {Name: "South", Values: []float64{1, 3, 2}}},
		propMode:       ModeStacked,
		propWidth:      5,
		propHeight:     4,
		propShowAxis:   true,
		propShowLegend: true,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	buf := drawCmdsToBuffer(size.Width, size.Height, inst.Paint(0, 0))
	if got := strings.TrimRight(bufferRowText(buf, 1), " "); got != "█ North" {
		t.Fatalf("legend row 1 = %q, want %q", got, "█ North")
	}
	if got := strings.TrimRight(bufferRowText(buf, 2), " "); got != "█ South" {
		t.Fatalf("legend row 2 = %q, want %q", got, "█ South")
	}
	if got := strings.TrimRight(bufferRowText(buf, size.Height-1), " "); got != "X Y Z" {
		t.Fatalf("label row = %q, want %q", got, "X Y Z")
	}

	columnX := 0
	topHasAccent := false
	bottomHasPrimary := false
	for y := 3; y < 3+inst.plotHeight(); y++ {
		cell := buf.Cells[y][columnX]
		if cell.Cluster != "█" {
			continue
		}
		if cell.Style.FG == palette.SeriesColor(0) {
			bottomHasPrimary = true
		}
		if cell.Style.FG == palette.SeriesColor(1) {
			topHasAccent = true
		}
	}
	if !topHasAccent || !bottomHasPrimary {
		t.Fatalf("stacked colors missing: primary=%v accent=%v", bottomHasPrimary, topHasAccent)
	}
}

func TestInstanceVerticalValueRows(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:     "Value Rows",
		propLabels:    []string{"A", "B", "C"},
		propValues:    []float64{12, 7, 15},
		propWidth:     5,
		propHeight:    4,
		propShowAxis:  true,
		propShowValue: true,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	buf := drawCmdsToBuffer(size.Width, size.Height, inst.Paint(0, 0))
	if got := strings.TrimRight(bufferRowText(buf, size.Height-1), " "); got != "Values: 12 7 15" {
		t.Fatalf("value row = %q, want %q", got, "Values: 12 7 15")
	}
}

func TestInstanceStackedValueRows(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:      "Stacked Values",
		propLabels:     []string{"X", "Y", "Z"},
		propSeries:     []Series{{Name: "North", Values: []float64{3, 2, 1}}, {Name: "South", Values: []float64{1, 3, 2}}},
		propMode:       ModeStacked,
		propWidth:      5,
		propHeight:     4,
		propShowAxis:   true,
		propShowLegend: true,
		propShowValue:  true,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	buf := drawCmdsToBuffer(size.Width, size.Height, inst.Paint(0, 0))
	if got := strings.TrimRight(bufferRowText(buf, size.Height-3), " "); got != "North: 3 2 1" {
		t.Fatalf("series value row 1 = %q, want %q", got, "North: 3 2 1")
	}
	if got := strings.TrimRight(bufferRowText(buf, size.Height-2), " "); got != "South: 1 3 2" {
		t.Fatalf("series value row 2 = %q, want %q", got, "South: 1 3 2")
	}
	if got := strings.TrimRight(bufferRowText(buf, size.Height-1), " "); got != "Total: 4 5 3" {
		t.Fatalf("total value row = %q, want %q", got, "Total: 4 5 3")
	}
}

func TestCompactLabel(t *testing.T) {
	tests := []struct {
		label string
		width int
		want  string
	}{
		{label: "North America", width: 2, want: "NA"},
		{label: "Region 24", width: 1, want: "4"},
		{label: "Alpha", width: 2, want: "Aa"},
	}

	for _, tt := range tests {
		if got := compactLabel(tt.label, tt.width); got != tt.want {
			t.Fatalf("compactLabel(%q, %d) = %q, want %q", tt.label, tt.width, got, tt.want)
		}
	}
}

func TestInstanceHorizontalPaint(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:       "Horizontal Bars",
		propLabels:      []string{"Alpha", "Beta", "Gamma"},
		propSeries:      []Series{{Name: "Revenue", Values: []float64{3, 5, 2}}, {Name: "Cost", Values: []float64{2, 4, 1}}},
		propMode:        ModeGrouped,
		propOrientation: OrientationHorizontal,
		propWidth:       14,
		propHeight:      5,
		propShowAxis:    true,
		propShowLegend:  true,
		propShowValue:   true,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	buf := drawCmdsToBuffer(size.Width, size.Height, inst.Paint(0, 0))

	if got := strings.TrimRight(bufferRowText(buf, 0), " "); got != "Horizontal Bars" {
		t.Fatalf("title row = %q, want %q", got, "Horizontal Bars")
	}
	if got := strings.TrimRight(bufferRowText(buf, 1), " "); got != "█ Revenue" {
		t.Fatalf("legend row 1 = %q, want %q", got, "█ Revenue")
	}
	if got := strings.TrimRight(bufferRowText(buf, 2), " "); got != "█ Cost" {
		t.Fatalf("legend row 2 = %q, want %q", got, "█ Cost")
	}
	if got := strings.TrimRight(bufferRowText(buf, 3), " "); !strings.Contains(got, "Alpha") {
		t.Fatalf("first plot row = %q, want label Alpha", got)
	}
	if got := strings.TrimRight(bufferRowText(buf, 3), " "); !strings.Contains(got, "3") {
		t.Fatalf("first plot row = %q, want inline value 3", got)
	}
	if got := strings.TrimRight(bufferRowText(buf, size.Height-1), " "); !strings.Contains(got, "────") {
		t.Fatalf("axis row = %q, want horizontal axis", got)
	}

	hasPrimary := false
	hasAccent := false
	for y := 3; y < size.Height-1; y++ {
		for x := 0; x < buf.Width; x++ {
			cell := buf.Cells[y][x]
			if cell.IsContinuation || cell.Cluster != "█" {
				continue
			}
			if cell.Style.FG == palette.SeriesColor(0) {
				hasPrimary = true
			}
			if cell.Style.FG == palette.SeriesColor(1) {
				hasAccent = true
			}
		}
	}
	if !hasPrimary || !hasAccent {
		t.Fatalf("horizontal colors missing: primary=%v accent=%v", hasPrimary, hasAccent)
	}
}

func TestInstanceASCIIRenderModePaint(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:      "ASCII Bars",
		propLabels:     []string{"A", "B", "C"},
		propSeries:     []Series{{Name: "Revenue", Values: []float64{3, 5, 2}}, {Name: "Cost", Values: []float64{2, 4, 1}}},
		propMode:       ModeGrouped,
		propWidth:      8,
		propHeight:     4,
		propShowAxis:   true,
		propShowLegend: true,
		propRenderMode: RenderModeASCII,
	})

	size := inst.Measure(layout.UnboundedConstraints())
	buf := drawCmdsToBuffer(size.Width, size.Height, inst.Paint(0, 0))
	rendered := bufferText(buf)

	if !strings.Contains(rendered, "# Revenue") {
		t.Fatalf("ASCII render =\n%s\nwant legend marker '# Revenue'", rendered)
	}
	if !strings.Contains(rendered, "--------") {
		t.Fatalf("ASCII render =\n%s\nwant ASCII axis", rendered)
	}
	if strings.Contains(rendered, "█") || strings.Contains(rendered, "─") {
		t.Fatalf("ASCII render contains non-ASCII chart glyphs:\n%s", rendered)
	}
}

func TestInstanceHorizontalLabelFolding(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTitle:       "Horizontal Fold",
		propLabels:      []string{"North America", "Latin America"},
		propValues:      []float64{10, 8},
		propOrientation: OrientationHorizontal,
		propWidth:       10,
		propHeight:      3,
		propShowAxis:    true,
	})

	if got := inst.horizontalLabelWidth(); got != 5 {
		t.Fatalf("horizontalLabelWidth() = %d, want 5", got)
	}

	size := inst.Measure(layout.UnboundedConstraints())
	buf := drawCmdsToBuffer(size.Width, size.Height, inst.Paint(0, 0))
	if got := strings.TrimRight(bufferRowText(buf, 1), " "); !strings.Contains(got, "NA") {
		t.Fatalf("first plot row = %q, want folded label containing NA", got)
	}
	if got := strings.TrimRight(bufferRowText(buf, 3), " "); !strings.Contains(got, "LA") {
		t.Fatalf("second category row = %q, want folded label containing LA", got)
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

func bufferText(buf *paint.Buffer) string {
	var builder strings.Builder
	for y := 0; y < buf.Height; y++ {
		if y > 0 {
			builder.WriteByte('\n')
		}
		builder.WriteString(bufferRowText(buf, y))
	}
	return builder.String()
}
